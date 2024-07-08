package main

import "fmt"

type Inf interface {
	Set(int)
}

type A struct {
	num int
}

func (a *A) Set(data int) {
	a.num = data
}

type B struct {
	num int
}

func (b *B) Set(data int) {
	b.num = data
}

type C interface {
	*A | *B

	Inf
}

type D[T C] struct {
	data T
}

func main() {
	d := D[*B]{
		data: &B{},
	}
	d.data.Set(10)

	fmt.Println(d.data.num)
}
