package parser

import (
	"fmt"
	"iter"
	"math/rand/v2"
	"slices"
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
	if ctx.UnaryExpression() != nil {
		return LowerUnaryExpression(ctx.UnaryExpression())
	}
	return &ast.BinaryExpression{
		Span:  GetSpan(ctx),
		Op:    ast.BinaryOp(ctx.GetOp().GetText()),
		Left:  LowerBinaryExpression(ctx.GetLeft()),
		Right: LowerBinaryExpression(ctx.GetRight()),
	}
}

func DesugarIsExpression(ctx IIsExpressionContext) condition {
	// Input:
	// expression is name: type
	// Becomes:
	// is_expr_8fa932bc_: auto = expression
	// isVariant(type)(is_expr_8fa932bc_) ->
	//     name: type = getVariant(type)(is_expr_8fa932bc_)
	//     ...
}

func LowerBranchStatement(ctx IBranchStatementContext) ast.BranchStatement {
	switch {
	case ctx.Expression() != nil:
		return ast.BranchStatement{
			Span:      GetSpan(ctx),
			Condition: LowerExpression(ctx.Expression()),
			Then:      LowerStatementBody(ctx.StatementBody()),
			Else: ast.Block{
				Span:       GetSpan(ctx.ElseBlock()),
				Statements: LowerStatements(ctx.ElseBlock().AllStatement()),
			},
		}
	case ctx.IsExpression() != nil:
		condition := LowerExpression(ctx.IsExpression())
		thenBlock := LowerStatementBody(ctx.StatementBody())
		thenBlock.Statements = util.Prepend(thenBlock.Statements)
		elseBlock := ast.Block{
			Span:       GetSpan(ctx.ElseBlock()),
			Statements: LowerStatements(ctx.ElseBlock().AllStatement()),
		}
		return ast.BranchStatement{
			Span:      GetSpan(ctx),
			Condition: condition,
			Then:      thenBlock,
			Else:      elseBlock,
		}
	default:
		panic("unreachable")
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
		Parameters: util.Map(ctx.FunctionParameters().AllFunctionParameter(), LowerFunctionParameter),
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
	case ctx.Expression() != nil: // parses expression in parentheses: (expression)
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

func LowerClosureExpression(ctx IClosureExpressionContext) ast.Expression {
	if ctx.Closure() != nil {
		return LowerClosure(ctx.Closure())
	}
	return &ast.BinaryExpression{
		Span:  GetSpan(ctx),
		Op:    ast.BinaryOp(ctx.GetOp().GetText()),
		Left:  LowerBinaryExpression(ctx.GetLeft()),
		Right: LowerClosureExpression(ctx.GetRight()),
	}
}

func LowerStatement(ctx IStatementContext) iter.Seq[ast.Statement] {
	return func(yield func(ast.Statement) bool) {
		switch {
		case ctx.VariableDeclaration() != nil:
			LowerVariableDeclaration(ctx.VariableDeclaration())(yield)
		case ctx.Assignment() != nil:
			stmt := LowerAssignment(ctx.Assignment())
			yield(&stmt)
		case ctx.ClosureExpression() != nil:
			yield(&ast.ExpressionStatement{
				Expression: LowerClosureExpression(ctx.ClosureExpression()),
			})
		case ctx.Expression() != nil:
			yield(&ast.ExpressionStatement{
				Expression: LowerExpression(ctx.Expression()),
			})
		case ctx.BranchStatement() != nil:
			stmt := LowerBranchStatement(ctx.BranchStatement())
			yield(&stmt)
		case ctx.IsExpression() != nil:
			panic("todo")
		default:
			panic("unreachable")
		}
	}
}

func LowerStatements(statements []IStatementContext) (output []ast.Statement) {
	return slices.Collect(func(yield func(ast.Statement) bool) {
		for _, stmt := range statements {
			LowerStatement(stmt)(yield)
		}
	})
}

func LowerStatementBody(ctx IStatementBodyContext) ast.Block {
	return ast.Block{
		Span:       GetSpan(ctx),
		Statements: LowerStatements(ctx.AllStatement()),
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
	case ctx.GetOp() != nil:
		expr := ast.UnaryExpression{
			Span:       GetSpan(ctx),
			Op:         ast.UnaryOp(ctx.GetOp().GetText()),
			Expression: LowerPrimaryExpression(ctx.PrimaryExpression()),
		}
		return &expr
	case ctx.PrimaryExpression() != nil:
		return LowerPrimaryExpression(ctx.PrimaryExpression())
	default:
		panic("unreachable")
	}

}

func LowerVariableDeclaration(ctx IVariableDeclarationContext) iter.Seq[ast.Statement] {
	return func(yield func(ast.Statement) bool) {
		target := ctx.Target()

		// Regular variable declaration
		if len(target.AllName()) == 1 {
			yield(&ast.VariableDeclaration{
				Span: GetSpan(ctx),
				Name: LowerName(target.Name(0)),
				Type: LowerType(target.Type_(0)),
				Body: LowerStatementBody(ctx.StatementBody()),
			})
			return
		}
		// Tuple pattern matching that needs to be desugared
		// Input:
		// (x: Int, y: String) = body
		// Becomes:
		// tuple_3af823129b_: (Int, String) = body
		// x: Int = getTupleElement(tuple_3af823129b_, 0)
		// y: String = getTupleElement(tuple_3af823129b_, 1)
		tupleSpan := GetSpan(ctx)
		tupleName := fmt.Sprintf("tuple_%x_", rand.Uint64()) // TODO: do not use a random number
		elementTypes := util.Map(target.AllType_(), LowerType)
		tupleType := ast.Type{
			Expression: &ast.Tuple{
				Span: tupleSpan,
				Elements: util.Map(elementTypes, func(t ast.Type) ast.Expression {
					return t.Expression
				}),
			},
		}
		if !yield(&ast.VariableDeclaration{
			Span: tupleSpan,
			Name: ast.Name{
				Span:   tupleSpan,
				String: tupleName,
			},
			Type: tupleType,
			Body: LowerStatementBody(ctx.StatementBody()),
		}) {
			return
		}
		for i, name := range target.AllName() {
			_type := elementTypes[i]
			if !yield(&ast.VariableDeclaration{
				Span: GetSpan(ctx),
				Name: LowerName(name),
				Type: _type,
				Body: LowerStatementBody(ctx.StatementBody()),
			}) {
				return
			}
		}
	}
}

func LowerClosure(ctx IClosureContext) *ast.Closure {
	return &ast.Closure{
		Span:       GetSpan(ctx),
		Parameters: util.Map(ctx.ClosureParameters().AllFunctionParameter(), LowerFunctionParameter),
		ReturnType: LowerType(ctx.Type_()),
		Body:       LowerStatementBody(ctx.StatementBody()),
	}
}
