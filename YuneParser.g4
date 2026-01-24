parser grammar YuneParser;

options { tokenVocab = YuneLexer; }

module
    : topLevelDeclaration* EOF
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
    // only this should have a newline because all statements end in an expression
    | expression NEWLINE
    ;

variableDeclaration
    : name COLON type EQUAL statementBody
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

primaryExpression
    : function=primaryExpression argument=primaryExpression
    | INTEGER
    | FLOAT
    | STRING
    | bool=(TRUE | FALSE)
    | variable
    | LPAREN expression RPAREN
    | tuple
    | macro
    ;

variable
    : name
    ;

tuple
    : LPAREN RPAREN
    | LPAREN expression (COMMA expression)+ COMMA? RPAREN
    ;

macro
    : variable HASHTAG (MACROLINE NEWLINE)* MACROLINE?
    ;

unaryExpression
    : primaryExpression
    | op=MINUS primaryExpression
    ;

binaryExpression
    : unaryExpression
    | binaryExpression op=(STAR | SLASH) binaryExpression
    | binaryExpression op=(PLUS | MINUS) binaryExpression
    | binaryExpression op=(LESS | GREATER) binaryExpression
    | binaryExpression op=(LESSEQUAL | GREATEREQUAL) binaryExpression
    | binaryExpression op=(EQEQUAL | NOTEQUAL) binaryExpression
    | binaryExpression op=AND binaryExpression
    | binaryExpression op=OR binaryExpression
    ;

expression
    : binaryExpression
    ;

branchStatement
    : expression RARROW statementBody
    ;
