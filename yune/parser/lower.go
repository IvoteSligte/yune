package parser

import (
	"strconv"
	"strings"
	"yune/ast"

	"github.com/antlr4-go/antlr/v4"
)

func Map[T, V any](slice []T, function func(T) V) []V {
	result := make([]V, len(slice))
	for i, t := range slice {
		result[i] = function(t)
	}
	return result
}

func FlatMap[T, V any](slice []T, function func(T) []V) []V {
	result := make([]V, len(slice))
	for _, t := range slice {
		result = append(result, function(t)...)
	}
	return result
}

func Prepend[T any](element T, slice []T) []T {
	return append([]T{element}, slice...)
}

func LowerAssignment(ctx IAssignmentContext) ast.Assignment {
	return ast.Assignment{
		Target: LowerName(ctx.Name()),
		Op:     LowerAssignmentOp(ctx.AssignmentOp()),
		Body:   LowerStatementBody(ctx.StatementBody()),
	}
}

func LowerAssignmentOp(ctx IAssignmentOpContext) ast.AssignmentOp {
	switch {
	case ctx.EQUAL() != nil:
		return ast.Assign
	case ctx.PLUSEQUAL() != nil:
		return ast.AddAssign
	case ctx.MINUSEQUAL() != nil:
		return ast.SubtractAssign
	case ctx.STAREQUAL() != nil:
		return ast.MultiplyAssign
	case ctx.SLASHEQUAL() != nil:
		return ast.DivideAssign
	default:
		panic("unreachable")
	}
}

func LowerBinaryExpression(ctx IBinaryExpressionContext) any {
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
	return ast.BinaryExpression{
		Op:    op,
		Left:  LowerBinaryExpression(ctx.BinaryExpression(0)),
		Right: LowerBinaryExpression(ctx.BinaryExpression(1)),
	}
}

func LowerBranchStatement(ctx IBranchStatementContext, statementsAfter []IStatementContext) ast.BranchStatement {
	elseBlock := []ast.Statement{}
	if len(statementsAfter) > 0 {
		elseBlock = LowerStatement(statementsAfter[0], statementsAfter[1:])
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

func LowerFunctionCall(ctx IFunctionCallContext) ast.FunctionCall {
	return ast.FunctionCall{
		Function: LowerName(ctx.Name()),
		Argument: LowerPrimaryExpression(ctx.PrimaryExpression()),
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
		Name: LowerName(ctx.Name()),
		Type: LowerTypeAnnotation(ctx.TypeAnnotation()),
	}
}

func LowerFunctionParameters(ctx IFunctionParametersContext) []ast.FunctionParameter {
	return Map(ctx.AllFunctionParameter(), LowerFunctionParameter)
}

func LowerName(ctx INameContext) string {
	return ctx.IDENTIFIER().GetText()
}

func LowerMacro(ctx IMacroContext) ast.Macro {
	return ast.Macro{
		Language: LowerName(ctx.Name()),
		Text:     strings.Join(Map(ctx.AllMACROLINE(), antlr.TerminalNode.GetText), "\n"),
	}
}

func LowerModule(ctx IModuleContext) ast.Module {
	return ast.Module{
		Declarations: Map(ctx.AllTopLevelDeclaration(), LowerTopLevelDeclaration),
	}
}

func LowerPrimaryExpression(ctx IPrimaryExpressionContext) any {
	switch {
	case ctx.FunctionCall() != nil:
		return LowerFunctionCall(ctx.FunctionCall())
	case ctx.Name() != nil:
		return LowerName(ctx.Name())
	case ctx.INTEGER() != nil:
		integer, err := strconv.ParseInt(ctx.INTEGER().GetText(), 0, 64)
		if err != nil {
			panic("ANTLR parser-accepted integer failed to parse by strconv.ParseInt")
		}
		return integer
	case ctx.FLOAT() != nil:
		float, err := strconv.ParseFloat(ctx.FLOAT().GetText(), 64)
		if err != nil {
			panic("ANTLR parser-accepted float failed to parse by strconv.ParseFloat")
		}
		return float
	case ctx.Expression() != nil:
		return LowerExpression(ctx.Expression())
	case ctx.Tuple() != nil:
		return LowerTuple(ctx.Tuple())
	case ctx.Macro() != nil:
		return LowerMacro(ctx.Macro())
	default:
		panic("unreachable")
	}
}

func LowerStatement(ctx IStatementContext, statementsAfter []IStatementContext) []ast.Statement {
	var single ast.Statement

	switch {
	case ctx.VariableDeclaration() != nil:
		single = LowerVariableDeclaration(ctx.VariableDeclaration())
	case ctx.Assignment() != nil:
		single = LowerAssignment(ctx.Assignment())
	case ctx.Expression() != nil:
		single = LowerExpression(ctx.Expression())
	case ctx.BranchStatement() != nil:
		return []ast.Statement{LowerBranchStatement(ctx.BranchStatement(), statementsAfter)}
	default:
		panic("unreachable")
	}
	if len(statementsAfter) == 0 {
		return []ast.Statement{single}
	}
	return Prepend(single, LowerStatement(statementsAfter[0], statementsAfter[1:]))
}

func LowerStatementBlock(ctx IStatementBlockContext) []ast.Statement {
	if len(ctx.AllStatement()) == 0 {
		panic("A statement block should contain at least one statement.")
	}
	return LowerStatement(ctx.AllStatement()[0], ctx.AllStatement()[1:])
}

func LowerStatementBody(ctx IStatementBodyContext) []ast.Statement {
	if ctx.Statement() != nil {
		return LowerStatement(ctx.Statement(), []IStatementContext{})
	}
	return LowerStatementBlock(ctx.StatementBlock())
}

func LowerTopLevelDeclaration(ctx ITopLevelDeclarationContext) ast.TopLevelDeclaration {
	switch {
	case ctx.ConstantDeclaration() != nil:
		return LowerConstantDeclaration(ctx.ConstantDeclaration())
	case ctx.FunctionDeclaration() != nil:
		return LowerFunctionDeclaration(ctx.FunctionDeclaration())
	default:
		panic("unreachable(" + ctx.GetText() + ")")
	}
}

func LowerTuple(ctx ITupleContext) []ast.Expression {
	return Map(ctx.AllExpression(), LowerExpression)
}

func LowerTypeAnnotation(ctx ITypeAnnotationContext) ast.Type {
	return ast.Type(LowerName(ctx.Name()))
}

func LowerUnaryExpression(ctx IUnaryExpressionContext) any {
	switch {
	case ctx.MINUS() != nil:
		return ast.UnaryExpression{
			Op:         ast.Negate,
			Expression: LowerPrimaryExpression(ctx.PrimaryExpression()),
		}
	case ctx.PrimaryExpression() != nil:
		return LowerPrimaryExpression(ctx.PrimaryExpression())
	default:
		panic("unreachable")
	}
}

func LowerVariableDeclaration(ctx IVariableDeclarationContext) ast.VariableDeclaration {
	return ast.VariableDeclaration{
		ConstantDeclaration: LowerConstantDeclaration(ctx.ConstantDeclaration()),
	}
}
