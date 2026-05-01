lexer grammar YuneLexer;

options { superClass=YuneLexerBase; }

tokens {
    INDENT, DEDENT, MACROLINE
}

LPAREN           : '(';
LBRACKET         : '[';
LBRACE           : '{';
RPAREN           : ')';
RBRACKET         : ']';
RBRACE           : '}';
BAR              : '|';
DOT              : '.';
COLON            : ':';
COMMA            : ',';
SEMI             : ';';
PLUS             : '+';
MINUS            : '-';
STAR             : '*';
SLASH            : '/';
LESS             : '<';
GREATER          : '>';
EQUAL            : '=';
EQEQUAL          : '==';
NOTEQUAL         : ';=';
LESSEQUAL        : '<=';
GREATEREQUAL     : '>=';
PLUSEQUAL        : '+=';
MINUSEQUAL       : '-=';
STAREQUAL        : '*=';
SLASHEQUAL       : '/=';
COLONEQUAL       : ':=';
RARROW           : '->';

IMPORT   : 'import';
IN       : 'in';
IS       : 'is';
AS       : 'as';
AND      : 'and';
OR       : 'or';
LET      : 'let';
VAR      : 'var';
CONST    : 'const';
TRUE     : 'true';
FALSE    : 'false';

IDENTIFIER : [a-zA-Z][a-zA-Z0-9]*|[A-Z]([A-Z0-9_]|[_][A-Z0-9])*;

INTEGER    : [0-9]+;
FLOAT      : [0-9]+ '.' [0-9]+;

NEWLINE    : '\r'? '\n';

COMMENT    : '//' ~[\r\n]*                -> skip;
WHITESPACE : [ \t\f]+                     -> skip;

STRING     : '"' (~["\\]|[\\].)* '"';

RAW_STRING : '`' .*? '`';

HASHTAG : '#';
