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
    : name assignmentOp EQUAL statementBody
    ;

assignmentOp
    : EQUAL
    | PLUSEQUAL
    | MINUSEQUAL
    | STAREQUAL
    | SLASHEQUAL
    ;

primaryExpression
    : functionCall
    | name
    | INTEGER
    | FLOAT
    | LPAREN expression RPAREN
    | tuple
    | macro
    ;

functionCall
    : name primaryExpression
    ;

tuple
    : LPAREN RPAREN
    | LPAREN expression COMMA RPAREN
    | LPAREN expression (COMMA expression)+ COMMA? RPAREN
    ;

macro
    : name HASHTAG (MACROLINE NEWLINE)* MACROLINE?
    ;

unaryExpression
    : primaryExpression
    | MINUS primaryExpression
    ;

binaryExpression
    : unaryExpression
    | binaryExpression (STAR | SLASH) binaryExpression
    | binaryExpression (PLUS | MINUS) binaryExpression
    | binaryExpression (LESS | GREATER) binaryExpression
    | binaryExpression (LESSEQUAL | GREATEREQUAL) binaryExpression
    | binaryExpression (EQEQUAL | NOTEQUAL) binaryExpression
    ;

expression
    : binaryExpression
    ;

branchStatement
    : expression RARROW statementBody
    ;
