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
    : name functionParameters typeAnnotation EQUAL statementBody
    ;

functionParameters
    : LPAREN RPAREN
    | LPAREN functionParameter (COMMA functionParameter)* RPAREN
    ;

functionParameter
    : name typeAnnotation
    ;

constantDeclaration
    : name typeAnnotation EQUAL statementBody
    ;

typeAnnotation
    : COLON name
    ;

statementBody
    : statement
    | NEWLINE INDENT statement+ DEDENT
    ;

statement
    : variableDeclaration
    | assignment
    | branchStatement
    | expression NEWLINE
    ;

variableDeclaration
    : name typeAnnotation EQUAL statementBody
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
    | variable
    | INTEGER
    | FLOAT
    | LPAREN expression RPAREN
    | tuple
    | macro
    ;

variable
    : name
    ;

tuple
    : LPAREN RPAREN
    | LPAREN expression COMMA RPAREN
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
    ;

expression
    : binaryExpression
    ;

branchStatement
    : expression RARROW statementBody
    ;
