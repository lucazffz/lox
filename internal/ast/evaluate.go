package ast

import (
	"fmt"
	"github.com/LucazFFz/lox/internal/token"
)

type EvaluateExpr interface {
	Evaluate() (LoxValue, error)
}

type EvaluateStmt interface {
	Evaluate() error
}

// evaluating a break statement will return a BreakError
// (a type of runtime error) which we can catch when
// evaluating a while loop and break out of the loop
// unsure if this is the best way to handle this
type BreakError struct {
	RuntimeError
}

type ReturnError struct {
	RuntimeError
	Value LoxValue
}

type RuntimeError struct {
	message string
}

func NewRuntimeError(message string) RuntimeError {
	return RuntimeError{message: message}
}

func (r RuntimeError) Error() string {
	return "runtime error - " + r.message + "\n"
}

// statements
func (s ExpressionStmt) Evaluate() error {
	_, err := s.Expr.Evaluate()
	return err
}

func (s PrintStmt) Evaluate() error {
	value, err := s.Expr.Evaluate()
	if err != nil {
		return err
	}

	str, err := valueToString(value)
	if err != nil {
		return err
	}

	println(str)
	return nil
}

func (s BlockStmt) Evaluate() error {
	previous := global_env
	global_env = NewEnvironment(global_env)

	for _, statement := range s.Statements {
		err := statement.Evaluate()
		if err != nil {
			global_env = previous
			return err
		}
	}

	global_env = previous
	return nil
}

func (s VarStmt) Evaluate() error {
	if (s.Initializer == NothingExpr{}) {
		global_env.Define(s.Name.Lexme, LoxNil{})
	}

	value, err := s.Initializer.Evaluate()
	if err != nil {
		return err
	}

	global_env.Define(s.Name.Lexme, value)
	return nil
}

