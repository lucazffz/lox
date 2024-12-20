package ast

type Stack[T any] struct {
	stack []T
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{stack: make([]T, 0)}
}

func StackWithCapacity[T any](capacity int) *Stack[T] {
	return &Stack[T]{stack: make([]T, 0, capacity)}
}

func (s *Stack[T]) Push(item T) {
	s.stack = append(s.stack, item)
}

func (s *Stack[T]) Pop() (T, bool) {
	if s.IsEmpty() {
		var zero T
		return zero, false
	}
	item := s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]
	return item, true
}

func (s *Stack[T]) Peek() (T, bool) {
	if s.IsEmpty() {
		var zero T
		return zero, false
	}
	return s.stack[len(s.stack)-1], true
}

func (s *Stack[T]) IsEmpty() bool {
	return len(s.stack) == 0
}

func (s *Stack[T]) Len() int {
	return len(s.stack)
}
