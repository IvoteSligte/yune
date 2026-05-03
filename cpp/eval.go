package cpp

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"

	fj "github.com/valyala/fastjson"
)

func newLogFile() *os.File {
	file, err := os.CreateTemp("", "yune-log-*.cpp")
	if err != nil {
		log.Println("Failed to open evaluation log file for writing. Error:", err)
		return nil
	}
	log.Printf("Created interpreter log file '%s'.\n", file.Name())
	return file
}

type ProxyStderr struct{}

// Write implements io.Writer.
func (ProxyStderr) Write(p []byte) (n int, err error) {
	n, err = os.Stderr.Write(p)
	if err != nil {
		return
	}
	if strings.Contains(string(p), "error: Parsing failed.") {
		err = fmt.Errorf("clang-repl returned an error, stopping evaluation.")
		// NOTE: using this for now since errors returned from this function are simply ignored
		log.Fatalf("%s", err)
	}
	return
}

var _ io.Writer = ProxyStderr{}

// should be unused according to https://en.wikipedia.org/wiki/List_of_TCP_and_UDP_port_numbers
// (synchronised with ipc.hpp)
const YuneCompilerPort = 11555

func NewInterpreter() *Interpreter {
	// Create connection
	// TODO: close connection at some point
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", YuneCompilerPort))
	if err != nil {
		log.Fatalln("Failed to start TCP connection with clang-repl. Error:", err)
	}
	// Start REPL and setup inputs/outputs
	cmd := exec.Command("clang-repl", "-Xcc=-std=c++23")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalln("Failed to get stdin pipe from clang-repl command. Error:", err)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = ProxyStderr{}
	if err = cmd.Start(); err != nil {
		log.Fatalln("Failed to run clang-repl. Error:", err)
	}
	r := &Interpreter{
		writer:   stdin,
		reader:   nil, // set later
		command:  cmd,
		listener: listener,
		logFile:  newLogFile(),
		Declared: "",
	}
	compileTimeIncludes := os.ExpandEnv(`#include "$PWD/cpp/pb.hpp"
#include "$PWD/cpp/ipc.hpp"
#include <thread>
`)
	if err = r.Declare(compileTimeIncludes); err != nil {
		log.Fatalln("Failed to declare compile-time #includes through clang-repl. Error:", err)
	}
	// have ipc.hpp connect
	conn, err := listener.Accept()
	r.reader = bufio.NewReader(conn)
	return r
}

func sanitize(s string) string {
	s = strings.TrimRight(s, " \n\t")
	s = strings.ReplaceAll(s, "\n", "\\\n")
	return s
}

// clang-Interpreter wrapper struct
type Interpreter struct {
	writer   io.WriteCloser
	reader   *bufio.Reader
	command  *exec.Cmd
	listener net.Listener
	logFile  *os.File
	Declared string
}

func (r *Interpreter) log(message string) {
	_, err := r.logFile.WriteString(message)
	if err != nil {
		log.Println("Failed to write to open evaluation log file. Error:", err)
	}
}

func (r *Interpreter) Close() {
	if err := r.writer.Close(); err != nil {
		log.Println("Failed to close interpreter write file.")
	}
	// NOTE: should this remove the logfile?
	// it is useful to be able to look at logs after compilation
	if err := r.logFile.Close(); err != nil {
		log.Println("Failed to close interpreter log file. Error:", err)
	}
	if err := r.listener.Close(); err != nil {
		log.Println("Failed to close connection to C++. Error:", err)
	}
	log.Println("Waiting for clang-repl to exit...")
	if err := r.command.Wait(); err != nil {
		log.Println("Failed to wait for clang-repl to exit. Error:", err)
	} else {
		log.Println("Clang-repl successfully exited.")
	}
}

type getType = func(string) Type
type registerNamedType = func(string, *fj.Value)
type checkNamedType = func(string)

func (r *Interpreter) Evaluate(expr Expression, getType getType, registerNamedType registerNamedType, checkNamedType checkNamedType) (output *fj.Value, err error) {
	// A thread is created and detached because in order to write the result of a getType query
	// more code needs to be evaluated by the interpreter, which causes a deadlock with only a single thread.
	text := "std::thread([]() { compiler_connection.yield(toJson_(" + sanitize(expr) + ")); }).detach();\n"
	r.log(text)
	_, err = r.writer.Write([]byte(text))
	if err != nil {
		return
	}
	output, err = r.readResult(getType, registerNamedType, checkNamedType)
	log.Printf("clang-repl evaluated '%s' to '%s'\n", expr, output)
	return
}

func (r *Interpreter) readResult(getType getType, registerNamedType registerNamedType, checkNamedType checkNamedType) (result *fj.Value, err error) {
	for {
		var read string
		read, err = r.reader.ReadString('\n')
		if err != nil {
			return
		}
		log.Printf("clang-repl message: %s", read)
		message := fj.MustParse(read)
		if message.Get("finished") != nil {
			err = fmt.Errorf("Compiler connection reported 'finished' while waiting for a result. Message: '%s'.", message)
			return
		}
		if result = message.Get("result"); result != nil {
			return
		}
		if nameBytes := message.GetStringBytes("getType"); nameBytes != nil {
			name := string(nameBytes)
			// set the type and signal that it has been set
			_type := getType(name)
			if err = r.Write(fmt.Sprintf("compiler_connection.set_type(%s);", _type)); err != nil {
				err = fmt.Errorf("Failed to set type after getType request. Message: '%s'. Error: %s", message, err)
				return
			}
			continue
		}
		if requestJson := message.Get("registerNamedType"); requestJson != nil {
			name := string(requestJson.GetStringBytes("name"))
			value := requestJson.Get("value")
			registerNamedType(name, value)
			continue
		}
		if requestJson := message.Get("checkNamedType"); requestJson != nil {
			name := string(requestJson.GetStringBytes("name"))
			checkNamedType(name)
			continue
		}
		err = fmt.Errorf("Could not parse evaluation message: '%s'", message)
		return
	}
}

// The interpreter by default only waits for results.
// This means that Interpreter.Declare may cause an error to be reported by clang-repl,
// which is then ignored by the interpreter that already closed the connection.
func (r *Interpreter) WaitForFinish() (err error) {
	r.Write(`compiler_connection.send_finished();`)
	read, err := r.reader.ReadString('\n')
	if err != nil {
		return
	}
	message := fj.MustParse(read)
	if message.Get("finished") == nil {
		err = fmt.Errorf("Expected 'finished' message, found: %s", message)
	}
	return
}

// Write text without expecting a response, such as for global declarations.
// This does not write to the `Declared` field, which is useful for compile-time-only declarations.
func (r *Interpreter) Write(text string) (err error) {
	input := sanitize(text) + "\n"
	r.log(input)
	_, err = r.writer.Write([]byte(input))
	return
}

// Write text without expecting a response, such as for global declarations.
// Text written is appended to the `Declared` field.
func (r *Interpreter) Declare(text string) (err error) {
	r.Write(text)
	r.Declared += text + "\n"
	return
}
