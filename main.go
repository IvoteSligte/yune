package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"yune/ast"
	"yune/cpp"
	"yune/parser"

	"github.com/antlr4-go/antlr/v4"
)

func readFile(path string) string {
	bytes, err := os.ReadFile(path)
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

func loadModule(filePath string) ast.Module {
	sourceCode := readFile(filePath)
	inputStream := antlr.NewInputStream(sourceCode)
	lexer := parser.NewYuneLexer(inputStream)
	tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	// printTokens(lexer, tokenStream)

	_parser := parser.NewYuneParser(tokenStream)
	parseTreeModule := _parser.Module()

	// FIXME: does not panic on recoverable error
	if _parser.HasError() {
		log.Fatalln("Parse error:", _parser.GetError())
	}
	fmt.Println("Lowering Parse Tree to AST...")
	parser.FileName = filePath
	parser.SourceCode = sourceCode
	return parser.LowerModule(parseTreeModule)
}

func main() {
	stdAstModule := loadModule("std.un")
	astModule := loadModule("test.un")
	astModule = ast.JoinModules(stdAstModule, astModule)

	fmt.Println("Lowering AST to CPP...")
	cppModule, errors := astModule.Lower()
	if len(errors) > 0 {
		for _, err := range errors {
			fmt.Println("Error:", err)
		}
		log.Fatalln("Errors found, exiting.")
	}
	fmt.Println("--- Output ---")
	cpp.Run(cppModule)
}
