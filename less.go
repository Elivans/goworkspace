package main

import "fmt"

type Integer int

func (a Integer) Less(b Integer) bool {
	return a < b
}

func (a *Integer) Add(b Integer) {
	*a += b
	//fmt.Println("func add a =",a)
}

type LessAdder interface {
	Less(b Integer) bool
	Add(b Integer)
}

func main() {
	var a Integer = 1
	if a.Less(2) {
		fmt.Println(a, "less 2")
	}
	a.Add(4)
	fmt.Println("a =", a)

	var b LessAdder = &a
	//var b LessAdder = a
	if b.Less(2) {
		fmt.Println(a, "less 2")
	} else {
		fmt.Println(a, "bigger 2")
	}

}
