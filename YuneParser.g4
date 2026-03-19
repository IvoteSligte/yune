parser grammar YuneParser;

options { tokenVocab = YuneLexer; }

module
    : NEWLINE? anImport* topLevelDeclaration* EOF
    ;

// "import" is an ANTLR keyword
anImport
    : IMPORT STRING NEWLINE
    ;

topLevelDeclaration
    : functionDeclaration
    | constantDeclaration
    ;

name
    : IDENTIFIER
    ;

functionDeclaration
    : name functionParameters COLON type EQUAL statementBody
    ;

functionParameters
    : LPAREN RPAREN
    | LPAREN functionParameter (COMMA functionParameter)* RPAREN
    ;

functionParameter
    : name COLON type
    ;

constantDeclaration
    : name COLON type EQUAL statementBody
    ;

type: expression;

statementBody
    : statement
    | NEWLINE INDENT statement+ DEDENT
    ;

statement
    : variableDeclaration
    | assignment
    | branchStatement
    | closureExpression
    // only this should have a newline because all statements end in an expression
    | expression NEWLINE
    | isStatement
    ;

variableDeclaration
    : target EQUAL statementBody
    | name COLONEQUAL statementBody
    ;

target
    : name COLON type
    | LPAREN RPAREN
    | LPAREN name COLON type (COMMA name COLON type)* RPAREN
    ;

assignment
    : variable assignmentOp statementBody
    ;

assignmentOp
    : EQUAL
    | PLUSEQUAL
    | MINUSEQUAL
    | STAREQUAL
    | SLASHEQUAL
    ;

// Expressions that are never ambiguous, so they never need to be wrapped in parentheses.
primaryExpression
    : INTEGER
    | FLOAT
    | STRING
    | bool=(TRUE | FALSE)
    | variable
    | parenExpression
    | list
    | macro
    ;

parenExpression
    : LPAREN expression RPAREN
    | tuple
    ;

variable
    : name
    ;

list
    : LBRACKET RBRACKET
    | LBRACKET expression (COMMA expression)* COMMA? RBRACKET
    ;

tuple
    : LPAREN RPAREN
    // disallows "LPAREN expression RPAREN" as unit tuples do not exist
    | LPAREN expression (COMMA expression)+ COMMA? RPAREN
    ;

macro
    : variable HASHTAG (MACROLINE NEWLINE)* MACROLINE
    ;

closure
    : closureParameters COLON type EQUAL statementBody
    ;

closureParameters
    : BAR BAR
    | BAR functionParameter (COMMA functionParameter)* BAR
    ;

binaryExpression
    : primary=primaryExpression
    | unaryOp=(MINUS | SEMI) primaryExpression // unary expression
    | function=primaryExpression argument=parenExpression // function call
    | left=binaryExpression op=(STAR | SLASH) right=binaryExpression
    | left=binaryExpression op=(PLUS | MINUS) right=binaryExpression
    | left=binaryExpression op=(LESS | GREATER) right=binaryExpression
    | left=binaryExpression op=(LESSEQUAL | GREATEREQUAL) right=binaryExpression
    | left=binaryExpression op=(EQEQUAL | NOTEQUAL) right=binaryExpression
    | left=binaryExpression op=AND right=binaryExpression
    | left=binaryExpression op=OR right=binaryExpression
    ;

// Expressions that may not consume the following newline.
expression
    : binaryExpression
    | function=primaryExpression argument=expression // function call such as `func 5 + 6`
    ;

// Expressions that consume the following newline.
closureExpression
    : closure
    | unaryOp=(MINUS | SEMI) closureExpression // unary expression
    | function=primaryExpression argument=closureExpression // function call
    | left=binaryExpression op=(STAR | SLASH) right=closureExpression
    | left=binaryExpression op=(PLUS | MINUS) right=closureExpression
    | left=binaryExpression op=(LESS | GREATER) right=closureExpression
    | left=binaryExpression op=(LESSEQUAL | GREATEREQUAL) right=closureExpression
    | left=binaryExpression op=(EQEQUAL | NOTEQUAL) right=closureExpression
    | left=binaryExpression op=AND right=closureExpression
    | left=binaryExpression op=OR right=closureExpression
    ;

block
    : statement*
    ;

isExpression
    : expression IS name COLON type
    ;

isStatement
    : isExpression NEWLINE block
    ;

branchStatement
    : (expression | isExpression) RARROW statementBody block
    ;
