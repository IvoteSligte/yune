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
    | closureExpression
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

// FIXME: precedence is most likely very incorrect
binaryExpression
    : unaryExpression
    | left=binaryExpression op=(STAR | SLASH) right=binaryExpression
    | left=binaryExpression op=(PLUS | MINUS) right=binaryExpression
    | left=binaryExpression op=(LESS | GREATER) right=binaryExpression
    | left=binaryExpression op=(LESSEQUAL | GREATEREQUAL) right=binaryExpression
    | left=binaryExpression op=(EQEQUAL | NOTEQUAL) right=binaryExpression
    | left=binaryExpression op=AND right=binaryExpression
    | left=binaryExpression op=OR right=binaryExpression
    ;

expression: binaryExpression;

closureExpression
    : closure
    | function=primaryExpression argument=closureExpression
    | left=binaryExpression op=(STAR | SLASH) right=closureExpression
    | left=binaryExpression op=(PLUS | MINUS) right=closureExpression
    | left=binaryExpression op=(LESS | GREATER) right=closureExpression
    | left=binaryExpression op=(LESSEQUAL | GREATEREQUAL) right=closureExpression
    | left=binaryExpression op=(EQEQUAL | NOTEQUAL) right=closureExpression
    | left=binaryExpression op=AND right=closureExpression
    | left=binaryExpression op=OR right=closureExpression
    ;

elseBlock
    : statement*
    ;

branchStatement
    : expression RARROW statementBody elseBlock
    ;
