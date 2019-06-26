package main

import "fmt"

func Min_int(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func Min_int32(x, y int32) int32 {
	if x < y {
		return x
	}
	return y
}

func Max_float64(x, y float64) float64 {
	if x > y {
		return x
	}
	return y
}

func Max_string(x, y string) string {
	if x > y {
		return x
	}
	return y
}

func Sum_int(nums ...int) int {
	result := int(0)
	for _, x := range nums {
		result += x
	}
	return result
}

func Product_int(nums ...int) int {
	result := int(1)
	for _, x := range nums {
		result *= x
	}
	return result
}

func Sum_byte(nums ...byte) byte {
	result := byte(0)
	for _, x := range nums {
		result += x
	}
	return result
}
func main() {
	fmt.Println(Min_int(7, 9))
	fmt.Println(Min_int32(int32(10), 93))
	fmt.Println(Max_float64(3.14, 31.4))
	fmt.Println(Max_string("A", "B"))

	fmt.Println(Sum_int(1, 2, 3, 4, 5, 6, 7, 8, 9, 10))
	fmt.Println(Product_int(1, 2, 3, 4, 5))

	bytes := ([]byte)("Hello, world!")
	fmt.Println(Sum_byte(bytes...))
}
