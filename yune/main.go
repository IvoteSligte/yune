package main

import (
	"fmt"
	"log"
	"os"
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

func printTokens(lexer antlr.Recognizer, tokenStream *antlr.CommonTokenStream) {
	maxLen := 0
	for _, symbol := range lexer.GetSymbolicNames() {
		maxLen = max(maxLen, len(symbol))
	}
	for _, token := range tokenStream.GetAllTokens() {
		symbol := "<EOF>"
		if token.GetTokenType() != antlr.TokenEOF {
			symbol = lexer.GetSymbolicNames()[token.GetTokenType()]
		}
		fmt.Printf("Token (%s) ", symbol)
		for range maxLen - len(symbol) {
			fmt.Print(" ")
		}
		fmt.Printf("'%s'\n",
			token.GetText(),
		)
	}
}

func main() {
	data := readFile()
	inputStream := antlr.NewInputStream(data)
	lexer := parser.NewYuneLexer(inputStream)
	tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	tokenStream.Fill() // lex all tokens in advance (for debugging)

	printText(tokenStream)
	printTokens(lexer, tokenStream)

	yuneParser := parser.NewYuneParser(tokenStream)
	parseTreeModule := yuneParser.Module()

	// FIXME: does not panic on recoverable error
	if yuneParser.HasError() {
		log.Fatalln("Parse error:", yuneParser.GetError())
	}
	_ = parser.LowerModule(parseTreeModule)
}
