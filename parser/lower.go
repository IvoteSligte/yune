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

// Input:
// expression is name: type
// Becomes:
// is_expr_8fa932bc_ := expression
// isVariant(type)(is_expr_8fa932bc_) ->
//
//	name := getVariant(type)(is_expr_8fa932bc_)
//	...
func DesugarIsExpression(ctx IIsExpressionContext, thenBlock ast.Block, elseBlock ast.Block) iter.Seq[ast.Statement] {
	return func(yield func(ast.Statement) bool) {
		span := GetSpan(ctx)
		temporary := ast.Name{
			String: fmt.Sprintf("is_expr_%x_", rand.Uint64()),
			Span:   span,
		}
		if !yield(&ast.VariableDeclaration{
			Span:      span,
			Name:      temporary,
			InferType: true,
			Body: ast.Block{
				Span: GetSpan(ctx.Expression()),
				Statements: []ast.Statement{
					&ast.ExpressionStatement{Expression: LowerExpression(ctx.Expression())},
				},
			},
		}) {
			return
		}
		// isVariant(type)(is_expr_xxxx_)
		condition := &ast.FunctionCall{
			Span: span,
			Function: &ast.FunctionCall{
				Span: span,
				Function: &ast.Variable{
					Name: ast.Name{
						Span:   GetSpan(ctx.Type_()),
						String: "isVariant",
					},
				},
				// NOTE: the type is currently evaluated twice, first here
				Argument: LowerExpression(ctx.Type_().Expression()),
			},
			Argument: &ast.Variable{Name: temporary},
		}
		// name := getVariant(type)(is_expr_xxxx_)
		thenBlock.Statements = append([]ast.Statement{&ast.VariableDeclaration{
			Span: GetSpan(ctx.Name()),
			Name: ast.Name{
				Span:   GetSpan(ctx.Name()),
				String: ctx.Name().GetText(),
			},
			InferType: true,
			// NOTE: the type is currently evaluated twice, second here
			Type: LowerType(ctx.Type_()),
			Body: ast.Block{
				Span: span,
				Statements: []ast.Statement{
					&ast.ExpressionStatement{Expression: &ast.FunctionCall{
						Span: span,
						Function: &ast.FunctionCall{
							Span: span,
							Function: &ast.Variable{
								Name: ast.Name{
									Span:   GetSpan(ctx.Type_()),
									String: "getVariant",
								},
							},
							// NOTE: the type is currently evaluated twice, first here
							Argument: LowerExpression(ctx.Type_().Expression()),
						},
						Argument: &ast.Variable{Name: temporary},
					}},
				},
			},
		}}, thenBlock.Statements...)
		yield(&ast.BranchStatement{
			Span:      span,
			Condition: condition,
			Then:      thenBlock,
			Else:      elseBlock,
		})
	}
}

func LowerBranchStatement(ctx IBranchStatementContext) iter.Seq[ast.Statement] {
	return func(yield func(ast.Statement) bool) {
		switch {
		case ctx.Expression() != nil:
			yield(&ast.BranchStatement{
				Span:      GetSpan(ctx),
				Condition: LowerExpression(ctx.Expression()),
				Then:      LowerStatementBody(ctx.StatementBody()),
				Else: ast.Block{
					Span:       GetSpan(ctx.Block()),
					Statements: LowerStatements(ctx.Block().AllStatement()),
				},
			})
		case ctx.IsExpression() != nil:
			elseBlock := ast.Block{
				Span:       GetSpan(ctx.Block()),
				Statements: LowerStatements(ctx.Block().AllStatement()),
			}
			DesugarIsExpression(ctx.IsExpression(), LowerStatementBody(ctx.StatementBody()), elseBlock)(yield)
		default:
			panic("unreachable")
		}
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
			LowerBranchStatement(ctx.BranchStatement())(yield)
		case ctx.IsStatement() != nil:
			LowerIsStatement(ctx.IsStatement())
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

func LowerBlock(ctx IBlockContext) ast.Block {
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

func LowerIsStatement(ctx IIsStatementContext) iter.Seq[ast.Statement] {
	return func(yield func(ast.Statement) bool) {
		span := GetSpan(ctx)
		thenBlock := LowerBlock(ctx.Block())
		elseBlock := ast.Block{
			Span: span,
			Statements: []ast.Statement{
				&ast.ExpressionStatement{Expression: &ast.FunctionCall{
					Span: span,
					Function: &ast.Variable{
						Name: ast.Name{
							Span:   span,
							String: "panic",
						},
					},
					Argument: &ast.String{
						Span:  span,
						Value: fmt.Sprintf("is-statement assertion at %s returned false", span),
					},
				}},
			},
		}
		DesugarIsExpression(ctx.IsExpression(), thenBlock, elseBlock)(yield)
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
		// x := getTupleElement(tuple_3af823129b_, 0)
		// y := getTupleElement(tuple_3af823129b_, 1)
		tupleSpan := GetSpan(ctx)
		tupleName := fmt.Sprintf("tuple_%x_", rand.Uint64()) // TODO: do not use a random number
		tupleType := ast.Type{
			Expression: &ast.Tuple{
				Span: tupleSpan,
				Elements: util.Map(target.AllType_(), func(t ITypeContext) ast.Expression {
					return LowerExpression(t.Expression())
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
		for _, name := range target.AllName() {
			if !yield(&ast.VariableDeclaration{
				Span:      GetSpan(ctx),
				Name:      LowerName(name),
				InferType: true,
				Body:      LowerStatementBody(ctx.StatementBody()),
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
