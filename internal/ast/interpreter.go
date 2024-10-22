package ast

import (
	"errors"
)

var environment = NewEnvironment()

func Interpret(statements []Stmt, report func(error)) error {
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
