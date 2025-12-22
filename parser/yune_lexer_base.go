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
		// FIXME: not accurate since they are interpreter-managed, not stream-managed
		l.GetLine(),
		l.GetCharPositionInLine())
	return t
}

func (l *YuneLexerBase) pushToken(token antlr.Token) {
	l.queue = append(l.queue, token)
}

func (l *YuneLexerBase) Consume() {
	l.Interpreter.Consume(l.GetInputStream())
}

// Handles lexing indentation on the new line.
func (l *YuneLexerBase) onNewline() (indent int) {
	input := l.GetInputStream()
	indent = 0
	for {
		switch input.LA(1) {
		case ' ':
			indent++
			l.Consume()
		case '\t':
			indent = (indent/4 + 1) * 4
			l.Consume()
		case '\n':
			indent = 0
			l.Consume()
		case EOF:
			return 0
		default:
			return indent
		}
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
			l.Consume()
			continue
		case '\t':
			indent = (indent/4 + 1) * 4
			l.Consume()
			continue
		case '\n':
			// empty lines are also passed to macros
			return l.indent
		case EOF:
			return 0
		}
		break
	}
	if indent%4 != 0 {
		log.Fatalln("Indentation is not a multiple of 4.")
	}
	return indent
}

func (l *YuneLexerBase) updateIndent(indent int) {
	if indent%4 != 0 {
		log.Fatalln("Indentation is not a multiple of 4.")
	}
	if indent < l.indent {
		for range (l.indent - indent) / 4 {
			l.pushToken(l.makeCommonToken(YuneParserDEDENT, "<DEDENT>"))
		}
	} else if indent > l.indent {
		if l.indent+4 != indent {
			log.Fatalln("Indentation is not the next multiple of 4.")
		}
		l.pushToken(l.makeCommonToken(YuneParserINDENT, "<INDENT>"))
	}
	l.indent = indent
}

func (l *YuneLexerBase) lexMacro() {
	// macros have increased indentation
	l.indent += 4
	input := l.GetInputStream()
	// parse the whole macro in here because macros can have empty lines,
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
				// remove indentation that was artificially added for the macro
				l.indent -= 4
				return
			}
		case EOF:
			l.pushToken(l.makeCommonToken(YuneParserMACROLINE, text))
			// remove indentation that was artificially added for the macro
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
	l.pushToken(token)
	switch token.GetTokenType() {
	case YuneParserHASHTAG:
		l.lexMacro()
	case YuneParserNEWLINE:
		l.updateIndent(l.onNewline())
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
