package main

import "fmt"

// Elems returns a channel that streams all the supplied elements one by one.
func Elems(a ...type T) <-chan T {
	ch := make(chan T)
	go func() {
		for _, x := range a {
			ch <- x
		}
		close(ch)
	}()
	return ch
}

// Pipe redirects all the data from one channel to another one.
func Pipe(from <-chan type T, to chan<- T) {
	for x := range from {
		to <- x
	}
	close(to)
}

// Map transforms a channel of type T into a channel of type U by transforming each of the
// values sent on the channel using the supplied function.
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

// Merge merges all the supplied channels into a single channel. Any value sent by any of
// the original channels will appear on the merged channel. The returned channel gets closed when
// all of the supplied channels get closed.
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

func main() {
	letters := Elems("A", "B", "C", "D", "E")
	numbers := Elems(1, 2, 3, 4, 5)

	everything := Merge(letters, Map(numbers, func(x int) string {
		return fmt.Sprint(x)
	}))

	for s := range everything {
		fmt.Println(s)
	}
}
