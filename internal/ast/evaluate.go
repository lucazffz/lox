package ast

import "github.com/LucazFFz/lox/internal/token"

type EvaluateExpr interface {
	Evaluate() (Value, error)
}

type EvaluateStmt interface {
	Evaluate() error
}

// evaluating a break statement will return a BreakError
// (a type of runtime error) which we can catch when
// evaluating a while loop and break out of the loop
// unsure if this is the best way to handle this
type BreakError struct{}

func (b BreakError) Error() string {
	return "runtime error - unexpected break statement outside of loop\n"
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

func (s Block) Evaluate() error {
	previous := environment
	environment = NewEnvironment(environment)

	for _, statement := range s.Statements {
		err := statement.Evaluate()
		if err != nil {
			environment = previous
			return err
		}
	}

	environment = previous
	return nil
}

func (s Var) Evaluate() error {
	if (s.Initializer == Nothing{}) {
		environment.Define(s.Name, Nil{})
	}

	value, err := s.Initializer.Evaluate()
	if err != nil {
		return err
	}

	environment.Define(s.Name, value)
	return nil
}

func (s If) Evaluate() error {
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

func (s While) Evaluate() error {
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

func (s Break) Evaluate() error {
	return BreakError{}
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

	evaluateOperands := func() (Value, Value, error) {
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
			return Number(left.AsNumber() + right.AsNumber()), nil
		}

		if err := checkStringOperands(left, right); err == nil {
			return String(left.AsString() + right.AsString()), nil
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
		return Number(left.AsNumber() - right.AsNumber()), nil
	case token.STAR:
		left, right, err := evaluateOperands()
		if err != nil {
			return nil, err
		}
		if err := checkNumberOperands(left, right); err != nil {
			return nil, err
		}
		return Number(left.AsNumber() * right.AsNumber()), nil
	case token.SLASH:
		left, right, err := evaluateOperands()
		if err != nil {
			return nil, err
		}
		if err := checkNumberOperands(left, right); err != nil {
			return nil, err
		}

		if right.AsNumber() == 0 {
			return nil, NewRuntimeError("division by zero")
		}

		return Number(left.AsNumber() / right.AsNumber()), nil
	case token.GREATER:
		left, right, err := evaluateOperands()
		if err != nil {
			return nil, err
		}
		if err := checkNumberOperands(left, right); err == nil {
			return Boolean(left.AsNumber() > right.AsNumber()), nil
		}

		if err := checkStringOperands(left, right); err == nil {
			return Boolean(left.AsString() > right.AsString()), nil
		}

		return nil, NewRuntimeError("operands must be of same type")
	case token.GREATER_EQUAL:
		left, right, err := evaluateOperands()
		if err != nil {
			return nil, err
		}
		if err := checkNumberOperands(left, right); err == nil {
			return Boolean(left.AsNumber() >= right.AsNumber()), nil
		}

		if err := checkStringOperands(left, right); err == nil {
			return Boolean(left.AsString() >= right.AsString()), nil
		}

		return nil, NewRuntimeError("operands must be of same type")
	case token.LESS:
		left, right, err := evaluateOperands()
		if err != nil {
			return nil, err
		}
		if err := checkNumberOperands(left, right); err == nil {
			return Boolean(left.AsNumber() < right.AsNumber()), nil
		}

		if err := checkStringOperands(left, right); err == nil {
			return Boolean(left.AsString() < right.AsString()), nil
		}

		return nil, NewRuntimeError("operands must be of same type")
	case token.LESS_EQUAL:
		left, right, err := evaluateOperands()
		if err != nil {
			return nil, err
		}
		if err := checkNumberOperands(left, right); err == nil {
			return Boolean(left.AsNumber() <= right.AsNumber()), nil
		}

		if err := checkStringOperands(left, right); err == nil {
			return Boolean(left.AsString() <= right.AsString()), nil
		}

		return nil, NewRuntimeError("operands must be of same type")
	case token.EQUAL_EQUAL:
		left, right, err := evaluateOperands()
		if err != nil {
			return nil, err
		}
		return Boolean(equals(left, right)), nil
	case token.BANG_EQUAL:
		left, right, err := evaluateOperands()
		if err != nil {
			return nil, err
		}
		return Boolean(!equals(left, right)), nil
	}

	panic("should never reach here (binary)")
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
	value, err := environment.Get(t.Name)
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

	if err := environment.Assign(t.Name, value); err != nil {
		return nil, NewRuntimeError("undefined variable '" + t.Name.Lexme + "'")
	}

	return value, nil
}

func (t Nothing) Evaluate() (Value, error) {
	return Nil{}, nil
}
