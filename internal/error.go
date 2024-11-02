package lox

import (
	"fmt"
	"github.com/LucazFFz/lox/internal/ast"
)

type RuntimeError struct {
	Code ErrorCode
}

func (e RuntimeError) Error() string {
	return fmt.Sprintf("runtime error - %s\n", e.Code.Print())
}

func NewRuntimeError(code ErrorCode) RuntimeError {
	return RuntimeError{Code: code}
}

type ScanError struct {
}

type ParseError struct {
}

type BreakError struct {
	RuntimeError
}

type ReturnError struct {
	RuntimeError
	Value ast.LoxValue
}

type ErrorCode uint8

type ErrorContext struct {
	Source     *[]byte
	ByteOffset uint
	Len        int
}

type Position struct {
	Line   int
	Column int
}

func (e ErrorCode) Print() string {
	panic("todo")
}

const (
	// Scanner errors
	InvalidToken ErrorCode = iota

	// Parser errors

	// Runtime errors

)
