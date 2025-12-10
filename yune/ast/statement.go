package ast

import "yune/cpp"

type VariableDeclaration struct {
	Span
	Name Name
	Type Type
	Body Block
}

// Lower implements Statement.
func (d *VariableDeclaration) Lower() cpp.Statement {
	return cpp.VariableDeclaration{
		Name: d.Name.String,
		Type: d.Type.Lower(),
		Body: nil, // TODO
	}
}

// Analyze implements Statement.
func (d *VariableDeclaration) Analyze() (queries Queries, finalizer Finalizer) {
	queries, typeFin := d.Type.Analyze()
	_queries, bodyFin := d.Body.Analyze()
	queries = append(queries, _queries...)

	// TODO: check for duplicate definition and add definition to environment

	finalizer = func(env Env) (errors Errors) {
		errors = append(typeFin(env), bodyFin(env)...)
		if len(errors) > 0 {
			return
		}
		if !d.Type.InferredType.Eq(d.Body.InferredType) {
			errors = append(errors, TypeMismatch{
				Expected: d.Type.InferredType,
				Found:    d.Body.GetType(),
				At:       d.Body.Statements[len(d.Body.Statements)-1].GetSpan(),
			})
		}
		return
	}
	return
}

func (d VariableDeclaration) GetName() string {
	return d.Name.String
}

func (d VariableDeclaration) GetType() InferredType {
	return d.Type.InferredType
}

type Assignment struct {
	Span
	Target Variable
	Op     AssignmentOp
	Body   Block
}

// Lower implements Statement.
func (a *Assignment) Lower() cpp.Statement {
	return cpp.Assignment{
		Target: a.Target.GetName(),
		Op:     cpp.AssignmentOp(a.Op),
		Value:  nil, // TODO
	}
}

// Analyze implements Statement.
func (a *Assignment) Analyze() (queries Queries, finalizer Finalizer) {
	queries, targetFin := a.Target.Analyze()
	_queries, bodyFin := a.Body.Analyze()
	queries = append(queries, _queries...)

	finalizer = func(env Env) (errors Errors) {
		errors = append(targetFin(env), bodyFin(env)...)
		if !a.Target.GetType().Eq(a.Body.GetType()) {
			errors = append(errors, TypeMismatch{
				Expected: a.Target.GetType(),
				Found:    a.Body.GetType(),
				At:       a.Body.Statements[len(a.Body.Statements)-1].GetSpan(),
			})
		}
		return
	}
	return
}

func (a Assignment) GetType() InferredType {
	return a.Target.GetType()
}

type AssignmentOp string

// Always the last statement in a list, since the remaining
// statements in a block are is in its .Else field.
type BranchStatement struct {
	Span
	InferredType
	Condition Expression
	Then      Block
	Else      Block
}

// Lower implements Statement.
func (b *BranchStatement) Lower() cpp.Statement {
	return cpp.BranchStatement{
		Condition: b.Condition.Lower(),
		Then:      b.Then.Lower(),
		Else:      b.Else.Lower(),
	}
}

// Analyze implements Statement.
func (b *BranchStatement) Analyze() (queries Queries, finalizer Finalizer) {
	queries, conditionFin := b.Condition.Analyze()
	_queries, thenFin := b.Then.Analyze()
	queries = append(queries, _queries...)
	_queries, elseFin := b.Else.Analyze()
	queries = append(queries, _queries...)

	finalizer = func(env Env) (errors Errors) {
		errors = append(conditionFin(env), thenFin(env)...)
		errors = append(errors, elseFin(env)...)
		if len(errors) > 0 {
			return
		}
		if !b.Then.GetType().Eq(b.Else.GetType()) {
			errors = append(errors, BranchTypeNotEqual{
				Then:   b.Then.GetType(),
				ThenAt: b.Then.Statements[len(b.Then.Statements)-1].GetSpan(),
				Else:   b.Else.GetType(),
				ElseAt: b.Else.Statements[len(b.Else.Statements)-1].GetSpan(),
			})
			return
		}
		b.InferredType = b.Then.GetType()
		return
	}
	return
}

type Block struct {
	InferredType
	Statements []Statement
}

func (b *Block) Analyze() (queries Queries, finalizer Finalizer) {
	finalizers := make([]Finalizer, len(b.Statements))
	for i := range b.Statements {
		_queries, _finalizer := b.Statements[i].Analyze()
		queries = append(queries, _queries...)
		finalizers = append(finalizers, _finalizer)
	}
	finalizer = func(env Env) (errors Errors) {
		for i, stmt := range b.Statements {
			errors = append(errors, finalizers[i](env)...)
			if len(errors) > 0 {
				return
			}
			decl, ok := stmt.(Declaration)
			if ok {
				env.declarations[decl.GetName()] = decl
			}
		}
		return
	}
	return
}

func (b Block) Lower() cpp.Block {
	// TODO: function body vs variable body
	panic("unimplemented")
}

type ExpressionStatement struct {
	Expression
}

// Lower implements Statement.
func (e *ExpressionStatement) Lower() cpp.Statement {
	return e.Expression.Lower()
}

type Statement interface {
	INode
	Analyze() (queries Queries, finalizer Finalizer)
	GetType() InferredType
	Lower() cpp.Statement
}

var _ Statement = &VariableDeclaration{}
var _ Statement = &Assignment{}
var _ Statement = &BranchStatement{}
