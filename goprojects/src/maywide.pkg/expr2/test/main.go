// main.go
package main

import (
	"fmt"
	"strings"

	"maywide.pkg/expr2"
)

func f1(s int) int {
	return s + 1
}
func main3() {
	fmt.Println(strings.Replace("112233", "2", "a", 1))
}

func main() {

	e, err := expr2.MustCompile2(`""=="FSDP_ZX"`)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(e.String())
	fmt.Println(e.Bool(expr2.V{}))
}

func main2() {

	e, err := expr2.MustCompile(`"11" == 11`)
	fmt.Println(e.String())
	fmt.Println(e.Bool(expr2.V{"$v": "vv", "b": 2}))

	e, err = expr2.MustCompile("3  > 2")
	if err != nil {
		fmt.Println(err)
		return
	}
	bb := e.Bool(expr2.V{})
	fmt.Println("..........")
	fmt.Println(bb)
	fmt.Println("..........")

	e, _ = expr2.MustCompile2(`""!=" "`)
	fmt.Println(e.String())
	n, _ := e.Eval(expr2.V{})

	//n, _ := e.Eval(expr2.V{"a": 1})
	fmt.Println(n)

}
