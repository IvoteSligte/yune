package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"yune/ast"
	"yune/cpp"
	"yune/parser"

	"github.com/antlr4-go/antlr/v4"
)

func readFile(path string) string {
	bytes, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to open file at '%s'. Error: %s", path, err)
	}
	return string(bytes)
}

type ParserErrorListener struct {
	*antlr.DefaultErrorListener
	Errors []error
}

// SyntaxError implements antlr.ErrorListener.
func (p *ParserErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol any, line int, column int, msg string, e antlr.RecognitionException) {
	p.Errors = append(p.Errors, fmt.Errorf("line %d:%d %s", line, column, msg))
}

var _ antlr.ErrorListener = (*ParserErrorListener)(nil)

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
		fmt.Printf("%s", strings.Repeat(" ", (maxLineStrLen-len(strconv.Itoa(token.GetLine())))+
			(maxColumnStrLen-len(strconv.Itoa(token.GetColumn())))))
		fmt.Printf(" (%s) ", symbol)
		fmt.Printf("%s", strings.Repeat(" ", maxSymbolLen-len(symbol)))
		fmt.Printf("'%s'\n", token.GetText())
	}
}

func loadModule(fileName string, sourceCode string) ast.Module {
	inputStream := antlr.NewInputStream(sourceCode)
	errorListener := ParserErrorListener{}
	lexer := parser.NewYuneLexer(inputStream)
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(&errorListener)
	tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	// printTokens(lexer, tokenStream)

	_parser := parser.NewYuneParser(tokenStream)
	_parser.RemoveErrorListeners()
	_parser.AddErrorListener(&errorListener)
	parseTreeModule := _parser.Module()

	if len(errorListener.Errors) > 0 {
		for _, error := range errorListener.Errors {
			log.Printf("Parse error in file '%s': %s\n", fileName, error)
		}
		log.Fatalf("%d parse errors found. Stopping compilation.", len(errorListener.Errors))
	}
	log.Printf("Lowering Parse Tree to AST for file '%s'...\n", fileName)
	parser.FileName = fileName
	parser.SourceCode = sourceCode
	module := parser.LowerModule(parseTreeModule)

	// NOTE: maybe move this to ast.Module.Lower?
	for _, _import := range module.Imports {
		// TODO: prevent import cycles
		// TODO: import resolution magic for standard library files?
		module = ast.JoinModules(module, loadModuleFromFile(_import))
	}
	return module
}

func loadModuleFromFile(filePath string) ast.Module {
	return loadModule(filePath, readFile(filePath))
}

func runModule(fileName string, sourceCode string) (stdout, stderr string) {
	// TODO: embedding
	astModule := loadModule(fileName, sourceCode)

	log.Printf("Lowering AST to CPP for file '%s'...\n", fileName)
	cppModule, hasMainFunction, errors := astModule.Lower()
	if len(errors) > 0 {
		for _, err := range errors {
			fmt.Println("Error:", err)
		}
		log.Fatalln("Errors found, exiting.")
	}
	fmt.Println("--- Output ---")
	if !hasMainFunction {
		fmt.Println("Module does not have a `main` function, so there is no output.")
		return
	} else {
		return cpp.Run(cppModule)
	}
}

func runModuleFromFile(filePath string) (stdout, stderr string) {
	return runModule(filePath, readFile(filePath))
}

func main() {
	filePath := "test.un"
	if len(os.Args) > 1 {
		filePath = os.Args[2]
	}
	runModuleFromFile(filePath)
}
