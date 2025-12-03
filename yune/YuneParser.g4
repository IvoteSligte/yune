parser grammar YuneParser;

options { tokenVocab = YuneLexer; }

module
	: topLevelDeclaration+ EOF
    ;

topLevelDeclaration
    : functionDeclaration
    | constantDeclaration
    ;

functionDeclaration
    : IDENTIFIER LPAREN functionParameters RPAREN typeAnnotation statementBody
    ;

functionParameters
    : functionParameter (COMMA functionParameter)*
    ;

functionParameter
    : IDENTIFIER typeAnnotation
    ;

constantDeclaration
    : IDENTIFIER typeAnnotation statementBody
    ;

typeAnnotation
    : COLON IDENTIFIER
    ;

statementBody
    : EQUAL statement NEWLINE
    | EQUAL NEWLINE INDENT statementBlock DEDENT
    ;

statementBlock
    : (statement NEWLINE)+
    ;

statement
    : variableDeclaration
    | assignment
    | expression
    ;

variableDeclaration
    : LET constantDeclaration
    ;

assignment
    : IDENTIFIER assignmentOp statementBody
    ;

assignmentOp
    : PLUSEQUAL
    | MINUSEQUAL
    | STAREQUAL
    | SLASHEQUAL
    ;

primaryExpression
    : functionCall
    | IDENTIFIER
    | INTEGER
    | FLOAT
    | LPAREN expression RPAREN
    | tuple
    ;

functionCall
    : IDENTIFIER primaryExpression
    ;

unaryExpression
    : primaryExpression
    | MINUS primaryExpression
    ;

tuple
    : LPAREN RPAREN
    | LPAREN expression COMMA RPAREN
    | LPAREN expression (COMMA expression)+ COMMA? RPAREN
    ;

binaryExpression
    : unaryExpression
    | binaryExpression (STAR | SLASH) binaryExpression
    | binaryExpression (PLUS | MINUS) binaryExpression
    | binaryExpression (LESS | GREATER) binaryExpression
    | binaryExpression (LESSEQUAL | GREATEREQUAL) binaryExpression
    | binaryExpression (EQEQUAL | NOTEQUAL) binaryExpression
    | binaryExpression RARROW binaryExpression
    ;

expression
    : binaryExpression
    ;
