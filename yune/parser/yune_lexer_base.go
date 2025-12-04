// partially adapted from https://github.com/antlr/grammars-v4/blob/master/python/python3/Go/python3_lexer_base.go
package parser

import (
	"log"

	"github.com/antlr4-go/antlr/v4"
)

// TODO: MACRO

type YuneLexerBase struct {
	*antlr.BaseLexer
	// Indentation in number of spaces.
	indent int
	// Indentation before the last token.
	oldIndent  int
	reachedEOF bool
}

func (l *YuneLexerBase) EmitToken(t antlr.Token) {
	log.Fatalln() // TODO
}

func (l *YuneLexerBase) Reset() {
	l.indent = 0
	l.oldIndent = 0
	l.reachedEOF = false
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

func (l *YuneLexerBase) NextToken() antlr.Token {
	if l.reachedEOF {
		if l.indent < l.oldIndent {
			l.indent -= 4
			return l.makeCommonToken(YuneParserDEDENT, "")
		}
		return l.makeCommonToken(YuneParserEOF, "<EOF>")
	}
	if l.indent < l.oldIndent {
		l.oldIndent -= 4
		return l.makeCommonToken(YuneParserDEDENT, "")
	}
	if l.indent > l.oldIndent {
		l.oldIndent += 4
		return l.makeCommonToken(YuneParserINDENT, "")
	}
	if l.GetInputStream().LA(1) == '#' {
		// offset := 2
		log.Fatalln("TODO")
	}
	token := l.BaseLexer.NextToken()

	for token.GetTokenType() == YuneParserNEWLINE {
		newlineToken := token
		token = l.BaseLexer.NextToken()
		newIndent := 0

		if token.GetTokenType() == YuneParserWHITESPACE {
			for _, c := range token.GetText() {
				switch c {
				case ' ':
					newIndent += 1
				case '\t':
					// next multiple of 4
					newIndent = (newIndent/4 + 1) * 4
				}
			}
		}
		if token.GetTokenType() == YuneParserNEWLINE {
			continue
		}
		if l.indent < newIndent && l.indent+4 != newIndent {
			// TODO: handle properly
			log.Fatalln("Indentation is not the next multiple of 4 from the previous indentation.")
		}
		l.oldIndent = l.indent
		l.indent = newIndent
		return newlineToken
	}
	if token.GetTokenType() == YuneParserEOF {
		l.reachedEOF = true
		if l.indent < l.oldIndent {
			l.indent -= 4
			return l.makeCommonToken(YuneParserDEDENT, "")
		}
	}
	return token
}
