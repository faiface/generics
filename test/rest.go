package main

import "fmt"

func Min(x, y type T ord) T {
	if x < y {
		return x
	}
	return y
}

func Max(x, y type T ord) T {
	if x > y {
		return x
	}
	return y
}

func Minimum(xs ...type X ord) X {
	min := xs[0]
	for i := 1; i < len(xs); i++ {
		min = Min(min, xs[i])
	}
	return min
}

func Sum(nums ...type T num) T {
	result := T(0)
	for _, x := range nums {
		result += x
	}
	return result
}

func Unique(a []type T eq) []T {
	var uniq []T
	seen := make(map[T]bool)
	for _, x := range a {
		if seen[x] {
			continue
		}
		uniq = append(uniq, x)
		seen[x] = true
	}
	return uniq
}

func main() {
	fmt.Println(Sum(Unique([]int{1, 2, 3, 4, 5, 6, 7, 1, 2, 3, 4, 5, 6, 7})...))
}
