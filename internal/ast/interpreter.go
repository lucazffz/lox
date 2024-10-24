package ast

import (
	"errors"
)

var environment = NewEnvironment(nil)

func Interpret(statements []Stmt, report func(error)) error {
	var errorHasOccured = false
	for _, stmt := range statements {
        // print("eval")
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