func (s IfStmt) Evaluate() error {
	value, err := s.Condition.Evaluate()
	if err != nil {
		return err
	}

	if isTruthy(value) {
		err := s.ThenBranch.Evaluate()
		if err != nil {
			return err
		}
	} else if s.ElseBranch != nil {
		err := s.ElseBranch.Evaluate()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s WhileStmt) Evaluate() error {
	value, err := s.Condition.Evaluate()
	if err != nil {
		return err
	}

	for isTruthy(value) {
		err := s.Body.Evaluate()
		if err != nil {
			// if we encounter a breakError,
			// we want to break out of the loop
			if _, ok := err.(BreakError); ok {
				return nil
			}

			return err
		}

		value, err = s.Condition.Evaluate()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s BreakStmt) Evaluate() error {
	return BreakError{NewRuntimeError("unexpected break statement")}
}

func (s ReturnStmt) Evaluate() error {
	var value LoxValue = LoxNil{}
	var err error
	if s.Expr != nil {
		value, err = s.Expr.Evaluate()
	}

	if err != nil {
		return err
	}

	return ReturnError{
		RuntimeError: NewRuntimeError("unexpected return statement"),
		Value:        value,
	}
}

func (t CallStmt) Evaluate() (LoxValue, error) {
	callee, err := t.Callee.Evaluate()
	if err != nil {
		return nil, err
	}

	arguments := []LoxValue{}
	for _, arg := range t.Arguments {
		arg, err := arg.Evaluate()
		if err != nil {
			return nil, err
		}

		arguments = append(arguments, arg)
	}

	if function, ok := callee.(Callable); ok {
		if len(arguments) != function.Arity() {
			return nil, NewRuntimeError(
				fmt.Sprintf("expected {%d} arguments but got {%d} arguments",
					len(arguments),
					function.Arity()))
		}

		value, err := function.Call(arguments)
		if err != nil {
			return nil, err
		}

		return value, nil
	}

	return nil, NewRuntimeError("can only invoke functions and methods")
}

func (t FunctionStmt) Evaluate() error {
	function := LoxFunction{FunctionStmt: t, Closure: global_env}
	global_env.Define(t.Name.Lexme, function)
	return nil
}

// expressions
func (t LiteralExpr) Evaluate() (LoxValue, error) {
	return t.Value, nil
}

func (t GroupingExpr) Evaluate() (LoxValue, error) {
	return t.Expr.Evaluate()
}

func (t UnaryExpr) Evaluate() (LoxValue, error) {
	right, err := t.Right.Evaluate()
	if err != nil {
		return nil, err
	}
	switch t.Op.Type {
	case token.BANG:
		return LoxBoolean(!isTruthy(right)), nil
	case token.MINUS:
		if !isNumber(right) {
			return nil, NewRuntimeError("operand must be a number")
		}
		return LoxNumber(-AsNumber(right)), nil

	}

	panic("should never reach here")
}

func (t BinaryExpr) Evaluate() (LoxValue, error) {
	checkNumberOperands := func(left, right LoxValue) error {
		if !isNumber(left) || !isNumber(right) {
			return NewRuntimeError("both operands must be numbers")
		}

		return nil
	}

	checkStringOperands := func(left, right LoxValue) error {
		if !isString(left) || !isString(right) {
			return NewRuntimeError("both operands must be strings")
		}

		return nil
	}

	evaluateOperands := func() (LoxValue, LoxValue, error) {
		left, err := t.Left.Evaluate()
		if err != nil {
			return nil, nil, err
		}
		right, err := t.Right.Evaluate()
		if err != nil {
			return nil, nil, err
		}
		return left, right, nil
	}

	switch t.Op.Type {
	case token.AND:
		fallthrough
	case token.OR:
		left, err := t.Left.Evaluate()
		if err != nil {
			return nil, err
		}

		if token.OR == t.Op.Type {
			if isTruthy(left) {
				return left, nil
			}
		} else {
			if !isTruthy(left) {
				return left, nil
			}
		}

		// if AND we know that left is true here, if OR we know
		// that left is false
		return t.Right.Evaluate()
	case token.PLUS:
		left, right, err := evaluateOperands()
		if err != nil {
			return nil, err
		}
		if err := checkNumberOperands(left, right); err == nil {
			return LoxNumber(AsNumber(left) + AsNumber(right)), nil
		}

		if err := checkStringOperands(left, right); err == nil {
			return LoxString(AsString(left) + AsString(right)), nil
		}

		return nil, NewRuntimeError("operands must be of same type")
	case token.MINUS:
		left, right, err := evaluateOperands()
		if err != nil {
			return nil, err
		}
		if err := checkNumberOperands(left, right); err != nil {
			return nil, err
		}
		return LoxNumber(AsNumber(left) - AsNumber(right)), nil
	case token.STAR:
		left, right, err := evaluateOperands()
		if err != nil {
			return nil, err
		}
		if err := checkNumberOperands(left, right); err != nil {
			return nil, err
		}
		return LoxNumber(AsNumber(left) * AsNumber(right)), nil
	case token.SLASH:
		left, right, err := evaluateOperands()
		if err != nil {
			return nil, err
		}
		if err := checkNumberOperands(left, right); err != nil {
			return nil, err
		}

		if AsNumber(right) == 0 {
			return nil, NewRuntimeError("division by zero")
		}

		return LoxNumber(AsNumber(left) / AsNumber(right)), nil
	case token.GREATER:
		left, right, err := evaluateOperands()
		if err != nil {
			return nil, err
		}
		if err := checkNumberOperands(left, right); err == nil {
			return LoxBoolean(AsNumber(left) > AsNumber(right)), nil
		}

		if err := checkStringOperands(left, right); err == nil {
			return LoxBoolean(AsString(left) > AsString(right)), nil
		}

		return nil, NewRuntimeError("operands must be of same type")
	case token.GREATER_EQUAL:
		left, right, err := evaluateOperands()
		if err != nil {
			return nil, err
		}
		if err := checkNumberOperands(left, right); err == nil {
			return LoxBoolean(AsNumber(left) >= AsNumber(right)), nil
		}

		if err := checkStringOperands(left, right); err == nil {
			return LoxBoolean(AsString(left) >= AsString(right)), nil
		}

		return nil, NewRuntimeError("operands must be of same type")
	case token.LESS:
		left, right, err := evaluateOperands()
		if err != nil {
			return nil, err
		}
		if err := checkNumberOperands(left, right); err == nil {
			return LoxBoolean(AsNumber(left) < AsNumber(right)), nil
		}

		if err := checkStringOperands(left, right); err == nil {
			return LoxBoolean(AsString(left) < AsString(right)), nil
		}

		return nil, NewRuntimeError("operands must be of same type")
	case token.LESS_EQUAL:
		left, right, err := evaluateOperands()
		if err != nil {
			return nil, err
		}
		if err := checkNumberOperands(left, right); err == nil {
			return LoxBoolean(AsNumber(left) <= AsNumber(right)), nil
		}

		if err := checkStringOperands(left, right); err == nil {
			return LoxBoolean(AsString(left) <= AsString(right)), nil
		}

		return nil, NewRuntimeError("operands must be of same type")
	case token.EQUAL_EQUAL:
		left, right, err := evaluateOperands()
		if err != nil {
			return nil, err
		}
		return LoxBoolean(equals(left, right)), nil
	case token.BANG_EQUAL:
		left, right, err := evaluateOperands()
		if err != nil {
			return nil, err
		}
		return LoxBoolean(!equals(left, right)), nil
	}

	panic("should never reach here (binary)")
}

func (t TernaryExpr) Evaluate() (LoxValue, error) {
	condition, err := t.Condition.Evaluate()
	if err != nil {
		return nil, err
	}

	if isTruthy(condition) {
		return t.Left.Evaluate()
	}

	return t.Right.Evaluate()
}

func (t VariableExpr) Evaluate() (LoxValue, error) {
	value, err := global_env.Get(t.Name)
	if err != nil {
		return nil, NewRuntimeError("undefined variable '" + t.Name.Lexme + "'")
	}

	return value, nil
}

func (t AssignExpr) Evaluate() (LoxValue, error) {
	value, err := t.Value.Evaluate()
	if err != nil {
		return nil, err
	}

	if err := global_env.Assign(t.Name.Lexme, value); err != nil {
		return nil, NewRuntimeError("undefined variable '" + t.Name.Lexme + "'")
	}

	return value, nil
}

func (t NothingExpr) Evaluate() (LoxValue, error) {
	return LoxNil{}, nil
}
