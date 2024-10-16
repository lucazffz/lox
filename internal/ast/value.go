package ast

type ValueType uint8

type Object struct{}

type Boolean bool

type Number float64

type String string

type Nil struct{}

const (
	BOOLEAN ValueType = iota
	NUMBER
	NIL
	STRING
	OBJECT
)

type Value interface {
	PrettyPrint
	Type() ValueType
	AsBoolean() bool
	AsNumber() float64
	AsObject() Object
	AsString() string
}

func isBoolValue(v Value) bool {
	return v.Type() == BOOLEAN
}

func isNumberValue(v Value) bool {
	return v.Type() == NUMBER
}

func isNilValue(v Value) bool {
	return v.Type() == NIL
}

func isObjectValue(v Value) bool {
	return v.Type() == OBJECT
}

func isStringValue(v Value) bool {
	return v.Type() == STRING
}

func isTruthy(v Value) bool {
	switch v.Type() {
	case BOOLEAN:
		return v.AsBoolean()
	case NIL:
		return false
	default:
		return true
	}
}

func equals(v1 Value, v2 Value) bool {
	if v1.Type() != v2.Type() {
		return false
	}

	switch v1.Type() {
	case BOOLEAN:
		return v1.AsBoolean() == v2.AsBoolean()
	case NUMBER:
		return v1.AsNumber() == v2.AsNumber()
	case NIL:
		return true
	case STRING:
		return v1.AsString() == v2.AsString()
	case OBJECT:
		return true
	default:
		return false
	}
}

// BOOLEAN
func (v Boolean) Type() ValueType {
	return BOOLEAN
}

func (v Boolean) AsBoolean() bool {
	return bool(v)
}

func (v Boolean) AsNumber() float64 {
	panic("Cannot convert boolean to number")
}

func (v Boolean) AsObject() Object {
	panic("Cannot convert boolean to object")
}

func (v Boolean) AsString() string {
	panic("Cannot convert boolean to string")
}

// NUMBER
func (v Number) Type() ValueType {
	return NUMBER
}

func (v Number) AsBoolean() bool {
	panic("Cannot convert number to boolean")
}

func (v Number) AsNumber() float64 {
	return float64(v)
}

func (v Number) AsObject() Object {
	panic("Cannot convert number to object")
}

func (v Number) AsString() string {
	panic("Cannot convert number to string")
}

// NIL
func (v Nil) Type() ValueType {
	return NIL
}

func (v Nil) AsBoolean() bool {
	panic("Cannot convert nil to boolean")
}

func (v Nil) AsNumber() float64 {
	panic("Cannot convert nil to number")
}

func (v Nil) AsObject() Object {
	panic("Cannot convert nil to object")
}

func (v Nil) AsString() string {
	panic("Cannot convert nil to string")
}

// OBJECT
func (v Object) Type() ValueType {
	return OBJECT
}

func (v Object) AsBoolean() bool {
	panic("Cannot convert object to boolean")
}

func (v Object) AsNumber() float64 {
	panic("Cannot convert object to number")
}

func (v Object) AsObject() Object {
	return v
}

func (v Object) AsString() string {
	panic("Cannot convert object to string")
}

// STRING
func (v String) Type() ValueType {
	return STRING
}

func (v String) AsBoolean() bool {
	panic("Cannot convert string to boolean")
}

func (v String) AsNumber() float64 {
	panic("Cannot convert string to number")
}

func (v String) AsObject() Object {
	panic("Cannot convert string to object")
}

func (v String) AsString() string {
	return string(v)
}
