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
TRUE     : 'true';
FALSE    : 'false';

IDENTIFIER : [a-zA-Z][a-zA-Z0-9]*;

INTEGER    : [0-9]+;
FLOAT      : [0-9]+ '.' [0-9]+;

NEWLINE    : '\r'? '\n';

COMMENT    : '//' ~[\r\n]*                -> skip;
WHITESPACE : [ \t\f]+                     -> skip;

STRING     : '"' .*? '"';

HASHTAG : '#';
