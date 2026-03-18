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
	// TODO: random filename?
	filename := "/tmp/yune-eval-log.cpp"
	_ = os.Remove(filename)
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Failed to open evaluation log file for writing. Error:", err)
		return nil
	}
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
	cmd := exec.Command("clang-repl", "-Xcc=-std=c++20")
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
		logFile:  newLogFile(),
		Declared: "",
	}
	if err = r.Declare(os.ExpandEnv(`#include "$PWD/cpp/pb.hpp"`)); err != nil {
		log.Fatalln("Failed to declare PB header through clang-repl. Error:", err)
	}
	if err = r.Write(os.ExpandEnv(`#include "$PWD/cpp/ipc.hpp"`) + "\n"); err != nil {
		log.Fatalln("Failed to declare IPC header through clang-repl. Error:", err)
	}
	if err = r.Write("#include <thread>\n"); err != nil {
		log.Fatalln("Failed to declare Yune evaluator-specific C++ includes. Error:", err)
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
	if err := r.logFile.Close(); err != nil {
		log.Println("Failed to close interpreter log file.")
	}
}

func (r *Interpreter) Evaluate(expr Expression, getType func(string) Type) (output *fj.Value, err error) {
	// A thread is created and detached because in order to write the result of a getType query
	// more code needs to be evaluated by the interpreter, which causes a deadlock with only a single thread.
	text := "std::thread([]() { compiler_connection.yield(ty::serialize(" + sanitize(expr) + ")); }).detach();\n"
	r.log(text)
	_, err = r.writer.Write([]byte(text))
	if err != nil {
		return
	}
	output, err = r.readResult(getType)
	log.Printf("clang-repl evaluated '%s' to '%s'\n", expr, output)
	return
}

func (r *Interpreter) readResult(getType func(string) Type) (result *fj.Value, err error) {
	for {
		var read string
		read, err = r.reader.ReadString('\n')
		if err != nil {
			return
		}
		message := fj.MustParse(read)
		if result = message.Get("result"); result != nil {
			return
		}
		if nameBytes := message.GetStringBytes("getType"); nameBytes != nil {
			name := string(nameBytes)
			// set the type and signal that it has been set
			if err = r.Write(fmt.Sprintf("compiler_connection.set_type(%s);", getType(name))); err != nil {
				err = fmt.Errorf("Failed to set type after getType request. Message: '%s'. Error: %s", message, err)
				return
			}
			continue
		}
		err = fmt.Errorf("Could not parse evaluation message: '%s'", message)
		return
	}
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
