package main

import (
	"fmt"
)

type List(type T) struct {
	First T
	Rest  *List(T)
}

func Empty(type T) *List(T) {
	return nil
}

func Cons(x type T, l *List(T)) *List(T) {
	return &List(T){
		First: x,
		Rest:  l,
	}
}

func Map(l *List(type T), f func(T) type U) *List(U) {
	if l == nil {
		return nil
	}
	return Cons(f(l.First), Map(l.Rest, f))
}

func FromSlice(a ...type T) *List(T) {
	list := Empty(T)
	for i := len(a)-1; i >= 0; i-- {
		list = Cons(a[i], list)
	}
	return list
}

func (l *List(type T)) Slice() []T {
	var a []T
	for l != nil {
		a = append(a, l.First)
		l = l.Rest
	}
	return a
}

type Pair(type A, type B) struct {
	A A
	B B
}

func (p *Pair(type A, A)) Swap() {
	p.A, p.B = p.B, p.A
}

func main() {
	list := FromSlice(1, 2, 3, 4, 5, 6, 7)
	list.First += 10
	list2 := Map(list, func(x int) string {
		return fmt.Sprint(x)
	})
	fmt.Println(list2.Slice())

	list3 := &List(int){
		First: 1,
		Rest: nil,
	}
	fmt.Println(list3.Slice())

	p := Pair(string, int){"A", 2}
	fmt.Println(p.A, p.B)

	q := Pair(string, string){"A", "B"}
	q.Swap()
	fmt.Println(q.A, q.B)
}
