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
    | closure // FIXME: expressions following a closure on the next line (same indentation) might be interpreted as part of the same expression as the closure
    ;

variable
    : name
    ;

tuple
    : LPAREN RPAREN
    // disallows "LPAREN expression RPAREN" as unit tuples do not exist
    | LPAREN expression (COMMA expression)+ RPAREN
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

elseBlock
    : statement*
    ;

branchStatement
    : expression RARROW statementBody elseBlock
    ;
