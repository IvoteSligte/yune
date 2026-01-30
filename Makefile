
all: parser

parser/yune_lexer.go: YuneLexer.g4
	antlr4 YuneLexer.g4 -Dlanguage=Go -no-visitor -no-listener -o parser

# depends on both parser and lexer grammars,
# since the parser depends on lexer-defined tokens
parser/yune_parser.go: YuneParser.g4 YuneLexer.g4
	antlr4 YuneParser.g4 -Dlanguage=Go -no-visitor -no-listener -o parser

parser: parser/yune_lexer.go parser/yune_parser.go

# compile Go Cap'n Proto files
pb/schema.capnp.go: schema.capnp
	capnp compile -ogo:pb schema.capnp

pb: pb/schema.capnp.go

run: parser pb
	go run .
