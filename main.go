package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"yune/cpp"
	"yune/parser"

	"github.com/antlr4-go/antlr/v4"
)

func readFile() string {
	bytes, err := os.ReadFile("./test.un")
	if err != nil {
		log.Fatalln("Failed to open test.un. Error:", err)
	}
	return string(bytes)
}

func printText(tokenStream *antlr.CommonTokenStream) {
	println("```")
	println(tokenStream.GetAllText())
	println("```")
}

func printPadding(amount int) {
	for range amount {
		fmt.Print(" ")
	}
}

func printTokens(lexer antlr.Recognizer, tokenStream *antlr.CommonTokenStream) {
	tokenStream.Fill()
	maxSymbolLen := 0
	for _, symbol := range lexer.GetSymbolicNames() {
		maxSymbolLen = max(maxSymbolLen, len(symbol))
	}
	maxColumnStrLen := 0
	for _, token := range tokenStream.GetAllTokens() {
		maxColumnStrLen = max(maxColumnStrLen, len(strconv.Itoa(token.GetColumn())))
	}

	lastToken := tokenStream.Get(tokenStream.Size() - 1)
	maxLineStrLen := len(strconv.Itoa(lastToken.GetLine()))
	for _, token := range tokenStream.GetAllTokens() {
		symbol := "<EOF>"
		if token.GetTokenType() != antlr.TokenEOF {
			symbol = lexer.GetSymbolicNames()[token.GetTokenType()]
		}
		fmt.Print("Token ")
		fmt.Printf("%d", token.GetLine())
		fmt.Printf(":%d", token.GetColumn())
		printPadding((maxLineStrLen - len(strconv.Itoa(token.GetLine()))) +
			(maxColumnStrLen - len(strconv.Itoa(token.GetColumn()))))
		fmt.Printf(" (%s) ", symbol)
		printPadding(maxSymbolLen - len(symbol))
		fmt.Printf("'%s'\n", token.GetText())
	}
}

func formatCpp(code string) (formatted string, err error) {
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

func printCpp(code string) {
	formatted, err := formatCpp(code)
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

func writeTempFile(dir string, name string, contents string) *os.File {
	file := createFile(dir, name)
	_, err := file.WriteString(contents)
	if err != nil {
		log.Fatalln("Failed to write to file during compilation process. Error:", err)
	}
	return file
}

func runCppModule(module cpp.Module) {
	dir, err := os.MkdirTemp("", "yune-build")
	if err != nil {
		log.Fatalln("Failed to create temporary directory during compilation process. Error:", err)
	}
	defer os.RemoveAll(dir)

	header := module.GenHeader()
	implementation := module.String()
	writeTempFile(dir, "code.hpp", header)
	writeTempFile(dir, "code.cpp", "#include \"code.hpp\"\n"+implementation)
	implementationPath := path.Join(dir, "code.cpp")
	binaryPath := path.Join(dir, "code.bin")

	cmd := exec.Command("clang", []string{implementationPath, "-o", binaryPath}...)
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

func main() {
	data := readFile()
	println(data)
	inputStream := antlr.NewInputStream(data)
	lexer := parser.NewYuneLexer(inputStream)
	tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	printTokens(lexer, tokenStream)

	yuneParser := parser.NewYuneParser(tokenStream)
	parseTreeModule := yuneParser.Module()

	// FIXME: does not panic on recoverable error
	if yuneParser.HasError() {
		log.Fatalln("Parse error:", yuneParser.GetError())
	}
	fmt.Println("Lowering Parse Tree to AST...")
	astModule := parser.LowerModule(parseTreeModule)

	fmt.Println("Lowering AST to CPP...")
	cppModule, errors := astModule.Lower()
	if len(errors) > 0 {
		for _, err := range errors {
			fmt.Println("Error:", err)
		}
		log.Fatalln("Errors found, exiting.")
	}
	fmt.Println("--- CPP header ---")
	printCpp(cppModule.GenHeader())
	fmt.Println("--- CPP implementation (should include header) ---")
	printCpp(cppModule.String())
	fmt.Println("--- Output ---")
	runCppModule(cppModule)
}
