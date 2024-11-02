package ast

import (
	"errors"
	"time"
)

var global_env = NewEnvironment(nil)

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
	global_env.Define(name, f)
}

func Interpret(statements []Stmt, report func(error)) error {
	addNativeFunction("type", typeFunc)
	addNativeFunction("clock", clockFunc)
	global_env.Define("str", LoxType{Typ: STRING})
	global_env.Define("num", LoxType{Typ: NUMBER})
	global_env.Define("func", LoxType{Typ: FUNCTION})
	global_env.Define("bool", LoxType{Typ: BOOLEAN})

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
