package ast

import "github.com/LucazFFz/lox/internal/token"

type EvaluateExpr interface {
	Evaluate() (Value, error)
}

type EvaluateStmt interface {
	Evaluate() error
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
func (s Expression) Evaluate() error {
	_, err := s.Expr.Evaluate()
	return err
}

func (s Print) Evaluate() error {
	expr, err := s.Expr.Evaluate()
	if err != nil {
		return err
	}

	println(expr.Print())
	return nil
}

func (s Var) Evaluate() error {
	if (s.Initializer == Nothing{}) {
		environment.Define(s.Name.Lexme, Nil{})
	}

	value, err := s.Initializer.Evaluate()
	if err != nil {
		return err
	}

	environment.Define(s.Name.Lexme, value)
	return nil
}

// expressions
func (t Literal) Evaluate() (Value, error) {
	return t.Value, nil
}

func (t Grouping) Evaluate() (Value, error) {
	return t.Expr.Evaluate()
}

func (t Unary) Evaluate() (Value, error) {
	right, err := t.Right.Evaluate()
	if err != nil {
		return nil, err
	}
	switch t.Op.Type {
	case token.BANG:
		return Boolean(!isTruthy(right)), nil
	case token.MINUS:
		if !isNumberValue(right) {
			return nil, NewRuntimeError("operand must be a number")
		}
		return Number(-right.AsNumber()), nil

	}

	panic("should never reach here")
}

func (t Binary) Evaluate() (Value, error) {
	checkNumberOperands := func(left, right Value) error {
		if !isNumberValue(left) || !isNumberValue(right) {
			return NewRuntimeError("both operands must be numbers")
		}

		return nil
	}

	checkStringOperands := func(left, right Value) error {
		if !isStringValue(left) || !isStringValue(right) {
			return NewRuntimeError("both operands must be strings")
		}

		return nil
	}
	right, err := t.Right.Evaluate()
	if err != nil {
		return nil, err
	}
	left, err := t.Left.Evaluate()
	if err != nil {
		return nil, err
	}

	switch t.Op.Type {
	case token.PLUS:
		if err := checkNumberOperands(left, right); err == nil {
			return Number(left.AsNumber() + right.AsNumber()), nil
		}

		if err := checkStringOperands(left, right); err == nil {
			return String(left.AsString() + right.AsString()), nil
		}

		return nil, NewRuntimeError("operands must be of same type")
	case token.MINUS:
		if err := checkNumberOperands(left, right); err != nil {
			return nil, err
		}
		return Number(left.AsNumber() - right.AsNumber()), nil
	case token.STAR:
		if err := checkNumberOperands(left, right); err != nil {
			return nil, err
		}
		return Number(left.AsNumber() * right.AsNumber()), nil
	case token.SLASH:
		if err := checkNumberOperands(left, right); err != nil {
			return nil, err
		}

		if right.AsNumber() == 0 {
			return nil, NewRuntimeError("division by zero")
		}

		return Number(left.AsNumber() / right.AsNumber()), nil
	case token.GREATER:
		if err := checkNumberOperands(left, right); err == nil {
			return Boolean(left.AsNumber() > right.AsNumber()), nil
		}

		if err := checkStringOperands(left, right); err == nil {
			return Boolean(left.AsString() > right.AsString()), nil
		}

		return nil, NewRuntimeError("operands must be of same type")
	case token.GREATER_EQUAL:
		if err := checkNumberOperands(left, right); err == nil {
			return Boolean(left.AsNumber() >= right.AsNumber()), nil
		}

		if err := checkStringOperands(left, right); err == nil {
			return Boolean(left.AsString() >= right.AsString()), nil
		}

		return nil, NewRuntimeError("operands must be of same type")
	case token.LESS:
		if err := checkNumberOperands(left, right); err == nil {
			return Boolean(left.AsNumber() < right.AsNumber()), nil
		}

		if err := checkStringOperands(left, right); err == nil {
			return Boolean(left.AsString() < right.AsString()), nil
		}

		return nil, NewRuntimeError("operands must be of same type")
	case token.LESS_EQUAL:
		if err := checkNumberOperands(left, right); err == nil {
			return Boolean(left.AsNumber() <= right.AsNumber()), nil
		}

		if err := checkStringOperands(left, right); err == nil {
			return Boolean(left.AsString() <= right.AsString()), nil
		}

		return nil, NewRuntimeError("operands must be of same type")
	case token.EQUAL_EQUAL:
		return Boolean(equals(left, right)), nil
	case token.BANG_EQUAL:
		return Boolean(!equals(left, right)), nil
	}

	panic("should never reach here")
}

func (t Ternary) Evaluate() (Value, error) {
	condition, err := t.Condition.Evaluate()
	if err != nil {
		return nil, err
	}

	if isTruthy(condition) {
		return t.Left.Evaluate()
	}

	return t.Right.Evaluate()
}

func (t Variable) Evaluate() (Value, error) {
	value, err := environment.Get(t.Name.Lexme)
	if err != nil {
		return nil, NewRuntimeError("undefined variable '" + t.Name.Lexme + "'")
	}

	return value, nil
}

func (t Assign) Evaluate() (Value, error) {
	value, err := t.Value.Evaluate()
	if err != nil {
        return nil, err
	}

	if err := environment.Assign(t.Name.Lexme, value); err != nil {
		return nil, NewRuntimeError("undefined variable '" + t.Name.Lexme + "'")
	}

	return value, nil
}

func (t Nothing) Evaluate() (Value, error) {
	return Nil{}, nil
}
