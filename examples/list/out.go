package main

import (
	"fmt"
	"math"
)

type List_float64 struct {
	First	float64
	Rest	*List_float64
}

func (l *List_float64) Prepend(x float64) *List_float64 {
	return &List_float64{First: x,
		Rest:	l}
}
func (l *List_float64) Empty() bool {
	return l == nil
}
func (l *List_float64) Slice() []float64 {
	var elems []float64
	for !l.Empty() {
		elems = append(elems, l.First)
		l = l.Rest
	}
	return elems
}
func Empty_float64() *List_float64 {
	return nil
}

func Elems_float64(xs ...float64) *List_float64 {
	list := Empty_float64()
	for i := len(xs) - 1; i >= 0; i-- {
		list = list.Prepend(xs[i])
	}
	return list
}

func Map_float64_float64(l *List_float64, f func(float64) float64) *List_float64 {
	if l.Empty() {
		return Empty_float64()
	}
	return Map_float64_float64(l.Rest, f).Prepend(f(l.First))
}
func main() {
	list1 := Elems_float64(1.0, 4.0, 9.0, 16.0)
	list2 := Map_float64_float64(list1, math.Sqrt)
	fmt.Println(list2.Slice())
}
