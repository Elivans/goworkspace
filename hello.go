package main

import (
	"fmt"
	"runtime"
	"sync"
)

var counter int = 0

func Count(lock *sync.Mutex) {
	lock.Lock()
	counter++
	fmt.Println(counter)
	lock.Unlock()
}

func main() {

	//字符串
	var str string
	str = "hello world"
	fmt.Printf("the length of \"%s\" is %d\n", str, len(str))

	//定义容量为10,初始长度为5的整型数组
	mySlice := make([]int, 5, 10)
	fmt.Println("len(mySlice):", len(mySlice))
	fmt.Println("cap(mySlice):", cap(mySlice))
	for i := 0; i < len(mySlice); i++ {
		fmt.Println("mySlice[", i, "]=", mySlice[i])
	}

	//追加指为1的元素到数组末尾
	mySlice = append(mySlice, 1)
	fmt.Println("len(mySlice):", len(mySlice))
	fmt.Println("cap(mySlice):", cap(mySlice))
	//用range方式遍历数组
	for i, v := range mySlice {
		fmt.Println("mySlice[", i, "]=", v)
	}

	mySlice2 := []int{21, 22, 23, 24, 25, 26, 27, 28, 29, 30}
	//mySlice = append(mySlice,mySlice2...)
	//复制数组
	copy(mySlice, mySlice2)
	fmt.Println("len(mySlice):", len(mySlice))
	fmt.Println("cap(mySlice):", cap(mySlice))
	for i, v := range mySlice {
		fmt.Println("mySlice[", i, "]=", v)
	}

	lock := &sync.Mutex{}
	for i := 0; i < 10; i++ {
		go Count(lock)
	}
	for {
		lock.Lock()
		c := counter
		lock.Unlock()
		runtime.Gosched()
		if c >= 10 {
			break
		}
	}
}
