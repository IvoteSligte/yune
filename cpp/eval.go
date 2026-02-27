package cpp

import (
	"bufio"
	_ "embed"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
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

func (r *repl) Evaluate(expr Expression) (output string, err error) {
	text := "compiler_connection.yield(ty::serialize(" + sanitize(expr) + "));\n"
	evalLog(text)
	_, err = r.writer.Write([]byte(text))
	if err != nil {
		return
	}
	output, err = r.readResult()
	if output == "" {
		log.Panicf("clang-repl evaluated '%s' to the empty string.\n", expr)
	} else {
		// NOTE: why is there an extra newline at the end of the string?
		if output[len(output)-1] == '\n' {
			output = output[:len(output)-1]
		}
		log.Printf("clang-repl evaluated '%s' to '%s'\n", expr, output)
	}
	return
}

func (r *repl) readResult() (result string, err error) {
	for {
		read, err := r.reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		splits := strings.Split(read, ":")
		prefix := splits[0]
		switch prefix {
		case "getType":
			panic("unimplemented")
			// continue
		case "result":
			return splits[1], nil
		default:
			panic(fmt.Sprintf("Unexpected prefix for evaluation: '%s'", prefix))
		}
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
