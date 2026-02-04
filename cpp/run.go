package cpp

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

//go:embed "json.hpp"
var jsonHeader string // nlohmann JSON library

//go:embed "pb.hpp"
var pbHeader string

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
		fmt.Println("Unformatted C++:")
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
	// NOTE: main function is assumed to exist

	dir, err := os.MkdirTemp("", "yune-build")
	if err != nil {
		log.Fatalln("Failed to create temporary directory during compilation process. Error:", err)
	}
	defer os.RemoveAll(dir)

	header := module.Header()
	fmt.Println("-- Header --")
	PrintFormatted(header)
	fmt.Println("-- End Header --")

	implementation := module.Implementation()
	fmt.Println("-- Implementation --")
	PrintFormatted(implementation)
	fmt.Println("-- End Implementation --")

	writeFile(dir, "json.hpp", jsonHeader)
	writeFile(dir, "pb.hpp", pbHeader)

	writeFile(dir, "code.h", `
#include "pb.hpp"
`+header) // TODO: close files

	writeFile(dir, "code.cpp", `
#include "code.h"
`+implementation)

	implementationPath := path.Join(dir, "code.cpp")
	binaryPath := path.Join(dir, "code.bin")

	fmt.Println("-- Clang++ log --")
	pbIncludes := os.ExpandEnv("-I$PWD/pb")
	cmd := exec.Command("clang++", []string{implementationPath, "-o", binaryPath, pbIncludes}...)
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Fatalln("Failed to compile code. Error:", err)
	}
	fmt.Println("-- Output --")
	cmd = exec.Command(binaryPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Fatalln("Failed to run code. Error:", err)
	}
	fmt.Println("-- Completed --")
}
