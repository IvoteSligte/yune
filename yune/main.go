package main

import (
	"log"
	"os"
	"yune/parser"

	"github.com/antlr4-go/antlr/v4"
)

func main() {
	bytes, err := os.ReadFile("./test.un")
	if err != nil {
		log.Fatalln("Failed to open test.un. Error:", err)
	}
	inputStream := antlr.NewInputStream(string(bytes))
	lexer := parser.NewYuneLexer(inputStream)
	tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	yuneParser := parser.NewYuneParser(tokenStream)
	parseTreeModule := yuneParser.Module()
	if yuneParser.HasError() {
		log.Fatalln("Parse error:", yuneParser.GetError())
	}
	_ = parser.LowerModule(parseTreeModule)
}
