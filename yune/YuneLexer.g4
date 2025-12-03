lexer grammar YuneLexer;

options { superClass=YuneLexerBase; }

tokens {
    INDENT, DEDENT, MACRO
}

LPAREN           : '(';
LBRACKET         : '[';
LBRACE           : '{';
RPAREN           : ')';
RBRACKET         : ']';
RBRACE           : '}';
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
NOTEQUAL         : '!=';
LESSEQUAL        : '<=';
GREATEREQUAL     : '>=';
PLUSEQUAL        : '+=';
MINUSEQUAL       : '-=';
STAREQUAL        : '*=';
SLASHEQUAL       : '/=';
RARROW           : '->';

IMPORT   : 'import';
IN       : 'in';
AND      : 'and';
OR       : 'or';
LET      : 'let';
VAR      : 'var';
CONST    : 'const';

IDENTIFIER : [a-zA-Z][a-zA-Z0-9]*;

INTEGER : [0-9]+;

FLOAT : [0-9]+ '.' [0-9]+;

NEWLINE : '\r'? '\n';

COMMENT : '//' ~[\r\n]*                   -> channel(HIDDEN);

WHITESPACE : [ \t\f]+                     -> channel(HIDDEN);

STRING : '"' .*? '"';
