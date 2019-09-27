package main

import "fmt"

func Count(ch chan<- int, in int) {
	ch <- in
	fmt.Println("Counting", in)
}

func main() {
	chs := make([]chan int, 10)
	for i := 0; i < 10; i++ {
		chs[i] = make(chan int, 1024)
		go Count(chs[i], i)
	}
	for j, ch := range chs {
		val := <-ch
		fmt.Println("chs[", j, "]=", val)
	}

	ch := make(chan int, 1024) //初始化ch，并分配1024缓冲区
	for i := 0; i < 10; i++ {
		select { //随机向ch写入0或者1
		case ch <- 0:
		case ch <- 1:
		}
		j := <-ch
		fmt.Println(i, "Value received:", j)
	}
	ch1 := make(chan int, 4)
	fmt.Println("The cap of ch1:", cap(ch1))
	fmt.Println("The len of ch1:", len(ch1))

	ch1 <- 12
	ch1 <- 34
	ch1 <- 2
	fmt.Println("The len of ch1:", len(ch1))

	//	for {
	//		if 0 == len(ch1) {
	//			break
	//		}
	//		fmt.Println("The ch1:", <-ch1)
	//	}

	for k := range ch1 {
		fmt.Println("The ch1:", k)
		if 0 == len(ch1) {
			break
		}
	}
	close(ch1)

}
