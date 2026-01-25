package cpp

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"yune/util"
	"yune/value"
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
	// NOTE: main function is assumed to exist

	dir, err := os.MkdirTemp("", "yune-build")
	if err != nil {
		log.Fatalln("Failed to create temporary directory during compilation process. Error:", err)
	}
	defer os.RemoveAll(dir)

	header := module.GenHeader()
	fmt.Println("-- Header --")
	fmt.Println(header)
	fmt.Println("-- End Header --")

	implementation := module.String()
	fmt.Println("-- Implementation --")
	fmt.Println(implementation)
	fmt.Println("-- End Implementation --")

	writeFile(dir, "code.hpp", header) // TODO: close files
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
}

func Evaluate(module Module, batch []Expression) []value.Value {
	// NOTE: main function is assumed not to exist

	fmt.Println("--- Start Evaluation ---")
	defer fmt.Println("--- End Evaluation ---")

	outputFile, err := os.CreateTemp("", "yune-eval")
	if err != nil {
		log.Fatalln("Failed to create temporary file during compile-time C++ evaluation. Error:", err)
	}
	defer os.Remove(outputFile.Name()) // TODO: close file

	statements := []Statement{
		Statement(RawCpp(fmt.Sprintf(`std::fstream outputFile("%s");`, outputFile.Name()))),
	}
	for _, e := range batch {
		if e != nil {
			statements = append(statements, Statement(RawCpp(`outputFile << `+e.String()+` << '\0';`)))
		}
	}
	module.Declarations = append(module.Declarations, FunctionDeclaration{
		Name:       "main",
		Parameters: []FunctionParameter{},
		ReturnType: "int",
		Body:       Block(statements),
	})
	Run(module)
	content, err := os.ReadFile(outputFile.Name())
	if err != nil {
		log.Fatalf("Failed to read compile-time evaluation output file '%s'. Error: %s", outputFile.Name(), err)
	}
	outputStrings := strings.Split(string(content), "\x00")
	if len(outputStrings) != len(batch) {
		log.Fatalf("Expected number of outputs of compile-time evaluation was '%d', found '%d'.", len(batch), len(outputStrings))
	}
	return util.Map(outputStrings, func(s string) value.Value {
		println(s) // TEMP
		return value.Value(s)
	})
}
