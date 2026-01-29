
all: parser

parser/yune_lexer.go: YuneLexer.g4
	antlr4 YuneLexer.g4 -Dlanguage=Go -no-visitor -no-listener -o parser

# depends on both parser and lexer grammars,
# since the parser depends on lexer-defined tokens
parser/yune_parser.go: YuneParser.g4 YuneLexer.g4
	antlr4 YuneParser.g4 -Dlanguage=Go -no-visitor -no-listener -o parser

parser: parser/yune_lexer.go parser/yune_parser.go

CPP_INCLUDES := cpp/includes

# protobuf files
PB_FILES := $(wildcard proto/*.proto) # source files
GO_PB_FILES := $(patsubst proto/%.proto,pb/%.pb.go,$(PB_FILES)) # generated Go
CPP_PB_FILES := $(patsubst proto/%.proto,$(CPP_INCLUDES)/proto/%.pb.cc,$(PB_FILES)) # generated C++

# compile Go protobuf files
pb/%.pb.go: proto/%.proto
	protoc -I=. --go_out=. $<

# compile C++ protobuf files
$(CPP_INCLUDES)/proto/%.pb.cc: proto/%.proto
	protoc -I=. --cpp_out=$(CPP_INCLUDES) $<

# shorthand for compiling all protobuf files that changed
protobuf: $(GO_PB_FILES) $(CPP_PB_FILES)

run: parser protobuf
	go run .
