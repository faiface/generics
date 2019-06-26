package main

import (
	"fmt"
	"sort"
)

// Reverse reverses any slice in place.
func Reverse(a []type T) {
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
}

// Concat returns a concatenation of multiple slices of the same type.
func Concat(slices ...[]type T) []T {
	total := 0
	for i := range slices {
		total += len(slices[i])
	}
	result := make([]T, 0, total)
	for i := range slices {
		result = append(result, slices[i]...)
	}
	return result
}

// Map returns a new slice where each element from the original slice is transformed by f.
func Map(a []type T, f func(T) type U) []U {
	result := make([]U, len(a))
	for i := range a {
		result[i] = f(a[i])
	}
	return result
}

// Interfaces converts a slice of an arbitrary type to a slice of empty interfaces.
func Interfaces(a []type T) []interface{} {
	ifaces := make([]interface{}, len(a))
	for i := range a {
		ifaces[i] = a[i]
	}
	return ifaces
}

// Sort sorts a slice of an arbitrary orderable type.
func Sort(a []type T ord) {
	sort.Slice(a, func(i, j int) bool {
		return a[i] < a[j]
	})
}

// SortBy sorts a slice by a key. For example:
//
//   SortBy(people, (*Person).Age)
func SortBy(a []type T, by func(T) type O ord) {
	sort.Slice(a, func(i, j int) bool {
		return by(a[i]) < by(a[j])
	})
}

// SortWith sorts a slice using a custom comparator.
func SortWith(a []type T, with func(T, T) bool) {
	sort.Slice(a, func(i, j int) bool {
		return with(a[i], a[j])
	})
}

type Person struct {
	name string
	age  int
}

func (p Person) Name() string {
	return p.name
}

func (p Person) Age() int {
	return p.age
}

func main() {
	people := []Person{
		{"Michal", 23},
		{"Viktória", 20},
		{"Jano", 21},
		{"Martin", 18},
	}
	SortBy(people, Person.Age)
	fmt.Println(people)

	names := Map(people, Person.Name)
	ages := Map(people, Person.Age)
	ageStrings := Map(ages, func(a int) string {
		return fmt.Sprint(a)
	})

	everything := Concat(names, ageStrings)
	Reverse(everything)
	fmt.Println(everything)
}
