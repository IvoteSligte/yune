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

// TODO: properly close file
var evalLogFile = func() *os.File {
	filename := "/tmp/yune-eval.log"
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
	err = r.Write(os.ExpandEnv(`#include "$PWD/cpp/pb.hpp"`))
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
	evalLog("`" + expr + "`\n")
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
func (r *repl) Write(text string) (err error) {
	evalLog(text + "\n")
	_, err = r.stdin.Write([]byte(sanitize(text) + "\n"))
	r.declared += text + "\n"
	return
}

func (r *repl) GetDeclared() string {
	return r.declared
}
