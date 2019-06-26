package main

import "fmt"

// Min returns the smaller one of the two.
func Min(x, y type T ord) T {
	if x < y {
		return x
	}
	return y
}

// Max returns the bigger one of the two.
func Max(x, y type T ord) T {
	if x > y {
		return x
	}
	return y
}

// Sum returns the sum of the numbers.
func Sum(nums ...type T num) T {
	result := T(0)
	for _, x := range nums {
		result += x
	}
	return result
}

// Product returns the product of the numbers.
func Product(nums ...type T num) T {
	result := T(1)
	for _, x := range nums {
		result *= x
	}
	return result
}

func main() {
	fmt.Println(Min(7, 9))
	fmt.Println(Min(int32(10), 93))
	fmt.Println(Max(3.14, 31.4))
	fmt.Println(Max("A", "B"))

	fmt.Println(Sum(1, 2, 3, 4, 5, 6, 7, 8, 9, 10))
	fmt.Println(Product(1, 2, 3, 4, 5))

	bytes := ([]byte)("Hello, world!")
	fmt.Println(Sum(bytes...))
}
