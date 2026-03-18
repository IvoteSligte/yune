// partially adapted from https://github.com/antlr/grammars-v4/blob/master/python/python3/Go/python3_lexer_base.go
package parser

import (
	"log"

	"github.com/antlr4-go/antlr/v4"
)

// TODO: windows line breaks

const EOF = -1

type YuneLexerBase struct {
	*antlr.BaseLexer
	// Indentation in number of spaces.
	indent int
	// Difference in indentation from previous line.
	deltaIndent int
	queue       []antlr.Token
}

func (l *YuneLexerBase) Reset() {
	l.indent = 0
	l.deltaIndent = 0
	l.BaseLexer.Reset()
}

func (l *YuneLexerBase) makeCommonToken(ttype int, text string) antlr.Token {
	ctf := l.GetTokenFactory()
	index := l.GetInputStream().Index()
	t := ctf.Create(
		l.GetTokenSourceCharStreamPair(),
		ttype,
		text,
		antlr.TokenDefaultChannel,
		index-len(text),
		index,
		l.GetLine(),
		l.GetCharPositionInLine())
	return t
}

func (l *YuneLexerBase) pushToken(token antlr.Token) {
	l.queue = append(l.queue, token)
}

// Consumes a character.
// l.GetInputStream().Consume() messes up the line count, so use this function instead.
func (l *YuneLexerBase) Consume() {
	l.Interpreter.Consume(l.GetInputStream())
}

func (l *YuneLexerBase) skipLineComment() bool {
	input := l.GetInputStream()
	if input.LA(1) == '/' && input.LA(2) == '/' {
		l.Consume()
		l.Consume()
		for {
			c := input.LA(1)
			l.Consume()
			if c == '\n' || c == EOF {
				return true
			}
		}
	}
	return false
}

// Handles lexing indentation on the new line.
func (l *YuneLexerBase) onNewline() (indent int) {
	input := l.GetInputStream()
	indent = 0
	for {
		if l.skipLineComment() {
			indent = 0
			continue
		}
		switch input.LA(1) {
		case ' ':
			indent++
		case '\t':
			indent = (indent/4 + 1) * 4
		case '\r': // skip
		case '\n':
			indent = 0
		case EOF:
			return 0
		default:
			return indent

		}
		l.Consume()
	}
}

// Handles lexing indentation on the new line, assuming that the lexer
// is lexing a macro. Returns the new indentation level.
func (l *YuneLexerBase) onMacroNewline() (indent int) {
	input := l.GetInputStream()
	indent = 0
	for indent < l.indent {
		switch input.LA(1) {
		case ' ':
			indent++
		case '\t':
			// Round to next multiple of 4
			indent = (indent/4 + 1) * 4
		case '\r': // skip
		case '\n':
			// Empty lines are also passed to macros
			return l.indent
		case EOF:
			return 0
		default:
			goto end
		}
		l.Consume()
	}
end:
	if indent%4 != 0 {
		log.Fatalln("Indentation is not a multiple of 4.")
	}
	return indent
}

// Increase indentation by 4 spaces and emit an INDENT token.
func (l *YuneLexerBase) Indent() {
	l.indent += 4
	l.pushToken(l.makeCommonToken(YuneParserINDENT, "<INDENT>"))
}

// Decrease indentation by 4 spaces and emit a DEDENT token.
func (l *YuneLexerBase) Dedent() {
	l.indent -= 4
	l.pushToken(l.makeCommonToken(YuneParserDEDENT, "<DEDENT>"))
}

func (l *YuneLexerBase) updateIndent(indent int) {
	if indent%4 != 0 {
		log.Fatalf("Indentation %d is not a multiple of 4 on line %d.", indent, l.GetLine())
	}
	for l.indent > indent {
		l.Dedent()
	}
	if indent > l.indent {
		if l.indent+4 != indent {
			log.Fatalf("Indentation %d is not the next multiple of 4 on line %d.", indent, l.GetLine())
		}
		l.Indent()
	}
}

func (l *YuneLexerBase) lexMacro() {
	// Macros have increased indentation
	l.indent += 4
	input := l.GetInputStream()
	// Parse the whole macro in here because macros can have empty lines,
	// but ANTLR cannot handle MACROLINE lexing the empty string
	text := ""
	for {
		c := input.LA(1)
		switch c {
		case '\n':
			l.pushToken(l.makeCommonToken(YuneParserMACROLINE, text))
			text = ""
			l.pushToken(l.BaseLexer.NextToken())
			indent := l.onMacroNewline()
			if indent < l.indent {
				// Remove indentation that was artificially added for the macro
				l.indent -= 4
				return
			}
		case EOF:
			l.pushToken(l.makeCommonToken(YuneParserMACROLINE, text))
			// Remove indentation that was artificially added for the macro
			l.indent -= 4
			return
		default:
			text += string(rune(c))
			l.Consume()
		}
	}
}

func (l *YuneLexerBase) update() {
	token := l.BaseLexer.NextToken()
	switch token.GetTokenType() {
	case YuneParserHASHTAG:
		// Push token *before* lexing macro tokens
		l.pushToken(token)
		l.lexMacro()
	case YuneParserNEWLINE:
		// Push token *before* lexing indent tokens
		l.pushToken(token)
		l.updateIndent(l.onNewline())
	case YuneParserEOF:
		// Emit the required DEDENT tokens at EOF
		for l.indent > 0 {
			l.Dedent()
		}
		// Only push EOF *after* DEDENT tokens
		l.pushToken(token)
	default:
		l.pushToken(token)
	}
}

func (l *YuneLexerBase) NextToken() antlr.Token {
	if len(l.queue) == 0 {
		l.update()
	}
	token := l.queue[0]
	l.queue = l.queue[1:]
	return token
}
