package parser

import (
	"strconv"
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
		Span:   GetSpan(ctx),
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
	case ctx.OR() != nil:
		op = ast.Or
	case ctx.AND() != nil:
		op = ast.And
	default:
		panic("unreachable")
	}
	expr := ast.BinaryExpression{
		Span:  GetSpan(ctx),
		Op:    op,
		Left:  LowerBinaryExpression(ctx.BinaryExpression(0)),
		Right: LowerBinaryExpression(ctx.BinaryExpression(1)),
	}
	return &expr
}

func LowerBranchStatement(ctx IBranchStatementContext) ast.BranchStatement {
	elseBlock := ast.Block{
		Span:       GetSpan(ctx.ElseBlock()),
		Statements: util.Map(ctx.ElseBlock().AllStatement(), LowerStatement),
	}
	return ast.BranchStatement{
		Span:      GetSpan(ctx),
		Condition: LowerExpression(ctx.Expression()),
		Then:      LowerStatementBody(ctx.StatementBody()),
		Else:      elseBlock,
	}
}

func LowerConstantDeclaration(ctx IConstantDeclarationContext) ast.ConstantDeclaration {
	return ast.ConstantDeclaration{
		Span: GetSpan(ctx),
		Name: LowerName(ctx.Name()),
		Type: LowerType(ctx.Type_()),
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
		Span:       GetSpan(ctx),
		Name:       LowerName(ctx.Name()),
		Parameters: LowerFunctionParameters(ctx.FunctionParameters()),
		ReturnType: LowerType(ctx.Type_()),
		Body:       LowerStatementBody(ctx.StatementBody()),
	}
}

func LowerFunctionParameter(ctx IFunctionParameterContext) ast.FunctionParameter {
	return ast.FunctionParameter{
		Span: GetSpan(ctx),
		Name: LowerName(ctx.Name()),
		Type: LowerType(ctx.Type_()),
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
		Span:     GetSpan(ctx),
		Function: LowerVariable(ctx.Variable()),
		Lines: util.Map(ctx.AllMACROLINE(), func(macroLine antlr.TerminalNode) ast.MacroLine {
			return ast.MacroLine{
				Span: ast.Span{
					Line:   macroLine.GetSymbol().GetStart(),
					Column: macroLine.GetSymbol().GetColumn(),
				},
				Text: macroLine.GetText(),
			}
		}),
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
		return &ast.Integer{
			Span:  GetSpan(ctx),
			Value: integer,
		}
	case ctx.FLOAT() != nil:
		float, err := strconv.ParseFloat(ctx.FLOAT().GetText(), 64)
		if err != nil {
			panic("ANTLR parser-accepted float failed to parse by strconv.ParseFloat")
		}
		return &ast.Float{
			Span:  GetSpan(ctx),
			Value: float,
		}
	case ctx.GetBool_() != nil:
		return &ast.Bool{
			Span:  GetSpan(ctx),
			Value: ctx.GetBool_().GetText() == "true",
		}
	case ctx.STRING() != nil:
		s := ctx.STRING().GetText()
		return &ast.String{
			Span:  GetSpan(ctx),
			Value: s[1 : len(s)-1], // strip ""
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

func LowerStatement(ctx IStatementContext) ast.Statement {
	switch {
	case ctx.VariableDeclaration() != nil:
		stmt := LowerVariableDeclaration(ctx.VariableDeclaration())
		return &stmt
	case ctx.Assignment() != nil:
		stmt := LowerAssignment(ctx.Assignment())
		return &stmt
	case ctx.Expression() != nil:
		return &ast.ExpressionStatement{
			Expression: LowerExpression(ctx.Expression()),
		}
	case ctx.BranchStatement() != nil:
		stmt := LowerBranchStatement(ctx.BranchStatement())
		util.PrettyPrint(stmt)
		return &stmt
	default:
		panic("unreachable")
	}
}

func LowerStatementBody(ctx IStatementBodyContext) ast.Block {
	return ast.Block{
		Span:       GetSpan(ctx),
		Statements: util.Map(ctx.AllStatement(), LowerStatement),
	}
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
		Span:     GetSpan(ctx),
		Elements: util.Map(ctx.AllExpression(), LowerExpression),
	}
}

func LowerType(ctx ITypeContext) ast.Type {
	return ast.Type{
		Expression: LowerExpression(ctx.Expression()),
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
		Name: LowerName(ctx.Name()),
		Type: LowerType(ctx.Type_()),
		Body: LowerStatementBody(ctx.StatementBody()),
	}

}
