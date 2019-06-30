package main

import "fmt"

// Heap is a container that lets you add elements into it and provides
// access to the minimal (or maximal) element.
type Heap(type T) struct {
	elems []T
	less  func(T, T) bool
}

// NewHeap constructs an empty heap container with the specified comparator.
//
// For the usual cases, a comparator can be created using one of the Min, Max, MinBy, and MaxBy functions:
//
//   NewHeap(Min(int))          // a minimum heap of ints
//   NewHeap(Max(string))       // a maximum heap of strings
//   NewHeap(MinBy(Person.Age)) // a minimum heap of Person values that sorts by the results of the Age method
func NewHeap(less func(x, y type T) bool) *Heap(T) {
	return &Heap(T){
		less: less,
	}
}

// The following four functions help create comparators. The same thing could (and probably should)
// be used for sorting, because it's very practical.

// Min returns a comparator for an orderable type T that compares using <.
func Min(type T ord) func(T, T) bool {
	return func(x, y T) bool {
		return x < y
	}
}

// Max returns a comparator for an oderable type T that compares using >.
func Max(type T ord) func(T, T) bool {
	return func(x, y T) bool {
		return x > y
	}
}

// MinBy returns a comparator that compares using < based on the results of the provided function.
func MinBy(by func(type T) type O ord) func(T, T) bool {
	return func(x, y T) bool {
		return by(x) < by(y)
	}
}

// MaxBy returns a comparator that compares using > based on the results of the provided function.
func MaxBy(by func(type T) type O ord) func(T, T) bool {
	return func(x, y T) bool {
		return by(x) > by(y)
	}
}

// Size returns the number of elements currently in the heap.
func (h *Heap(type T)) Size() int {
	return len(h.elems)
}

// Push adds an element to the heap.
func (h *Heap(type T)) Push(x T) {
	h.elems = append(h.elems, x)
	j := len(h.elems) - 1
	for {
		i := (j - 1) / 2 // parent
		if i == j || !h.less(h.elems[j], h.elems[i]) {
			break
		}
		h.elems[i], h.elems[j] = h.elems[j], h.elems[i]
		j = i
	}
}

// Top returns the current minimal (or maximal) element.
//
// Returns false if the heap is empty.
func (h *Heap(type T)) Top() (top T, ok bool) {
	if len(h.elems) == 0 {
		ok = false
		return
	}
	return h.elems[0], true
}

// Pop returns the current minimal (or maximal) element and removes it from the heap.
//
// Returns false if the heap is empty.
func (h *Heap(type T)) Pop() (top T, ok bool) {
	if len(h.elems) == 0 {
		ok = false
		return
	}
	n := len(h.elems) - 1
	h.elems[0], h.elems[n] = h.elems[n], h.elems[0]
	i := 0
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && h.less(h.elems[j2], h.elems[j1]) {
			j = j2 // = 2*i + 2  // right child
		}
		if !h.less(h.elems[j], h.elems[i]) {
			break
		}
		h.elems[i], h.elems[j] = h.elems[j], h.elems[i]
		i = j
	}
	top = h.elems[n]
	h.elems = h.elems[:n]
	return top, true
}

type Person struct {
	name string
	age  int
}

func (p Person) String() string {
	return fmt.Sprintf("%s is %d years old", p.name, p.age)
}

func (p Person) Name() string {
	return p.name
}

func (p Person) Age() int {
	return p.age
}

func main() {
	numbers := NewHeap(Min(int))
	for x := 10; x >= 1; x-- {
		numbers.Push(x)
	}
	for {
		x, ok := numbers.Pop()
		if !ok {
			break
		}
		fmt.Print(x, " ")
	}
	fmt.Println()

	fmt.Println()

	people := NewHeap(MaxBy(Person.Age))
	for _, person := range []Person{
		{"Michal", 23},
		{"Viktória", 20},
		{"Jano", 21},
		{"Martin", 18},
	} {
		people.Push(person)
	}
	for {
		p, ok := people.Pop()
		if !ok {
			break
		}
		fmt.Println(p)
	}
}
