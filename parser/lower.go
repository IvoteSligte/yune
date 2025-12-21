package parser

import (
	"strconv"
	"strings"
	"yune/ast"
	"yune/util"

	"github.com/antlr4-go/antlr/v4"
)

func GetSpan(ctx antlr.ParserRuleContext) ast.Span {
	return ast.Span{
		Line:   ctx.GetStart().GetLine(),
		Column: ctx.GetStart().GetColumn(),
	}
}

func LowerAssignment(ctx IAssignmentContext) ast.Assignment {
	return ast.Assignment{
		Target: LowerVariable(ctx.Variable()),
		Op:     LowerAssignmentOp(ctx.AssignmentOp()),
		Body:   LowerStatementBody(ctx.StatementBody()),
	}
}

func LowerAssignmentOp(ctx IAssignmentOpContext) ast.AssignmentOp {
	return ast.AssignmentOp(ctx.GetText())
}

func LowerBinaryExpression(ctx IBinaryExpressionContext) ast.Expression {
	var op ast.BinaryOp
	switch {
	case ctx.UnaryExpression() != nil:
		return LowerUnaryExpression(ctx.UnaryExpression())
	case ctx.STAR() != nil:
		op = ast.Multiply
	case ctx.SLASH() != nil:
		op = ast.Divide
	case ctx.PLUS() != nil:
		op = ast.Add
	case ctx.MINUS() != nil:
		op = ast.Subtract
	case ctx.LESS() != nil:
		op = ast.Less
	case ctx.GREATER() != nil:
		op = ast.Greater
	case ctx.LESSEQUAL() != nil:
		op = ast.LessEqual
	case ctx.GREATEREQUAL() != nil:
		op = ast.GreaterEqual
	case ctx.EQEQUAL() != nil:
		op = ast.Equal
	case ctx.NOTEQUAL() != nil:
		op = ast.NotEqual
	default:
		panic("unreachable")
	}
	expr := ast.BinaryExpression{
		Op:    op,
		Left:  LowerBinaryExpression(ctx.BinaryExpression(0)),
		Right: LowerBinaryExpression(ctx.BinaryExpression(1)),
	}
	return &expr
}

func LowerBranchStatement(ctx IBranchStatementContext, statementsAfter []IStatementContext) ast.BranchStatement {
	elseBlock := ast.Block{}
	if len(statementsAfter) > 0 {
		elseBlock = ast.Block{
			Statements: LowerStatement(statementsAfter[0], statementsAfter[1:]),
		}
	}
	return ast.BranchStatement{
		Condition: LowerExpression(ctx.Expression()),
		Then:      LowerStatementBody(ctx.StatementBody()),
		Else:      elseBlock,
	}
}

func LowerConstantDeclaration(ctx IConstantDeclarationContext) ast.ConstantDeclaration {
	return ast.ConstantDeclaration{
		Name: LowerName(ctx.Name()),
		Type: LowerTypeAnnotation(ctx.TypeAnnotation()),
		Body: LowerStatementBody(ctx.StatementBody()),
	}
}

func LowerExpression(ctx IExpressionContext) ast.Expression {
	switch {
	case ctx.BinaryExpression() != nil:
		return LowerBinaryExpression(ctx.BinaryExpression())
	default:
		panic("unreachable")
	}
}

func LowerFunctionDeclaration(ctx IFunctionDeclarationContext) ast.FunctionDeclaration {
	return ast.FunctionDeclaration{
		Name:       LowerName(ctx.Name()),
		Parameters: LowerFunctionParameters(ctx.FunctionParameters()),
		ReturnType: LowerTypeAnnotation(ctx.TypeAnnotation()),
		Body:       LowerStatementBody(ctx.StatementBody()),
	}
}

func LowerFunctionParameter(ctx IFunctionParameterContext) ast.FunctionParameter {
	return ast.FunctionParameter{
		Span: GetSpan(ctx),
		Name: LowerName(ctx.Name()),
		Type: LowerTypeAnnotation(ctx.TypeAnnotation()),
	}
}

func LowerFunctionParameters(ctx IFunctionParametersContext) []ast.FunctionParameter {
	return util.Map(ctx.AllFunctionParameter(), LowerFunctionParameter)
}

func LowerName(ctx INameContext) ast.Name {

	return ast.Name{
		Span:   GetSpan(ctx),
		String: ctx.IDENTIFIER().GetText(),
	}
}

func LowerVariable(ctx IVariableContext) ast.Variable {
	return ast.Variable{
		Name: LowerName(ctx.Name()),
	}
}

func LowerMacro(ctx IMacroContext) ast.Macro {
	return ast.Macro{
		Language: LowerVariable(ctx.Variable()),
		Text:     strings.Join(util.Map(ctx.AllMACROLINE(), antlr.TerminalNode.GetText), "\n"),
	}
}

func LowerModule(ctx IModuleContext) ast.Module {
	return ast.Module{
		Declarations: util.Map(ctx.AllTopLevelDeclaration(), LowerTopLevelDeclaration),
	}
}

