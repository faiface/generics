package main

import (
	"fmt"
	"math"
)

// List is a generic singly-linked list. Just like in LISP.
type List(type T) struct {
	First T
	Rest  *List(T)
}

// Empty returns an empty list of type T.
func Empty(type T) *List(T) {
	return nil
}

// Prepend returns a new list with x prepended before l.
func (l *List(type T)) Prepend(x T) *List(T) {
	return &List(T){
		First: x,
		Rest:  l,
	}
}

// Elems constructs a linked list containing the given elements.
func Elems(xs ...type T) *List(T) {
	list := Empty(T)
	for i := len(xs)-1; i >= 0; i-- {
		list = list.Prepend(xs[i])
	}
	return list
}

// Empty returns whether l is an empty list.
func (l *List(type T)) Empty() bool {
	return l == nil
}

// Slice collects all elements from the list into a slice.
func (l *List(type T)) Slice() []T {
	var elems []T
	for !l.Empty() {
		elems = append(elems, l.First)
		l = l.Rest
	}
	return elems
}

// Map returns a new list where each element from the original list is transformed by f.
func Map(l *List(type T), f func(T) type U) *List(U) {
	if l.Empty() {
		return Empty(U)
	}
	return Map(l.Rest, f).Prepend(f(l.First))
}

func main() {
	list1 := Elems(1.0, 4.0, 9.0, 16.0)
	list2 := Map(list1, math.Sqrt)
	fmt.Println(list2.Slice())
}
