package main

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
)

func main() {
	TestFile := "123.txt"
	infile, inerr := os.Open(TestFile)
	if inerr == nil {
		Md5Inst := md5.New()
		io.Copy(Md5Inst, infile)
		Result := Md5Inst.Sum([]byte(""))
		fmt.Printf("%x %s\n", Result, TestFile)

		Sha1Inst := sha1.New()
		io.Copy(Sha1Inst, infile)
		Result = Sha1Inst.Sum([]byte(""))
		fmt.Printf("%x %s\n", Result, TestFile)
	} else {
		fmt.Println(inerr)
		os.Exit(1)
	}

}
