package cpp

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"yune/pb"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
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
	PrintFormatted(header)
	fmt.Println("-- End Header --")

	implementation := module.String()
	fmt.Println("-- Implementation --")
	PrintFormatted(implementation)
	fmt.Println("-- End Implementation --")

	writeFile(dir, "code.hpp", header) // TODO: close files
	writeFile(dir, "code.cpp", "#include \"code.hpp\"\n"+implementation)
	implementationPath := path.Join(dir, "code.cpp")
	binaryPath := path.Join(dir, "code.bin")

	fmt.Println("-- Clang++ log --")
	cmd := exec.Command("clang++", []string{implementationPath, "-o", binaryPath}...)
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

// TODO: skip evaluation if batch is all-nil
func Evaluate(module Module, batch []Expression) []*anypb.Any {
	// NOTE: main function is assumed not to exist

	fmt.Println("--- Start Evaluation ---")
	defer fmt.Println("--- End Evaluation ---")

	outputFile, err := os.CreateTemp("", "yune-eval")
	if err != nil {
		log.Fatalln("Failed to create temporary file during compile-time C++ evaluation. Error:", err)
	}
	defer os.Remove(outputFile.Name()) // TODO: close file

	statements := []Statement{
		Statement(Raw(fmt.Sprintf(`std::ofstream outputFile("%s");`, outputFile.Name()))),
	}
	for _, e := range batch {
		statement := `outputFile << '\0';`
		if e != nil {
			statement = `outputFile << ` + e.String() + ` << '\0';`
		}
		statements = append(statements, Statement(Raw(statement)))
	}
	statements = append(statements, Statement(Raw(`outputFile.close();`)))

	module.Declarations = append(module.Declarations, FunctionDeclaration{
		Name:       "main",
		Parameters: []FunctionParameter{},
		ReturnType: "int",
		Body:       Block(statements),
	})
	Run(module)
	outputBytes, err := os.ReadFile(outputFile.Name())
	if err != nil {
		log.Fatalf("Failed to read compile-time evaluation output file '%s'. Error: %s", outputFile.Name(), err)
	}
	var pbMessages pb.Messages
	if err = proto.Unmarshal(outputBytes, &pbMessages); err != nil {
		log.Fatalf("Failed to parse compile-time evaluation outputs. Error: %s", err)
	}
	if len(pbMessages.Messages) != len(batch) {
		log.Fatalf("Expected number of outputs of compile-time evaluation was '%d', found '%d'.", len(batch), len(pbMessages.Messages))
	}
	return pbMessages.Messages
}
