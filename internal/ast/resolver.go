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
	return fmt.Sprintf("resolver error: %s", e.Message)
}

type resolver struct {
	scope          []map[string]bool
	withinFunction bool
	withinLoop     bool
	errOccurred    bool
	report         func(error)
}

func newResolver(report func(error)) *resolver {
	return &resolver{report: report,
		withinFunction: false,
		withinLoop:     false,
		errOccurred:    false,
		scope:          make([]map[string]bool, 0)}
}

func Resolve(stmts []Stmt, report func(error)) error {
	resolver := newResolver(report)
	resolver.BeginScope()
	defer resolver.EndScope()
	for _, stmt := range stmts {
		stmt.Resolve(resolver)
	}
	if resolver.errOccurred {
		return errors.New("resolver error")
	}

	return nil
}

// tells the interpreter the number of scopes between the
// current scope and the scope where the variable was declared
// used for variable expressions and assignment expressions
// resolution information stored in 'resolution' map
func (r *resolver) resolveLocal(expr Expr, name string) {
	// pop from scope stack until we find the variable
	for i := len(r.scope) - 1; i >= 0; i-- {
		if _, contains := r.scope[i][name]; contains {
			declDist := len(r.scope) - 1 - i
			hashable := ExprHashable(fmt.Sprintf("%v", expr))
			Locals[hashable] = declDist
			return
		}
	}
}

func (r *resolver) resolveFunction(resolvable Resolvable) {
	enclosingFunction := r.withinFunction
	r.withinFunction = true
	r.BeginScope()
	defer func() {
		r.EndScope()
		r.withinFunction = enclosingFunction
	}()

	switch function := resolvable.(type) {
	case FunctionStmt:
		for _, param := range function.Parameters {
			r.Declare(param.Lexme)
			r.Define(param.Lexme)
		}

		for _, stmt := range function.Body {
			stmt.Resolve(r)
		}
		break
	case FunctionExpr:
		for _, param := range function.Parameters {
			r.Declare(param.Lexme)
			r.Define(param.Lexme)
		}

		for _, stmt := range function.Body {
			stmt.Resolve(r)
		}
		break
	default:
		panic("invalid function type")
	}
}

func (r *resolver) ScopePeek() (map[string]bool, bool) {
	if len(r.scope) == 0 {
		return nil, false
	}
	return r.scope[len(r.scope)-1], true
}

func (r *resolver) BeginScope() {
	r.scope = append(r.scope, make(map[string]bool))
}

func (r *resolver) EndScope() {
	if len(r.scope) == 0 {
		return
	}
	r.scope = r.scope[:len(r.scope)-1]
}

func (r *resolver) Declare(name string) {
	if scope, ok := r.ScopePeek(); ok {
		if _, ok := scope[name]; ok {
			r.report(ResolveError{Message: "variable already declared in this scope"})
			r.errOccurred = true
		}
		scope[name] = false
	}
}

func (r *resolver) Define(name string) {
	if scope, ok := r.ScopePeek(); ok {
		scope[name] = true
	}
}

func (r *resolver) scopeDepth() int {
	return len(r.scope)
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
		if initialized, ok := scope[e.Name.Lexme]; ok && !initialized {
			r.report(ResolveError{Message: "variable used before initialization"})
			r.errOccurred = true
		}
	}

	r.resolveLocal(e, e.Name.Lexme)
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
	r.resolveLocal(e, e.Name.Lexme)
}

func (e FunctionExpr) Resolve(r *resolver) {
	r.resolveFunction(e)
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
		r.report(ResolveError{Message: "return statement outside of function"})
		r.errOccurred = true
	}
	if s.Expr != nil {
		s.Expr.Resolve(r)
	}
}

func (s FunctionStmt) Resolve(r *resolver) {
	r.Declare(s.Name.Lexme)
	r.Define(s.Name.Lexme)
	r.resolveFunction(s)
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
