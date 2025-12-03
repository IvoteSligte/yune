
all: parser/yune_lexer.go parser/yune_parser.go

parser/yune_lexer.go: YuneLexer.g4
	antlr YuneLexer.g4 -Dlanguage=Go -no-visitor -no-listener -o parser

parser/yune_parser.go: YuneParser.g4
	antlr YuneParser.g4 -Dlanguage=Go -no-visitor -no-listener -o parser
