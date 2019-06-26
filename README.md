# A proof-of-concept implementation of my generics proposal for Go

This program translates a Go file that uses generics into a regular Go file that can be run.

I have taken measures to prevent you from running this in production. Pleas, do **not** run this in production. The single measure taken is that this program only translates a single file. This means that generic functions and types are only usable within that one file.

Here's a trivial example.

```go
// reverse.go

package main

import "fmt"

func Reverse(a []type T) {
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
}

func main() {
	a := []int{1, 2, 3, 4, 5}
	b := []string{"A", "B", "C"}
	Reverse(a)
	Reverse(b)
	fmt.Println(a)
	fmt.Println(b)
}
```

Here we have a file called `reverse.go` that uses generics. Here's how we translate it:

```
$ generics -out out.go reverse.go
```

A here's what we get!

```go
package main

import "fmt"

func Reverse_int(a []int) {
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
}

func Reverse_string(a []string) {
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
}

func main() {
	a := []int{1, 2, 3, 4, 5}
	b := []string{"A", "B", "C"}
	Reverse_int(a)
	Reverse_string(b)
	fmt.Println(a)
	fmt.Println(b)
}
```

Then, of course, we can run `out.go`:

```
$ go run out.go
[5 4 3 2 1]
[C B A]
```

## More example

That was just a silly little example. For more complex examples, take a look into the [`examples`](examples/) directory:
- TODO
- TODO
- TODO
- TODO
- TODO

## The proposal

This is a refined version of a proposal I submited a few weeks ago. [You can find it here.](https://gist.github.com/faiface/e5f035f46e88e96231c670abf8cab63f)

This version is very similar to the original proposal, it only differs in three things:
1. The `gen` keyword has been replaced with two keywords: `type` and `const`. This implementation only implements the `type` keyword, `const` will be described below nonetheless.
2. An `ord` type restriction addition to the previously described `eq` and `num`.
3. The `type` keyword now must appear also in the declarations of generic types.

Now I will describe the proposal as concisely as I can. If you have questions, scroll down to the [FAQ](#FAQ) section.

### The `type` keyword

Let's start with generic functions. Here's a **pseudocode** of a generic `Map` function on slices:

```go
// PSEUDOCODE!!
func Map(a []T, f func(T) U) []U {
    result := make([]U, len(a))
    for i := range a {
        result[i] = f(a[i])
    }
    return result
}
```

> In case you don't know, a `Map` function takes a slice and a function and returns a new slice with each element replaced by the result of the function applied to the original element.
>
> For example: `Map([]float64{1, 4, 9, 16}, math.Sqrt)` returns a new slice `[]float64{1, 2, 3, 4}`, taking the square root of each of the original elements.

To make this **valid** Go code under my proposal, all you need to do is to mark the _first_ (and only the first) occurrence of each type parameter (= an unknown type) in the signature with the `type` keyword. The `Map` function has two:

```go
//           here              here
//            \/                \/
func Map(a []type T, f func(T) type U) []U {
    result := make([]U, len(a))
    for i := range a {
        result[i] = f(a[i])
    }
    return result
}
```

Nothing else changed.

There are three rules about the placement of the `type` keyword in signatures:
1. It's only allowed in package-level function declarations.
2. In functions, it's only allowed in the list of parameters. Particularly, it's **disallowed** in the list of results.
3. In methods, it's only allowed in the receiver type.

### Unnamed type parameters

Okay, so no `type` in the list of results. But how do we do a function like this? The only occurrence of the `T` type is in the result:

```go
// DISALLOWED!!
func Read() type T {
    var x T
    fmt.Scan(&x)
    return x
}
```

To make this function work, we need to use an _unnamed type parameter_. It's basically a dummy generic parameter:

```go
func Read(type T) T {
    var x T
    fmt.Scan(&x)
    return x
}
```

The value of the unnamed parameter is irrelavant. We're only interested in the type. That's why when calling the `Read` function, we pass in the type directly:

```go
func main() {
    name := Read(string)
    age := Read(int)
    fmt.Printf("%s is %d years old.", name, age)
}
```

Don't worry, this doesn't send us to the [dependent typing land](https://en.wikipedia.org/wiki/Dependent_type) because we can't return types, only accept them.

This notation also makes it possible to give a type to the built-in `new` function:

```go
func new(type T) *T {
    var x T
    return &x
}
```

### Restricting types

Some functions (or types) want to declare that they don't work with all types, but only with ones that satisfy some conditions. For example, the keys of a map must be comparable. That is a restriction. A `Min` function only works on types that are orderable (i.e. can be compared with `<`).

Initially, my proposal excluded any support for restricting types for the purpose of simplicity. The [contracts proposal](https://go.googlesource.com/proposal/+/master/design/go2draft-contracts.md) by the Go team has received (justifiable) criticism for introducing complexity by supporting _contracts_, which make it possible to specify arbitrary restrictions on types.

But some restrictions are extremely useful. That's why I eventually decided to include three possible restrictions that should cover majority of use-cases. This decision is governed by the [80/20 principle](https://en.wikipedia.org/wiki/Pareto_principle).

Here are the three possible restrictions:
1. **`eq`** - Comparable with `==` and `!=`. Usable as map keys.
2. **`ord`** - Comparable with `<`, `>`, `<=`, `>=`, `==`, `!=`. Subset of `eq`.
3. **`num`** - All numeric types: `int*`, `uint*`, `float*`, and `complex*`. Operators `+`, `-`, `*`, `/`, `==`, `!=`, and converting from untyped integer constants works.

To use a type restriction, place its name right after the type parameter name in its first occurrence.

For example, here's the generic `Min` function:

```go
//                   here
//                    v
func Min(x, y type T ord) T {
    if x < y {
        return x
    }
    return y
}
```

Notice that `num` is not a subset of `ord`. This is because the complex number types are not comparable with `<`. To only accept numeric types that are also orderable, combine the two restrictions like this: `type T ord num`.

### Generic types

We've covered everything about generic functions, let's move on to generic types.

To define a generic type, simply list the type parameters in parentheses right after the type name. Like this:

```go
// List is a generic singly-linked list.
type List(type T) struct {
    First T
    Rest  *List(T)
}
```

And as you can already see in the definition, to use a generic type, list the arguments in parentheses after the type name. For example `List(int)` is a list of integers, and `List(string)` is a list of strings.

When defining a type with multiple generic parameters, mark each one with `type`:

```go
// SyncMap is a generic hash-map usable from multiple goroutines simultaneously.
type SyncMap(type K eq, type V) struct {
    mu sync.Mutex
    m  map[K]V
}
```

Methods work as usual:

```go
func (sm *SyncMap(type K eq, type V)) Set(key K, value V) {
    sm.mu.Lock()
    sm.m[key] = value
    sm.mu.Unlock()
}
```

But don't forget that with methods, the `type` keyword is only allowed in the receiver type. For explanation, see [FAQ](#FAQ).

And that's all! Happy hacking!

## FAQ

TODO
