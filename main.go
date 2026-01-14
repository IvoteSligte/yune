package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
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
	cpp.PrintFormatted(cppModule.GenHeader())
	fmt.Println("--- CPP implementation (should include header) ---")
	cpp.PrintFormatted(cppModule.String())
	fmt.Println("--- Output ---")
	cpp.Run(cppModule)
}
