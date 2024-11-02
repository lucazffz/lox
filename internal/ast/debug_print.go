package ast

import (
	"fmt"
	"strconv"
	"strings"
)

type DebugPrint interface {
	DebugPrint() string
}

func (t BinaryExpr) DebugPrint() string {
	return parenthesize(t.Op.Lexme, t.Left, t.Right)
}

func (t GroupingExpr) DebugPrint() string {
	return parenthesize("group", t.Expr)
}

func (t LiteralExpr) DebugPrint() string {
	return t.Value.DebugPrint()
}

func (t UnaryExpr) DebugPrint() string {
	return parenthesize(t.Op.Lexme, t.Right)
}

func (t TernaryExpr) DebugPrint() string {
	return parenthesize("ternary", t.Condition, t.Left, t.Right)
}

func (t VariableExpr) DebugPrint() string {
	return parenthesize(t.Name.Lexme)
}

func (t AssignExpr) DebugPrint() string {
	return fmt.Sprintf("(assign %s %s)", t.Name.Lexme, t.Value.DebugPrint())
}


func (t NothingExpr) DebugPrint() string {
	return parenthesize("Nothing")
}

func parenthesize(name string, exprs ...DebugPrint) string {
	var builder = strings.Builder{}
	builder.WriteString("(")
	builder.WriteString(name)

	for _, expr := range exprs {
		builder.WriteString(" ")
		builder.WriteString(expr.DebugPrint())
	}

	builder.WriteString(")")
	return builder.String()
}

// values
func (v LoxBoolean) DebugPrint() string {
	if v {
		return "true"
	} else {
		return "false"
	}
}

func (v LoxNumber) DebugPrint() string {
	return strconv.FormatFloat(AsNumber(v), 'f', -1, 64)
}

func (v LoxNil) DebugPrint() string {
	return "nil"
}

func (v LoxObject) DebugPrint() string {
	return "object"
}

func (v LoxString) DebugPrint() string {
	return AsString(v)
}

func (v LoxType) DebugPrint() string {
    return "type"
}

// statements
func (s ExpressionStmt) DebugPrint() string {
	return parenthesize("expr", s.Expr)
}

func (s PrintStmt) DebugPrint() string {
	return parenthesize("print", s.Expr)
}

func (s VarStmt) DebugPrint() string {
	return parenthesize("var", s.Initializer)
}

func (s IfStmt) DebugPrint() string {
	if s.ElseBranch != nil {
		return parenthesize("if", s.Condition, s.ThenBranch, s.ElseBranch)
	}
	return parenthesize("if", s.Condition, s.ThenBranch)
}

func (s WhileStmt) DebugPrint() string {
	return parenthesize("while", s.Condition, s.Body)
}

func (s BlockStmt) DebugPrint() string {
	// cannot do parenthesize("block", s.Statements...)
	// because go will not convert from Stmt[] to PrettyPrint[]
	// because it generally does not do implicit conversions with time
	// complexity > O(1) apparently
	args := make([]DebugPrint, len(s.Statements))
	for i := range s.Statements {
		args[i] = s.Statements[i]
	}
	return parenthesize("block", args...)
}

func (s BreakStmt) DebugPrint() string {
	return parenthesize("break")
}

func (s ReturnStmt) DebugPrint() string {
    return parenthesize("return", s.Expr)
}

func (t FunctionStmt) DebugPrint() string {
	return parenthesize("function")
}


func (t CallStmt) DebugPrint() string {
	// args := make([]PrettyPrint, len(t.Arguments)+1)
	// args[0] = t.Callee
	// for i := range args {
	// 	args[i+1] = t.Arguments[i]
	// }
	// return parenthesize("call", args...)
	//
    return ""
}

