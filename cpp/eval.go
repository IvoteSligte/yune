package cpp

import (
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
	c := repl{stdin, stdout, ""}
	err = c.Declare(pbHeader)
	if err != nil {
		log.Fatalln("Failed to declare PB header through clang-repl. Error:", err)
	}
	return c
}()

// clang-repl wrapper struct
type repl struct {
	stdin    io.Writer
	stdout   io.Reader
	declared string
}

func (r *repl) Evaluate(expr Expression) (output string, err error) {
	text := "std::cout << ty::serialize(" + strings.ReplaceAll(expr, "\n", "") + ") << std::endl;"
	println("Evaluating:", text)
	_, err = r.stdin.Write([]byte(text))
	if err != nil {
		return
	}
	bytes := []byte{}
	_, err = r.stdout.Read(bytes)
	output = string(bytes)
	if output == "" {
		println(r.declared + text)
		log.Panicf("clang-repl evaluated '%s' to the empty string.\n", expr)
	} else {
		log.Printf("clang-repl evaluated '%s' to '%s'\n", expr, output)
	}
	return
}

// Write text without expecting a response, such as for function or constant declarations.
func (r *repl) Declare(text string) (err error) {
	_, err = r.stdin.Write([]byte(text))
	r.declared += "\n" + text
	return
}

func (r *repl) GetDeclared() string {
	return r.declared
}
