package main

import (
	"fmt"
)

func Iter(a ...type T) <-chan T {
	ch := make(chan T)
	go func() {
		for _, x := range a {
			ch <- x
		}
		close(ch)
	}()
	return ch
}

func Collect(ch <-chan type T) []T {
	var xs []T
	for x := range ch {
		xs = append(xs, x)
	}
	return xs
}

func Merge(chans ...<-chan type T) <-chan T {
	merged := make(chan T)
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

func Map(in <-chan type T, f func(T) type U) <-chan U {
	out := make(chan U)
	go func() {
		for x := range in {
			out <- f(x)
		}
		close(out)
	}()
	return out
}

func Interfaces(a []type T) []interface{} {
	ifaces := make([]interface{}, len(a))
	for i, x := range a {
		ifaces[i] = x
	}
	return ifaces
}

func ToString(type T) func(T) string {
	return func(x T) string {
		return fmt.Sprint(x)
	}
}

func FromString(type T) func(string) T {
	return func(s string) T {
		var x T
		fmt.Sscan(s, &x)
		return x
	}
}

func main() {
	ch1 := Map(Iter(1, 2, 3, 4), ToString(int))
	ch2 := Iter("A", "B", "C", "D")
	fmt.Println(Interfaces(Collect(Merge(ch1, ch2)))...)
}
