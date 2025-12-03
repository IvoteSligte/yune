// partially adapted from https://github.com/antlr/grammars-v4/blob/master/python/python3/Go/python3_lexer_base.go
package parser

import (
	"github.com/antlr4-go/antlr/v4"
)

// TODO: MACRO

type YuneLexerBase struct {
	*antlr.BaseLexer
	// Indentation in number of 4-space indentations.
	indent int
	queue  []antlr.Token
}

func (l *YuneLexerBase) Emit() antlr.Token {
	return l.BaseLexer.Emit()
}

func (l *YuneLexerBase) EmitToken(t antlr.Token) {
	l.queue = append(l.queue, t)
}

func (l *YuneLexerBase) Reset() {
	l.indent = 0
	l.queue = []antlr.Token{}
	l.BaseLexer.Reset()
}

func (l *YuneLexerBase) MakeCommonToken(ttype int, text string) antlr.Token {
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

func (l *YuneLexerBase) EmitDedents(amount int) {
	for range l.indent {
		l.EmitToken(l.MakeCommonToken(YuneParserDEDENT, ""))
	}
}

func (l *YuneLexerBase) NextToken() antlr.Token {
	if l.GetInputStream().LA(1) == antlr.TokenEOF && l.indent > 0 {
		l.EmitDedents(l.indent)
		l.indent = 0
		l.EmitToken(l.MakeCommonToken(antlr.TokenEOF, "<EOF>"))
	}
	next := l.BaseLexer.NextToken()
	if len(l.queue) == 0 {
		return next
	} else {
		x := l.queue[0]
		l.queue = l.queue[1:]
		return x
	}
}

func (l *YuneLexerBase) GetIndentationCount(spaces string) int {
	count := 0
	for _, ch := range spaces {
		if ch == '\t' {
			count += 4 - (count % 4)
		} else {
			count += 1
		}
	}
	return count
}
