// partially adapted from https://github.com/antlr/grammars-v4/blob/master/python/python3/Go/python3_lexer_base.go
package parser

import (
	"log"
	"math"

	"github.com/antlr4-go/antlr/v4"
)

// TODO: windows line breaks

const EOF = -1

type YuneLexerBase struct {
	*antlr.BaseLexer
	// Indentation in number of spaces.
	indent int
	// Indentation of the last token.
	prevIndent int
	eofToken   antlr.Token
	inMacro    bool
}

func (l *YuneLexerBase) EmitToken(t antlr.Token) {
	log.Fatalln() // TODO
}

func (l *YuneLexerBase) Reset() {
	l.indent = 0
	l.prevIndent = 0
	l.eofToken = nil
	l.BaseLexer.Reset()
}

func (l *YuneLexerBase) makeCommonToken(ttype int, text string) antlr.Token {
	stop := l.TokenStartCharIndex - 1
	start := stop
	if len(text) != 0 {
		start = stop - len(text) + 1
	}
	ctf := l.GetTokenFactory()
	t := ctf.Create(
		l.GetTokenSourceCharStreamPair(),
		ttype,
		text,
		antlr.TokenDefaultChannel,
		start,
		l.TokenStartCharIndex-1,
		l.TokenStartLine,
		l.TokenStartColumn)
	return t
}

func (l *YuneLexerBase) updateIndent(newIndent int) {
	if newIndent%4 != 0 {
		// TODO: handle properly
		log.Fatalln("Indentation is not a multiple of 4.")
	}
	if l.indent < newIndent && l.indent+4 != newIndent {
		// TODO: handle properly
		log.Fatalln("Indentation is not the next multiple of 4 from the previous indentation.")
	}
	l.prevIndent = l.indent
	l.indent = newIndent
}

func (l *YuneLexerBase) NextToken() antlr.Token {
	if l.indent < l.prevIndent {
		l.prevIndent -= 4
		return l.makeCommonToken(YuneParserDEDENT, "")
	}
	if l.indent > l.prevIndent {
		l.prevIndent += 4
		return l.makeCommonToken(YuneParserINDENT, "")
	}
	if l.eofToken != nil {
		return l.eofToken
	}
	if l.inMacro {
		input := l.GetInputStream()
		c := input.LA(1)
		if c == '\n' {
			token := l.BaseLexer.NextToken()
			indent := l.takeAtMostIndent(l.indent + 4)
			// macro ends on dedent
			if indent <= l.indent {
				l.updateIndent(indent)
				l.inMacro = false
				return token
			}
		}
		if c == EOF {
			return l.BaseLexer.NextToken()
		}
		line := l.takeLine()
		return l.makeCommonToken(YuneParserMACROLINE, line)
	}
	token := l.BaseLexer.NextToken()
	switch token.GetTokenType() {
	case YuneParserHASHTAG:
		l.inMacro = true
		return token
	case YuneParserNEWLINE:
		l.updateIndent(l.takeAtMostIndent(math.MaxInt))
		return token
	case YuneParserEOF:
		l.eofToken = token
		if l.indent < l.prevIndent {
			l.prevIndent -= 4
			return l.makeCommonToken(YuneParserDEDENT, "")
		}
	}
	return token
}

func (l *YuneLexerBase) takeAtMostIndent(max int) int {
	if max == 0 {
		return 0
	}
	input := l.GetInputStream()
	indent := 0
	for {
		switch input.LA(1) {
		case ' ':
			indent += 1
		case '\t':
			// next multiple of 4
			indent = (indent/4 + 1) * 4
		case '\n':
			input.Consume()
			indent = 0
			continue
		default:
			return indent
		}
		if indent >= max {
			return indent
		}
		input.Consume()
	}
}

// FIXME: invalid in BaseLexer.NextToken(): b.TokenStartColumn = b.Interpreter.GetCharPositionInLine()
// FIXME: invalid in BaseLexer.NextToken(): b.TokenStartLine = b.Interpreter.GetLine()

func (l *YuneLexerBase) takeLine() string {
	input := l.GetInputStream()
	text := ""
	for {
		c := input.LA(1)
		if c == EOF || c == '\n' {
			break
		}
		input.Consume()
		text += string(rune(c))
	}
	return text
}
