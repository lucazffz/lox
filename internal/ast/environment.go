package ast

import (
	"errors"
	"github.com/LucazFFz/lox/internal/token"
)

type Environment struct {
	enclosing   *Environment
	enviornment map[string]Value
}

func NewEnvironment(enclosing *Environment) *Environment {
	return &Environment{
		enviornment: make(map[string]Value),
		enclosing:   enclosing,
	}
}

func (e *Environment) Define(name token.Token, value Value) {
	e.enviornment[name.Lexme] = value
}

func (e *Environment) Assign(name token.Token, value Value) error {
	_, ok := e.enviornment[name.Lexme]
	if ok {
		e.enviornment[name.Lexme] = value
		return nil
	}

	if e.enclosing != nil {
		return e.enclosing.Assign(name, value)
	}

	return errors.New("")
}

func (e *Environment) Get(name token.Token) (Value, error) {
	// try to get variable for this scope
	if value, ok := e.enviornment[name.Lexme]; ok {
		return value, nil
	}

	if e.enclosing != nil {
		// try to get variable for enclosing scope recursively
		if value, err := e.enclosing.Get(name); err == nil {
			return value, nil
		}
	}

	return nil, errors.New("")
}
