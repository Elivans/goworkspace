// main.go
package main

import (
	"fmt"

	"maywide.pkg/expr"
)

func main() {
	err, isok := expr.Expr_value("(abd=abd)&&(1=1)")
	if err != nil {
		fmt.Println(err)
		return
	}
	if isok {
		fmt.Println("条件成立")
	} else {
		fmt.Println("条件不成立")
	}
}
