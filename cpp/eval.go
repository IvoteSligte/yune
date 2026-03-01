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

// TODO: properly close file
var evalLogFile = func() *os.File {
	filename := "/tmp/yune-eval-log.cpp"
	_ = os.Remove(filename)
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Failed to open evaluation log file for writing. Error:", err)
		return nil
	}
	return file
}()

func evalLog(s string) {
	if evalLogFile != nil {
		_, err := evalLogFile.WriteString(s)
		if err != nil {
			log.Println("Failed to write to open evaluation log file. Error:", err)
		}
	}
}

// should be unused according to https://en.wikipedia.org/wiki/List_of_TCP_and_UDP_port_numbers
// (synchronised with ipc.hpp)
const YuneCompilerPort = 11555

var Repl repl = func() repl {
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
	cmd.Stderr = os.Stderr // TODO: wrap this because when stderr is written to something went wrong
	if err = cmd.Start(); err != nil {
		log.Fatalln("Failed to run clang-repl. Error:", err)
	}
	r := repl{
		writer:    stdin,
		reader:    nil, // set later
		responder: nil, // set later
		Declared:  "",
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
	r.responder = conn
	return r
}()

// clang-repl wrapper struct
type repl struct {
	writer    io.Writer
	responder io.Writer
	reader    *bufio.Reader
	Declared  string
}

func sanitize(s string) string {
	s = strings.TrimRight(s, " \n\t")
	s = strings.ReplaceAll(s, "\n", "\\\n")
	return s
}

func (r *repl) Evaluate(expr Expression, getType func(string) Type) (output *fj.Value, err error) {
	// A thread is created and detached because in order to write the result of a getType query
	// more code needs to be evaluated by the interpreter, which causes a deadlock with only a single thread.
	text := "std::thread([]() { compiler_connection.yield(ty::serialize(" + sanitize(expr) + ")); }).detach();\n"
	evalLog(text)
	_, err = r.writer.Write([]byte(text))
	if err != nil {
		return
	}
	output, err = r.readResult(getType)
	log.Printf("clang-repl evaluated '%s' to '%s'\n", expr, output)
	return
}

func (r *repl) readResult(getType func(string) Type) (result *fj.Value, err error) {
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
			err = r.Write(fmt.Sprintf("compiler_connection.set_type(%s);", getType(name)))
			if err != nil {
				panic(fmt.Sprintf("Failed to set type after getType request. Message: '%s'. Error: %s", message, err))
			}
			continue
		}
		panic(fmt.Sprintf("Could not parse evaluation message: '%s'", message))
	}
}

// Write text without expecting a response, such as for global declarations.
// This does not write to the `Declared` field, which is useful for compile-time-only declarations.
func (r *repl) Write(text string) (err error) {
	input := sanitize(text) + "\n"
	evalLog(input)
	_, err = r.writer.Write([]byte(input))
	return
}

// Write text without expecting a response, such as for global declarations.
// Text written is appended to the `Declared` field.
func (r *repl) Declare(text string) (err error) {
	r.Write(text)
	r.Declared += text + "\n"
	return
}
