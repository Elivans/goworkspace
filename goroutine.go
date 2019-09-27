package main

import (
	"fmt"
	// "time"
)

var c chan int

func Add(x, y int) {
	z := x + y
	fmt.Println(z)
	c <- 1
}

func sum(values []int, resultChan chan int) {
	sum := 0
	for _, value := range values {
		sum += value
	}
	resultChan <- sum
}

func main() {
	values := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	resultChan := make(chan int, 2)
	go sum(values[:len(values)/2], resultChan)
	go sum(values[len(values)/2:], resultChan)
	sum1, sum2 := <-resultChan, <-resultChan
	fmt.Println("Result:", sum1, sum2, sum1+sum2)

	c = make(chan int)
	for i := 0; i < 3; i++ {
		go Add(1, i)
		go Add(1000, i)
		//time.Sleep(1000000000)
	}
	j := 0
L:
	for {
		select {
		case <-c:
			j++
			if j > 1 {
				break L
			}
		}
	}
}
