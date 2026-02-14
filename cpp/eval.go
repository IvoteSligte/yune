package cpp

import (
	_ "embed"
	"io"
	"log"
	"os"
	"os/exec"
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

func (c *repl) Evaluate(expr Expression) (output string, err error) {
	_, err = c.stdin.Write([]byte("std::cout << ty::serialize(" + expr + ") << std::endl;"))
	if err != nil {
		return
	}
	bytes := []byte{}
	_, err = c.stdout.Read(bytes)
	output = string(bytes)
	log.Printf("clang-repl evaluated '%s' to '%s'\n", expr, output)
	return
}

// Write text without expecting a response, such as for function or constant declarations.
func (c *repl) Declare(text string) (err error) {
	_, err = c.stdin.Write([]byte(text))
	c.declared += "\n" + text
	return
}

func (c *repl) GetDeclared() string {
	return c.declared
}
