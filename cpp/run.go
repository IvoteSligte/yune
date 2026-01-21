package cpp

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

func format(code string) (formatted string, err error) {
	cmd := exec.Command("clang-format")
	if cmd.Err != nil {
		err = cmd.Err
		return
	}
	cmd.Stdin = strings.NewReader(code)
	outputBytes, err := cmd.Output()
	if err != nil {
		return
	}
	formatted = string(outputBytes)
	return
}

func PrintFormatted(code string) {
	formatted, err := format(code)
	if err != nil {
		log.Println("Error formatting C++ with clang-format:", err)
		log.Println("Unformatted C++:")
		fmt.Println(code)
	} else {
		fmt.Println(formatted)
	}
}

func createFile(dir string, name string) *os.File {
	file, err := os.Create(path.Join(dir, name))
	if err != nil {
		log.Fatalln("Failed to create file during compilation process. Error:", err)
	}
	return file
}

func writeFile(dir string, name string, contents string) *os.File {
	file := createFile(dir, name)
	_, err := file.WriteString(contents)
	if err != nil {
		log.Fatalln("Failed to write to file during compilation process. Error:", err)
	}
	return file
}

func Run(module Module) {
	dir, err := os.MkdirTemp("", "yune-build")
	if err != nil {
		log.Fatalln("Failed to create temporary directory during compilation process. Error:", err)
	}
	defer os.RemoveAll(dir)

	header := module.GenHeader()
	implementation := module.String()
	writeFile(dir, "code.hpp", header)
	writeFile(dir, "code.cpp", "#include \"code.hpp\"\n"+implementation)
	implementationPath := path.Join(dir, "code.cpp")
	binaryPath := path.Join(dir, "code.bin")

	cmd := exec.Command("clang++", []string{implementationPath, "-o", binaryPath}...)
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Fatalln("Failed to compile code. Error:", err)
	}
	err = exec.Command(binaryPath).Run()
	if err != nil {
		log.Fatalln("Failed to run code. Error:", err)
	}
	// NOTE: main function is assumed to exist
}
