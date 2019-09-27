package main

import "fmt"

import "time"

func main() {
	ch := make(chan int, 1)
	//	for i := 0; i < 10; i++ {
	//		select {
	//		case ch <- 0:
	//		case ch <- 1:
	//		}
	//		cc := <-ch
	//		fmt.Println("Value received:", i, cc)
	//	}

	//匿名超时等待函数
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(10e9) //等待1秒钟
		timeout <- true
	}()

	select {
	case <-ch:
		//	从ch中读取到数据
	case <-timeout:
		//一直没有从ch中读到数据，但从timeout中读到
		fmt.Println("Value received timeout")
	}

}
