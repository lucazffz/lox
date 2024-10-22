package ast

import (
    "errors"
)

type Environment struct {
	enviornment map[string]Value
}

func NewEnvironment() *Environment {
	return &Environment{enviornment: make(map[string]Value)}
}

func (e *Environment) Define(name string, value Value) {
	e.enviornment[name] = value
}

func (e *Environment) Get(name string) (Value, error) {
	value, ok := e.enviornment[name]
	if !ok {
		return nil, errors.New("")
	}

	return value, nil
}
