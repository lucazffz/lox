package ast

import (
	"errors"
	"github.com/LucazFFz/lox/internal/token"
)

// Enironments are presistent meaning their values cannot change. Each
// variable declaration will create a new environment. However, the
// variables in a given environment can be reassigned.
// Neccessary for closures to work properly.
type Environment struct {
	enclosing   *Environment
	enviornment map[string]LoxValue
}

func NewEnvironment(enclosing *Environment) *Environment {
	return &Environment{
		enviornment: make(map[string]LoxValue),
		enclosing:   enclosing,
	}
}

func NewEnvironmentWithDefined(env *Environment, m map[string]LoxValue) *Environment {
	newEnv := NewEnvironment(env)
	newEnv.enviornment = m
	return newEnv
}

func Define(env *Environment, name string, value LoxValue) *Environment {
	newEnv := NewEnvironment(env)
	newEnv.enviornment[name] = value
	return newEnv
}

func (e *Environment) Assign(name string, value LoxValue) error {
	_, ok := e.enviornment[name]
	if ok {
		e.enviornment[name] = value
		return nil
	}

	if e.enclosing != nil {
		return e.enclosing.Assign(name, value)
	}

	return errors.New("")
}

func (e *Environment) Get(name token.Token) (LoxValue, error) {
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
