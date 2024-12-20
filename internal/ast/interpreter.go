package ast

import (
	"errors"
	"time"
)

// the global environment
var global_env = NewEnvironment(nil)

// the current scope depth we operate in (global is 0). Does not
// correspond to the environment since a new environment is created
// for each new declaration aswell, not just for each new scope

type ExprHashable string

var Locals = make(map[ExprHashable]int)

// the current environment (used for block scopes) we
// operate in, starts as the global environment but may be
// reassigned by block scopes
var current_env = global_env

var clockFunc = NativeFunction{
	paramLen: 0,
	Function: func(_ []LoxValue) (LoxValue, error) {
		return LoxNumber(float64(time.Now().UnixNano()) / 1e9), nil
	},
}

var typeFunc = NativeFunction{
	paramLen: 1,
	Function: func(args []LoxValue) (LoxValue, error) {
		return AsType(args[0]), nil
	},
}

func addNativeFunction(name string, f NativeFunction) {
	global_env.enviornment[name] = f
}

func executeBlock(statements []Stmt, env *Environment) error {
	previous := current_env
	current_env = env
	defer func() { current_env = previous }()

	for _, stmt := range statements {
		if err := stmt.Evaluate(); err != nil {
			return err
		}
	}

	return nil
}

func Interpret(statements []Stmt, report func(error)) error {
	addNativeFunction("type", typeFunc)
	addNativeFunction("clock", clockFunc)
	// global_env.Define("str", LoxType{Typ: STRING})
	// global_env.Define("num", LoxType{Typ: NUMBER})
	// global_env.Define("func", LoxType{Typ: FUNCTION})
	// global_env.Define("bool", LoxType{Typ: BOOLEAN})

	var errorHasOccured = false
	for _, stmt := range statements {
		if err := stmt.Evaluate(); err != nil {
			report(err)
			errorHasOccured = true
		}
	}

	if errorHasOccured {
		return errors.New("")
	}

	return nil
}
