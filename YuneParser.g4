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
    : name LPAREN functionParameters RPAREN typeAnnotation EQUAL statementBody
    ;

functionParameters
    : functionParameter (COMMA functionParameter)*
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
    | NEWLINE INDENT statementBlock DEDENT
    ;

statementBlock
    : statement+
    ;

statement
    : variableDeclaration NEWLINE
    | assignment          NEWLINE
    | branchStatement     // already has newline
    | expression          NEWLINE
    ;

variableDeclaration
    : LET constantDeclaration
    ;

assignment
    : variable assignmentOp EQUAL statementBody
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
