package ast

import (
	"fmt"
)

type LoxValue interface {
	DebugPrint
	Type() LoxValueType
}

type Callable interface {
	LoxValue
	Call(arguments []LoxValue) (LoxValue, error)
	Arity() int
}

//go:generate stringer -type=LoxValueType
type LoxValueType uint8

type LoxObject struct{}

type LoxBoolean bool

type LoxNumber float64

type LoxString string

type LoxType struct {
	Typ LoxValueType
}

type LoxNil struct{}

type LoxFunction struct {
	FunctionStmt
	Closure *Environment
}

type NativeFunction struct {
	paramLen int
	Function func([]LoxValue) (LoxValue, error)
}

const (
	BOOLEAN LoxValueType = iota
	NUMBER
	NIL
	STRING
	OBJECT
	FUNCTION
	TYPE
)

func isBool(v LoxValue) bool {
	return v.Type() == BOOLEAN
}

func isNumber(v LoxValue) bool {
	return v.Type() == NUMBER
}

func isNil(v LoxValue) bool {
	return v.Type() == NIL
}

func isObject(v LoxValue) bool {
	return v.Type() == OBJECT
}

func isString(v LoxValue) bool {
	return v.Type() == STRING
}

func isTruthy(v LoxValue) bool {
	switch v.Type() {
	case BOOLEAN:
		return AsBoolean(v)
	case NIL:
		return false
	default:
		return true
	}
}

func valueToString(v LoxValue) (string, error) {
	switch v.Type() {
	case BOOLEAN:
		return fmt.Sprintf("%t", AsBoolean(v)), nil
	case NUMBER:
		return fmt.Sprintf("%.1f", AsNumber(v)), nil
	case NIL:
		return "nil", nil
	case STRING:
		return fmt.Sprintf("%s", AsString(v)), nil
	case OBJECT:
		return "object", nil
	case FUNCTION:
		return "", NewRuntimeError("cannot convert function to string")
	case TYPE:
		return fmt.Sprintf("<class '%s'>", v.(LoxType).Typ.String()), nil
	default:
		panic("should not reach here")
	}
}

func equals(v1 LoxValue, v2 LoxValue) bool {
	//    // the value loxValueType nil and the loxType nil are equal
	//    // for the other types, it makes sense to seperate the
	//    // value from the type, so that "some string" != type("some string")
	//    // but nil = type(nil)
	// if v1.Type() == NIL && AsType(v2).Typ == NIL ||
	// 	AsType(v1).Typ == NIL && v2.Type() == NIL {
	// 	return true
	// }

	if v1.Type() != v2.Type() {
		return false
	}

	switch v1.Type() {
	case BOOLEAN:
		return AsBoolean(v1) == AsBoolean(v2)
	case NUMBER:
		return AsNumber(v1) == AsNumber(v2)
	case NIL:
		return true
	case STRING:
		return AsString(v1) == AsString(v2)
	case OBJECT:
		return true
	case TYPE:
		return v1.(LoxType).Typ == v2.(LoxType).Typ
	default:
		return false
	}
}

func (v LoxBoolean) Type() LoxValueType {
	return BOOLEAN
}

func AsBoolean(v LoxValue) bool {
	if v, ok := v.(LoxBoolean); ok {
		return bool(v)
	}
	panic("Cannot convert non-boolean to boolean")
}

func AsNumber(v LoxValue) float64 {
	if v, ok := v.(LoxNumber); ok {
		return float64(v)
	}
	panic("Cannot convert non-number to number")
}

func AsString(v LoxValue) string {
	if v, ok := v.(LoxString); ok {
		return string(v)
	}
	panic("Cannot convert non-string to string")
}

func AsType(v LoxValue) LoxType {
	return LoxType{Typ: v.Type()}
}

func (v LoxNumber) Type() LoxValueType {
	return NUMBER
}

func (v LoxNil) Type() LoxValueType {
	return NIL
}

func (v LoxObject) Type() LoxValueType {
	return OBJECT
}

func (v LoxString) Type() LoxValueType {
	return STRING
}

func (v LoxFunction) Type() LoxValueType {
	return FUNCTION
}

func (v LoxType) Type() LoxValueType {
	return TYPE
}

func (t LoxFunction) Call(arguments []LoxValue) (LoxValue, error) {
	env := NewEnvironment(t.Closure)
	for i, param := range t.Parameters {
		env.Define(param.Lexme, arguments[i])
	}

	for _, stmt := range t.Body {
		if err := stmt.Evaluate(); err != nil {
			if ret, ok := err.(ReturnError); ok {
				return ret.Value, nil
			}
			return nil, err
		}
	}

	return LoxNil{}, nil
}

func (t LoxFunction) Arity() int {
	return len(t.Parameters)
}

func (t NativeFunction) Type() LoxValueType {
	return FUNCTION
}

func (t NativeFunction) DebugPrint() string {
	return ""
}

func (t NativeFunction) Call(arguments []LoxValue) (LoxValue, error) {
	if len(arguments) != t.Arity() {
		return nil, NewRuntimeError(fmt.Sprintf("expected %d arguments but got %d", t.Arity(), len(arguments)))
	}

	val, err := t.Function(arguments)

	// native functions should not return nil as value but may accidentally
	if val == nil && err == nil {
		return LoxNil{}, nil
	}

	return val, err
}

func (t NativeFunction) Arity() int {
	return t.paramLen
}
