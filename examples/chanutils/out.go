package main

import "fmt"

func Elems_string(a ...string) <-chan string {
	ch := make(chan string)
	go func() {
		for _, x := range a {
			ch <- x
		}

		close(ch)
	}()
	return ch
}

func Elems_int(a ...int) <-chan int {
	ch := make(chan int)
	go func() {
		for _, x := range a {
			ch <- x
		}

		close(ch)
	}()
	return ch
}

func Map_int_string(in <-chan int, f func(int) string) <-chan string {
	out := make(chan string)
	go func() {
		for x := range in {
			out <- f(x)
		}

		close(out)
	}()
	return out
}

func Merge_string(chans ...<-chan string) <-chan string {
	merged := make(chan string)
	done := make(chan bool)
	for _, ch := range chans {
		ch := ch
		go func() {
			for x := range ch {
				merged <- x
			}

			done <- true
		}()
	}
	go func() {
		for range chans {
			<-done
		}

		close(merged)
	}()
	return merged
}
func main() {
	letters := Elems_string("A", "B", "C", "D", "E")
	numbers := Elems_int(1, 2, 3, 4, 5)

	everything := Merge_string(letters, Map_int_string(numbers, func(x int) string { return fmt.Sprint(x) }))
	for s := range everything {
		fmt.Println(s)
	}
}
