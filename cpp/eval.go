package cpp

import (
	"bufio"
	_ "embed"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

//go:embed "pb.hpp"
var pbHeader string

var Repl repl = func() repl {
	cmd := exec.Command("clang-repl")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalln("Failed to get stdin pipe from clang-repl command. Error:", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalln("Failed to get stdout pipe from clang-repl command. Error:", err)
	}
	cmd.Stderr = os.Stderr // TODO: also wrap this because when stderr is written to something went wrong
	err = cmd.Start()
	if err != nil {
		log.Fatalln("Failed to run clang-repl. Error:", err)
	}
	r := repl{
		stdin:    stdin,
		stdout:   bufio.NewReader(stdout),
		declared: "",
	}
	err = r.Declare(os.ExpandEnv(`#include "$PWD/cpp/pb.hpp"`))
	if err != nil {
		log.Fatalln("Failed to declare PB header through clang-repl. Error:", err)
	}
	return r
}()

// clang-repl wrapper struct
type repl struct {
	stdin    io.Writer
	stdout   *bufio.Reader
	declared string
}

func sanitize(s string) string {
	return strings.ReplaceAll(s, "\n", "")
}

func (r *repl) Evaluate(expr Expression) (output string, err error) {
	text := "std::cout << ty::serialize(" + sanitize(expr) + ") << std::endl;"
	_, err = r.stdin.Write([]byte(text + "\n"))
	if err != nil {
		return
	}
	output, err = r.stdout.ReadString('\n')
	if err != nil {
		return
	}
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

// Write text without expecting a response, such as for function or constant declarations.
func (r *repl) Declare(text string) (err error) {
	_, err = r.stdin.Write([]byte(sanitize(text) + "\n"))
	r.declared += text + "\n"
	return
}

func (r *repl) GetDeclared() string {
	return r.declared
}
