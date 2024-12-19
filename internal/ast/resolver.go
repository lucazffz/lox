package ast

import (
	"errors"
)

type Resolvable interface {
	Resolve(r *Resolver) error
}

type ResolvableStmt interface {
	Resolvable
}

type ResolvableExpr interface {
	Resolvable
	Depth() int
	SetDepth(int)
}

type Resolver struct {
	stmts []Resolvable
	scope []map[string]bool
}

func NewResolver(stmts []Resolvable) *Resolver {
	return &Resolver{stmts: stmts}
}

func (r *Resolver) Resolve(resolvable Resolvable) error {
	return resolvable.Resolve(r)
}

func (s *BlockStmt) Resolve(r *Resolver) error {
	r.BeginScope()
	defer r.EndScope()

	for _, stmt := range s.Statements {
		stmt.Resolve(r)
	}

	return nil
}

func (s *VarStmt) Resolve(r *Resolver) error {
	// split variable declaration and initialization into two separate steps
	// to prevent issues as: var a = a; (most sane to throw compile error here)
	r.Declare(s.Name.Lexme)

	if s.Initializer != nil {
		s.Initializer.Resolve(r)
	}

	r.Define(s.Name.Lexme)

	return nil
}

func (e *VariableExpr) Resolve(r *Resolver) error {
	if value, ok := r.ScopePeek()[e.Name.Lexme]; !ok {
		return errors.New("variable not found")
	} else if !value {
		return errors.New("variable is not defined")
	}

	r.resolveLocal(e, e.Name.Lexme)
	return nil
}

func (e *VariableExpr) Depth() int {
	return e.depth
}

func (e *VariableExpr) SetDepth(depth int) {
	e.depth = depth
}

func (r *Resolver) resolveLocal(expr ResolvableExpr, name string) {
	// pop from scope stack until we find the variable
	for i := len(r.scope) - 1; i >= 0; i-- {
		if _, contains := r.scope[i][name]; contains {
			scopeDepth := len(r.scope) - 1 - i
			expr.SetDepth(scopeDepth)
			return
		}
	}
}

func (r *Resolver) ScopePeek() map[string]bool {
	return r.scope[len(r.scope)-1]
}

func (r *Resolver) BeginScope() {
	r.scope = append(r.scope, make(map[string]bool))
}

func (r *Resolver) EndScope() {
	r.scope = r.scope[:len(r.scope)-1]
}

func (r *Resolver) Declare(name string) {
	if len(r.scope) == 0 {
		return
	}

	scope := r.ScopePeek()
	scope[name] = false
}

func (r *Resolver) Define(name string) {
	if len(r.scope) == 0 {
		return
	}

	scope := r.ScopePeek()
	scope[name] = true
}
