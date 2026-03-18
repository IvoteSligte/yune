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

// TODO: thread-safety?
var FileName string
var SourceCode string

func GetSpan(ctx antlr.ParserRuleContext) ast.Span {
	return ast.Span{
		File:    FileName,
		Source:  SourceCode,
		Line:    ctx.GetStart().GetLine(),
		Column:  ctx.GetStart().GetColumn(),
		Length:  len(ctx.GetText()),
		Content: ctx.GetText(),
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

func LowerIsExpression(ctx IIsExpressionContext, thenBlock ast.Block, elseBlock ast.Block) ast.Statement {
	return &ast.IsBranchStatement{
		Span:       GetSpan(ctx),
		Expression: LowerExpression(ctx.Expression()),
		Name:       LowerName(ctx.Name()),
		Type:       LowerType(ctx.Type_()),
		Then:       thenBlock,
		Else:       elseBlock,
	}
}

func LowerBranchStatement(ctx IBranchStatementContext) ast.Statement {
	switch {
	case ctx.Expression() != nil:
		return &ast.BranchStatement{
			Span:      GetSpan(ctx),
			Condition: LowerExpression(ctx.Expression()),
			Then:      LowerStatementBody(ctx.StatementBody()),
			Else: ast.Block{
				Span:       GetSpan(ctx.Block()),
				Statements: LowerStatements(ctx.Block().AllStatement()),
			},
		}
	case ctx.IsExpression() != nil:
		thenBlock := LowerStatementBody(ctx.StatementBody())
		elseBlock := ast.Block{
			Span:       GetSpan(ctx.Block()),
			Statements: LowerStatements(ctx.Block().AllStatement()),
		}
		if len(elseBlock.Statements) == 0 {
			panic(fmt.Sprintf("Empty else-block of is-expression at %s", elseBlock.Span))
		}
		return LowerIsExpression(ctx.IsExpression(), thenBlock, elseBlock)
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
			Span:     GetSpan(ctx),
			Function: LowerPrimaryExpression(ctx.GetFunction()),
			Argument: LowerPrimaryExpression(ctx.GetArgument()),
		}
	case ctx.Variable() != nil:
		variable := LowerVariable(ctx.Variable())
		return &variable
	case ctx.List() != nil:
		list := LowerList(ctx.List())
		return &list
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
			yield(LowerBranchStatement(ctx.BranchStatement()))
		case ctx.IsStatement() != nil:
			yield(LowerIsStatement(ctx.IsStatement()))
		default:
			panic("unreachable")
		}
	}
}

func LowerStatements(statements []IStatementContext) []ast.Statement {
	if len(statements) == 0 {
		// Empty block maps to empty tuple ()
		return []ast.Statement{
			&ast.ExpressionStatement{Expression: &ast.Tuple{}},
		}
	}
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

func LowerList(ctx IListContext) ast.List {
	return ast.List{
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

func LowerIsStatement(ctx IIsStatementContext) ast.Statement {
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
	return LowerIsExpression(ctx.IsExpression(), thenBlock, elseBlock)
}

func LowerVariableDeclaration(ctx IVariableDeclarationContext) iter.Seq[ast.Statement] {
	return func(yield func(ast.Statement) bool) {
		if ctx.Name() != nil {
			// variable declaration with inferred type
			yield(&ast.VariableDeclaration{
				Span:      GetSpan(ctx),
				Name:      LowerName(ctx.Name()),
				InferType: true,
				Body:      LowerStatementBody(ctx.StatementBody()),
			})
			return
		}
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
		// x := getTupleElement_(tuple_3af823129b_, 0)
		// y := getTupleElement_(tuple_3af823129b_, 1)
		tupleSpan := GetSpan(ctx)
		tupleName := ast.Name{
			Span:   tupleSpan,
			String: fmt.Sprintf("tuple_%x_", rand.Uint64()), // TODO: do not use a random number
		}
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
			Name: tupleName,
			Type: tupleType,
			Body: LowerStatementBody(ctx.StatementBody()),
		}) {
			return
		}
		for i, name := range target.AllName() {
			span := GetSpan(name)
			if !yield(&ast.VariableDeclaration{
				Span:      span,
				Name:      LowerName(name),
				InferType: true,
				Body: ast.Block{
					Span: tupleSpan,
					Statements: []ast.Statement{
						&ast.ExpressionStatement{
							Expression: &ast.FunctionCall{
								Span: span,
								Function: &ast.Variable{
									Name: ast.Name{
										Span:   span,
										String: "getTupleElement_",
									},
								},
								Argument: &ast.Tuple{Elements: []ast.Expression{
									&ast.Variable{Name: tupleName},
									&ast.Integer{
										Span:  span,
										Value: int64(i),
									},
								}},
							},
						},
					},
				},
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
