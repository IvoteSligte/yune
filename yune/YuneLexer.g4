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
VBAR             : '|';
AMPER            : '&';
LESS             : '<';
GREATER          : '>';
EQUAL            : '=';
PERCENT          : '%';
EQEQUAL          : '==';
NOTEQUAL         : '!=';
LESSEQUAL        : '<=';
GREATEREQUAL     : '>=';
TILDE            : '~';
CIRCUMFLEX       : '^';
LEFTSHIFT        : '<<';
RIGHTSHIFT       : '>>';
DOUBLESTAR       : '**';
PLUSEQUAL        : '+=';
MINEQUAL         : '-=';
STAREQUAL        : '*=';
SLASHEQUAL       : '/=';
PERCENTEQUAL     : '%=';
AMPEREQUAL       : '&=';
VBAREQUAL        : '|=';
CIRCUMFLEXEQUAL  : '^=';
LEFTSHIFTEQUAL   : '<<=';
RIGHTSHIFTEQUAL  : '>>=';
DOUBLESTAREQUAL  : '**=';
DOUBLESLASH      : '//';
DOUBLESLASHEQUAL : '//=';
AT               : '@';
ATEQUAL          : '@=';
RARROW           : '->';
ELLIPSIS         : '...';
COLONEQUAL       : ':=';
EXCLAMATION      : '!';

IMPORT   : 'import';
IN       : 'in';
AND      : 'and';
OR       : 'or';

IDENTIFIER : [a-zA-Z][a-zA-Z0-9]*;

NUMBER
    : INTEGER
    | FLOAT
    ;

INTEGER : [0-9]+;

FLOAT : [0-9]+ '.' [0-9]+;

STRING
    : STRING_LITERAL
    | BYTES_LITERAL
    ;

NEWLINE : '\r'? '\n';

COMMENT : "//" ~[\r\n]*                   -> channel(HIDDEN);

WHITESPACE : [ \t\f]+                     -> channel(HIDDEN);

STRING : '"' .*? '"';
