package main

import "fmt"

func Unique_int(a []int) []int {
	var uniq []int
	seen := make(map[int]bool)
	for _, x := range a {
		if seen[x] {
			continue
		}

		uniq = append(uniq, x)
		seen[x] = true
	}
	return uniq
}

func Sum_int(nums ...int) int {
	result := int(0)
	for _, x := range nums {
		result += x
	}
	return result
}
func main() {
	fmt.Println(Sum_int(Unique_int([]int{1, 2, 3, 4, 5, 6, 7, 1, 2, 3, 4, 5, 6, 7})...))
}
