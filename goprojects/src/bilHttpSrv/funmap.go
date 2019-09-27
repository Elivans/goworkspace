// funmap
package main

import (
	"bilHttpSrv/interface_inter/example_int"
)

var funmap = map[string]interface{}{
	"test":     example_int.TestDB,
	"test2":    example_int.TestDB2,
	"testcomm": example_int.TestComm,
}
