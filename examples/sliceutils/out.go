package main

import (
	"fmt"
	"sort"
)

type Person struct {
	name	string
	age	int
}

func (p Person) Name() string {
	return p.name
}
func (p Person) Age() int {
	return p.age
}

func SortBy_int_Person(a []Person, by func(Person) int) {
	sort.Slice(a, func(i, j int) bool { return by(a[i]) < by(a[j]) })
}
func Map_Person_string(a []Person, f func(Person) string) []string {
	result := make([]string, len(a))
	for i := range a {
		result[i] = f(a[i])
	}
	return result
}

func Map_Person_int(a []Person, f func(Person) int) []int {
	result := make([]int, len(a))
	for i := range a {
		result[i] = f(a[i])
	}
	return result
}

func Map_int_string(a []int, f func(int) string) []string {
	result := make([]string, len(a))
	for i := range a {
		result[i] = f(a[i])
	}
	return result
}

func Concat_string(slices ...[]string) []string {
	total := 0
	for i := range slices {
		total += len(slices[i])
	}

	result := make([]string, 0, total)
	for i := range slices {
		result = append(result, slices[i]...)
	}
	return result
}

func Reverse_string(a []string) {
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
}
func main() {
	people := []Person{{"Michal", 23}, {"Viktória", 20}, {"Jano", 21}, {"Martin", 18}}
	SortBy_int_Person(people, Person.Age)
	fmt.Println(people)

	names := Map_Person_string(people, Person.Name)
	ages := Map_Person_int(people, Person.Age)
	ageStrings := Map_int_string(ages, func(a int) string { return fmt.Sprint(a) })

	everything := Concat_string(names, ageStrings)
	Reverse_string(everything)
	fmt.Println(everything)
}
