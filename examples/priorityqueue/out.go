package main

import "fmt"

type Person struct {
	name	string
	age	int
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
func Min_int() func(int, int) bool {
	return func(x, y int) bool { return x < y }
}

type Heap_int struct {
	elems	[]int
	less	func(int, int) bool
}

func (h *Heap_int) Size() int {
	return len(h.elems)
}

func (h *Heap_int) Push(x int) {
	h.elems = append(h.elems, x)
	j := len(h.elems) - 1
	for {

		i := (j - 1) / 2
		if i == j || !h.less(h.elems[j], h.elems[i]) {
			break
		}

		h.elems[i], h.elems[j] = h.elems[j], h.elems[i]
		j = i
	}
}
func (h *Heap_int) Top() (top int, ok bool) {
	if len(h.elems) == 0 {
		ok = false
		return
	}
	return h.elems[0], true
}
func (h *Heap_int) Pop() (top int, ok bool) {
	if len(h.elems) == 0 {
		ok = false
		return
	}

	n := len(h.elems) - 1
	h.elems[0], h.elems[n] = h.elems[n], h.elems[0]
	i := 0
	for {

		j1 := 2*i + 1
		if j1 >= n || j1 < 0 {
			break
		}

		j := j1
		if j2 := j1 + 1; j2 < n && h.less(h.elems[j2], h.elems[j1]) {
			j = j2
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

func NewHeap_int(less func(x, y int) bool) *Heap_int	{ return &Heap_int{less: less} }

func MaxBy_int_Person(by func(Person) int) func(Person, Person) bool {
	return func(x, y Person) bool { return by(x) > by(y) }
}

type Heap_Person struct {
	elems	[]Person
	less	func(Person, Person) bool
}

func (h *Heap_Person) Size() int {
	return len(h.elems)
}

func (h *Heap_Person) Push(x Person) {
	h.elems = append(h.elems, x)
	j := len(h.elems) - 1
	for {

		i := (j - 1) / 2
		if i == j || !h.less(h.elems[j], h.elems[i]) {
			break
		}

		h.elems[i], h.elems[j] = h.elems[j], h.elems[i]
		j = i
	}
}
func (h *Heap_Person) Top() (top Person, ok bool) {
	if len(h.elems) == 0 {
		ok = false
		return
	}
	return h.elems[0], true
}
func (h *Heap_Person) Pop() (top Person, ok bool) {
	if len(h.elems) == 0 {
		ok = false
		return
	}

	n := len(h.elems) - 1
	h.elems[0], h.elems[n] = h.elems[n], h.elems[0]
	i := 0
	for {

		j1 := 2*i + 1
		if j1 >= n || j1 < 0 {
			break
		}

		j := j1
		if j2 := j1 + 1; j2 < n && h.less(h.elems[j2], h.elems[j1]) {
			j = j2
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

func NewHeap_Person(less func(x, y Person) bool) *Heap_Person	{ return &Heap_Person{less: less} }
func main() {
	numbers := NewHeap_int(Min_int())
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

	people := NewHeap_Person(MaxBy_int_Person(Person.Age))
	for _, person := range []Person{{"Michal", 23}, {"Viktória", 20}, {"Jano", 21}, {"Martin", 18}} {

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
