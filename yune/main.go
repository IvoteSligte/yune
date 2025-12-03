package main

import (
	"github.com/antlr4-go/antlr/v4"
	"log"
	"os"
	"yune/parser"
)

func main() {
	file, err := os.Open("test.un")
	if err != nil {
		log.Fatalln("Failed to open test.un. Error:", err)
	}
	var bytes []byte
	_, err = file.Read(bytes)
	if err != nil {
		log.Fatalln("Failed to read test.un. Error:", err)
	}
	inputStream := antlr.NewInputStream(string(bytes))
	lexer := parser.NewYuneLexer(inputStream)
	tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := parser.NewYuneParser(tokenStream)
	module := parser.Module()
	println(module)
}
