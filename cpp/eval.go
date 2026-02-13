package cpp

import (
	_ "embed"
	"io"
	"log"
	"os/exec"
)

//go:embed "pb.hpp"
var pbHeader string

var Cling cling = func() cling {
	cmd := exec.Command("cling")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalln("Failed to get stdin pipe from Cling command. Error:", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalln("Failed to get stdout pipe from Cling command. Error:", err)
	}
	err = cmd.Start()
	if err != nil {
		log.Fatalln("Failed to run Cling. Error:", err)
	}
	c := cling{stdin, stdout, ""}
	err = c.Declare(pbHeader)
	if err != nil {
		log.Fatalln("Failed to declare PB header in Cling. Error:", err)
	}
	return c
}()

type cling struct {
	stdin    io.Writer
	stdout   io.Reader
	declared string
}

func (c *cling) Evaluate(expr Expression) (output string, err error) {
	_, err = c.stdin.Write([]byte("std::cout << ty::serialize(" + expr + ") << std::endl;"))
	if err != nil {
		return
	}
	bytes := []byte{}
	_, err = c.stdout.Read(bytes)
	output = string(bytes)
	log.Printf("Cling evaluated '%s' to '%s'\n", expr, output)
	return
}

// Write text without expecting a response, such as for function or constant declarations.
func (c *cling) Declare(text string) (err error) {
	_, err = c.stdin.Write([]byte(text))
	c.declared += "\n" + text
	return
}

func (c *cling) GetDeclared() string {
	return c.declared
}
