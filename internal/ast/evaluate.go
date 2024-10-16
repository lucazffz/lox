package ast

import "github.com/LucazFFz/lox/internal/token"

type Evaluate interface {
	Evaluate() (Value, error)
}

type RuntimeError struct {
	message string
}

func NewRuntimeError(message string) RuntimeError {
	return RuntimeError{message: message}
}

func (r RuntimeError) Error() string {
	return "Runtime error: " + r.message
}

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
			return nil, NewRuntimeError("Operand must be a number")
		}
		return Number(-right.AsNumber()), nil

	}

	panic("should never reach here")
}

func (t Binary) Evaluate() (Value, error) {
	checkNumberOperands := func(left, right Value) error {
		if !isNumberValue(left) || !isNumberValue(right) {
			return NewRuntimeError("Operands must be numbers")
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
		if isStringValue(left) && isStringValue(right) {
			return String(left.AsString() + right.AsString()), nil
		}

		if isNumberValue(left) && isNumberValue(right) {
			return Number(left.AsNumber() + right.AsNumber()), nil
		}

		return nil, NewRuntimeError("Operands must be of the same type")

	case token.MINUS:
		if checkNumberOperands(left, right) != nil {
			return nil, err
		}
		return Number(left.AsNumber() - right.AsNumber()), nil
	case token.STAR:
		if checkNumberOperands(left, right) != nil {
			return nil, err
		}
		return Number(left.AsNumber() * right.AsNumber()), nil
	case token.SLASH:
		if checkNumberOperands(left, right) != nil {
			return nil, err
		}
		return Number(left.AsNumber() / right.AsNumber()), nil
	case token.GREATER:
		if checkNumberOperands(left, right) != nil {
			return nil, err
		}
		return Boolean(left.AsNumber() > right.AsNumber()), nil
	case token.GREATER_EQUAL:
		if checkNumberOperands(left, right) != nil {
			return nil, err
		}
		return Boolean(left.AsNumber() >= right.AsNumber()), nil
	case token.LESS:
		if checkNumberOperands(left, right) != nil {
			return nil, err
		}
		return Boolean(left.AsNumber() < right.AsNumber()), nil
	case token.LESS_EQUAL:
		if checkNumberOperands(left, right) != nil {
			return nil, err
		}
		return Boolean(left.AsNumber() <= right.AsNumber()), nil
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

func (t Nothing) Evaluate() (Value, error) {
	return Nil{}, nil
}