func LowerPrimaryExpression(ctx IPrimaryExpressionContext) ast.Expression {
	switch {
	case ctx.GetFunction() != nil:
		return &ast.FunctionCall{
			Function: LowerPrimaryExpression(ctx.GetFunction()),
			Argument: LowerPrimaryExpression(ctx.GetArgument()),
		}
	case ctx.Variable() != nil:
		variable := LowerVariable(ctx.Variable())
		return &variable
	case ctx.INTEGER() != nil:
		integer, err := strconv.ParseInt(ctx.INTEGER().GetText(), 0, 64)
		if err != nil {
			panic("ANTLR parser-accepted integer failed to parse by strconv.ParseInt")
		}
		return ast.Integer{
			Span:  GetSpan(ctx),
			Value: integer,
		}
	case ctx.FLOAT() != nil:
		float, err := strconv.ParseFloat(ctx.FLOAT().GetText(), 64)
		if err != nil {
			panic("ANTLR parser-accepted float failed to parse by strconv.ParseFloat")
		}
		return ast.Float{
			Span:  GetSpan(ctx),
			Value: float,
		}
	case ctx.Expression() != nil:
		return LowerExpression(ctx.Expression())
	case ctx.Tuple() != nil:
		tuple := LowerTuple(ctx.Tuple())
		return &tuple
	case ctx.Macro() != nil:
		macro := LowerMacro(ctx.Macro())
		return &macro
	default:
		panic("unreachable")
	}
}

func LowerStatement(ctx IStatementContext, statementsAfter []IStatementContext) []ast.Statement {
	var single ast.Statement

	switch {
	case ctx.VariableDeclaration() != nil:
		stmt := LowerVariableDeclaration(ctx.VariableDeclaration())
		single = &stmt
	case ctx.Assignment() != nil:
		stmt := LowerAssignment(ctx.Assignment())
		single = &stmt
	case ctx.Expression() != nil:
		stmt := ast.ExpressionStatement{
			Expression: LowerExpression(ctx.Expression()),
		}
		single = &stmt
	case ctx.BranchStatement() != nil:
		stmt := LowerBranchStatement(ctx.BranchStatement(), statementsAfter)
		single = &stmt
	default:
		panic("unreachable")
	}
	if len(statementsAfter) == 0 {
		return []ast.Statement{single}
	}
	return util.Prepend(single, LowerStatement(statementsAfter[0], statementsAfter[1:]))
}

func LowerStatementBlock(ctx IStatementBlockContext) ast.Block {
	if len(ctx.AllStatement()) == 0 {
		panic("A statement block should contain at least one statement.")
	}
	return ast.Block{
		Statements: LowerStatement(ctx.AllStatement()[0], ctx.AllStatement()[1:]),
	}
}

func LowerStatementBody(ctx IStatementBodyContext) ast.Block {
	if ctx.Statement() != nil {
		return ast.Block{
			Statements: LowerStatement(ctx.Statement(), []IStatementContext{}),
		}
	}
	return LowerStatementBlock(ctx.StatementBlock())
}

func LowerTopLevelDeclaration(ctx ITopLevelDeclarationContext) ast.TopLevelDeclaration {
	switch {
	case ctx.ConstantDeclaration() != nil:
		decl := LowerConstantDeclaration(ctx.ConstantDeclaration())
		return &decl
	case ctx.FunctionDeclaration() != nil:
		decl := LowerFunctionDeclaration(ctx.FunctionDeclaration())
		return &decl
	default:
		panic("unreachable(" + ctx.GetText() + ")")
	}
}

func LowerTuple(ctx ITupleContext) ast.Tuple {
	return ast.Tuple{
		Elements: util.Map(ctx.AllExpression(), LowerExpression),
	}
}

func LowerTypeAnnotation(ctx ITypeAnnotationContext) ast.Type {
	return ast.Type{
		Name: LowerName(ctx.Name()),
		Span: GetSpan(ctx),
	}
}

func LowerUnaryExpression(ctx IUnaryExpressionContext) ast.Expression {
	switch {
	case ctx.MINUS() != nil:
		expr := ast.UnaryExpression{
			Span:       GetSpan(ctx),
			Op:         ast.Negate,
			Expression: LowerPrimaryExpression(ctx.PrimaryExpression()),
		}
		return &expr
	case ctx.PrimaryExpression() != nil:
		return LowerPrimaryExpression(ctx.PrimaryExpression())
	default:
		panic("unreachable")
	}
}

func LowerVariableDeclaration(ctx IVariableDeclarationContext) ast.VariableDeclaration {
	return ast.VariableDeclaration{
		Span: GetSpan(ctx),
		Name: LowerName(ctx.ConstantDeclaration().Name()),
		Type: LowerTypeAnnotation(ctx.ConstantDeclaration().TypeAnnotation()),
		Body: LowerStatementBody(ctx.ConstantDeclaration().StatementBody()),
	}

}
