package ast

import (
	"errors"
	"fmt"
)

type Resolvable interface {
	Resolve(r *resolver)
}

type ResolveError struct {
	Message string
}

func (e ResolveError) Error() string {
	return fmt.Sprintf("resolver error: %s\n", e.Message)
}

type resolver struct {
	scope          []map[string]variable
	withinFunction bool
	withinLoop     bool
	errOccurred    bool
	report         func(error)
}

type variable struct {
	initialized bool
	captured    bool
}

func newVariable() variable {
	return variable{false, false}
}

func (v variable) initialize() variable {
	return variable{true, v.captured}
}

func (v variable) capture() variable {
	return variable{v.initialized, true}
}

func newResolver(report func(error)) *resolver {
	return &resolver{report: report,
		withinFunction: false,
		withinLoop:     false,
		errOccurred:    false,
		scope:          make([]map[string]variable, 0)}
}

func Resolve(stmts []Stmt, report func(error)) error {
	resolver := newResolver(report)

	resolver.BeginScope()
	for _, stmt := range stmts {
		stmt.Resolve(resolver)
	}

	resolver.EndScope()
	if resolver.errOccurred {
		return errors.New("resolver error")
	}

	return nil
}

func (r *resolver) ScopePeek() (map[string]variable, bool) {
	if len(r.scope) == 0 {
		return nil, false
	}
	return r.scope[len(r.scope)-1], true
}

func (r *resolver) BeginScope() {
	r.scope = append(r.scope, make(map[string]variable))
}

func (r *resolver) EndScope() {
	if len(r.scope) == 0 {
		return
	}

	for name, variable := range r.scope[len(r.scope)-1] {
		if !variable.captured {
			str := fmt.Sprintf("variable '%s' declared but never used", name)
			r.report(ResolveError{Message: str})
			r.errOccurred = true
		}
	}

	r.scope = r.scope[:len(r.scope)-1]
}

func (r *resolver) Declare(name string) {
	if scope, ok := r.ScopePeek(); ok {
		if _, ok := scope[name]; ok {
			str := fmt.Sprintf("variable '%s' already declared in this scope", name)
			r.report(ResolveError{Message: str})
			r.errOccurred = true
		}
		scope[name] = newVariable()
	}
}

func (r *resolver) Define(name string) {
	if scope, ok := r.ScopePeek(); ok {
		if _, ok := scope[name]; ok {
			scope[name] = scope[name].initialize()
		}
	}
}

// Expressions
func (e BinaryExpr) Resolve(r *resolver) {
	e.Left.Resolve(r)
	e.Right.Resolve(r)
}

func (e GroupingExpr) Resolve(r *resolver) {
	e.Expr.Resolve(r)
}

func (e LiteralExpr) Resolve(r *resolver) {}

func (e VariableExpr) Resolve(r *resolver) {
	if scope, ok := r.ScopePeek(); ok {
		if variable, ok := scope[e.Name.Lexme]; ok && !variable.initialized {
			r.report(ResolveError{Message: "variable used before initialization"})
			r.errOccurred = true
		}
	}

	for i := len(r.scope) - 1; i >= 0; i-- {
		name := e.Name.Lexme
		if variable, contains := r.scope[i][name]; contains {
			r.scope[i][name] = variable.capture()
			return
		}
	}
}

func (e UnaryExpr) Resolve(r *resolver) {
	e.Right.Resolve(r)
}

func (e TernaryExpr) Resolve(r *resolver) {
	e.Condition.Resolve(r)
	e.Left.Resolve(r)
	e.Right.Resolve(r)
}

func (e AssignExpr) Resolve(r *resolver) {
	e.Value.Resolve(r)
}

func (e FunctionExpr) Resolve(r *resolver) {
	enclosingFunction := r.withinFunction
	r.withinFunction = true
	r.BeginScope()

	for _, param := range e.Parameters {
		r.Declare(param.Lexme)
		r.Define(param.Lexme)
	}

	for _, stmt := range e.Body {
		stmt.Resolve(r)
	}

	r.EndScope()
	r.withinFunction = enclosingFunction
}

// Statements
func (s BlockStmt) Resolve(r *resolver) {
	r.BeginScope()
	defer r.EndScope()
	for _, stmt := range s.Statements {
		stmt.Resolve(r)
	}
}

func (s VarStmt) Resolve(r *resolver) {
	// split variable declaration and initialization into two separate steps
	// to prevent issues as: var a = a; (most sane to throw compile error here)
	r.Declare(s.Name.Lexme)

	if s.Initializer != nil {
		s.Initializer.Resolve(r)
	}

	r.Define(s.Name.Lexme)
}

func (s IfStmt) Resolve(r *resolver) {
	s.Condition.Resolve(r)
	s.ThenBranch.Resolve(r)
	if s.ElseBranch != nil {
		s.ElseBranch.Resolve(r)
	}
}

func (s WhileStmt) Resolve(r *resolver) {
	enclosingLoop := r.withinLoop
	r.withinLoop = true
	s.Condition.Resolve(r)
	s.Body.Resolve(r)
	r.withinLoop = enclosingLoop
}

func (s ReturnStmt) Resolve(r *resolver) {
	if !r.withinFunction {
		str := fmt.Sprintf("return statement outside of function")
		r.report(ResolveError{Message: str})
		r.errOccurred = true
	}
	if s.Expr != nil {
		s.Expr.Resolve(r)
	}
}

func (s FunctionStmt) Resolve(r *resolver) {
	r.Declare(s.Name.Lexme)
	r.Define(s.Name.Lexme)

	enclosingFunction := r.withinFunction
	r.withinFunction = true
	r.BeginScope()

	for _, param := range s.Parameters {
		r.Declare(param.Lexme)
		r.Define(param.Lexme)
	}

	for _, stmt := range s.Body {
		stmt.Resolve(r)
	}

	r.EndScope()
	r.withinFunction = enclosingFunction
}

func (s ExpressionStmt) Resolve(r *resolver) {
	s.Expr.Resolve(r)
}

func (s PrintStmt) Resolve(r *resolver) {
	s.Expr.Resolve(r)
}

func (s BreakStmt) Resolve(r *resolver) {
	if !r.withinLoop {
		r.report(ResolveError{Message: "break statement outside of loop"})
		r.errOccurred = true
	}
}

func (s CallStmt) Resolve(r *resolver) {
	s.Callee.Resolve(r)
	for _, arg := range s.Arguments {
		arg.Resolve(r)
	}
}

func (s NothingExpr) Resolve(r *resolver) {}
