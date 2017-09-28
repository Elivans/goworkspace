package main

import "fmt"
import "smath"
import "os"
import "strconv"

var Usage = func() {
	fmt.Println("USAGE: calc command [arguments] ...")
	fmt.Println("\nThe commands are:\n\tadd\tAddition of two values...")
}

func main() {
	args := os.Args
	fmt.Println("The progarm", args[0], "is running")
	fmt.Println("Length of args is", len(args))
	if args == nil || len(args) < 4 {
		Usage()
		return
	}

	switch args[1] {
	case "add":
		v1, err1 := strconv.Atoi(args[2])
		v2, err2 := strconv.Atoi(args[3])

		if err1 != nil || err2 != nil {
			fmt.Println("USAGE: calc add <integer1><integer2>")
			return
		}
		ret, _ := smath.Add(v1, v2)
		fmt.Println("Result: ", ret)
	case "multi":
		v1, err1 := strconv.Atoi(args[2])
		v2, err2 := strconv.Atoi(args[3])

		if err1 != nil || err2 != nil {
			fmt.Println("USAGE: calc multi <integer1><integer2>")
			return
		}
		ret, _ := smath.Multi(v1, v2)
		fmt.Println("Result: ", ret)
	case "divi":
		v1, err1 := strconv.ParseFloat(args[2], 32)
		v2, err2 := strconv.ParseFloat(args[3], 32)

		if err1 != nil || err2 != nil {
			fmt.Println("USAGE: calc divi <integer1><integer2>")
			return
		}
		red, err := smath.Divi(float32(v1), float32(v2))
		if err == nil {
			fmt.Println("Result: ", red)
		} else {
			fmt.Println(err)
			return
		}
	default:
		Usage()
	}
}
