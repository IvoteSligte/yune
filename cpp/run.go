package cpp

import (
	"fmt"
	"io"
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
		fmt.Println("Unformatted C++:")
		fmt.Println(code)
	} else {
		fmt.Println(formatted)
	}
}

func createFile(dir, name string) *os.File {
	file, err := os.Create(path.Join(dir, name))
	if err != nil {
		log.Fatalln("Failed to create file during compilation process. Error:", err)
	}
	return file
}

func writeFile(dir, name, contents string) *os.File {
	file := createFile(dir, name)
	_, err := file.WriteString(contents)
	if err != nil {
		log.Fatalln("Failed to write to file during compilation process. Error:", err)
	}
	return file
}

func CompileLibrary(module Module) {
	dir, err := os.MkdirTemp("", "yune-build")
	if err != nil {
		log.Fatalln("Failed to create temporary directory during compilation process. Error:", err)
	}
	defer os.RemoveAll(dir)

	writeFile(dir, "code.cpp", "#define RUNTIME\n"+module)

	implementationPath := path.Join(dir, "code.cpp")
	libraryPath := "./library.o"

	fmt.Println("-- Clang++ log --")
	includes := os.ExpandEnv("-I$PWD/cpp")
	// Creates a C++23 object file
	cmd := exec.Command("clang++", "-O1", "-c", "-std=c++23", implementationPath, "-o", libraryPath, includes)
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Fatalln("Failed to compile code. Error:", err)
	}
	fmt.Println("-- Compilation Finished --")
	return
}

func Run(module Module) (stdout, stderr string) {
	// NOTE: main function is assumed to exist

	dir, err := os.MkdirTemp("", "yune-build")
	if err != nil {
		log.Fatalln("Failed to create temporary directory during compilation process. Error:", err)
	}
	defer os.RemoveAll(dir)

	writeFile(dir, "code.cpp", "#define RUNTIME\n"+module+`
int main() {
    main_();
    return 0;
}`)

	implementationPath := path.Join(dir, "code.cpp")
	binaryPath := "./program"

	fmt.Println("-- Clang++ log --")
	includes := os.ExpandEnv("-I$PWD/cpp")
	cmd := exec.Command("clang++", "-O1", "-std=c++23", implementationPath, "-o", binaryPath, includes)
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Fatalln("Failed to compile code. Error:", err)
	}
	fmt.Println("-- Output --")
	stdoutWriter := strings.Builder{}
	stderrWriter := strings.Builder{}
	cmd = exec.Command(binaryPath)
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutWriter)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrWriter)
	err = cmd.Run()
	if err != nil {
		log.Fatalln("Failed to run code. Error:", err)
	}
	fmt.Println("-- Completed --")
	stdout = stdoutWriter.String()
	stderr = stderrWriter.String()
	return
}
