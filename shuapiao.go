package main

import (
	"fmt"
	"net/http"
)

//http://live.sports.ifeng.com/match/552.html
func vote(a chan int) {
	for i := 0; i <= 1000; i++ {
		http.Get("http://survey.news.ifeng.com/accumulator_ext.php?callback=jQuery1820030119983945041895_1490671752116&key=customLiveaway_support_552&format=js&_=1490671777810")
	}
	a <- 0
}

func main() {
	a := make(chan int, 30)
	for i := 0; i < 30; i++ {
		go vote(a)
	}
	for b := range a {
		fmt.Println(b)
	}
}
