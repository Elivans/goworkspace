package main

import (
	"fmt"
	"runtime"
)

//多核并行计算
//计算N个整数的总和

type Vector []float64

//分配给每个CPU的计算任务
func (v Vector) DoSome(i, n int, u Vector, c chan int) {
	for ; i < n; i++ {
		u[i/NCPU] += v[i]
	}
	c <- 1 //发信号告诉任务管理者我已经计算完成了
}

//假设共有4核
const NCPU = 4

//const NCPU := runtime.NumCPU()

func (v Vector) DoAll(u Vector) {
	c := make(chan int, NCPU) //接收每个CPU的任务完成信号

	runtime.GOMAXPROCS(NCPU) //设置使用4个CPU核心

	for i := 0; i < NCPU; i++ {
		go v.DoSome(i*len(v)/NCPU, (i+1)*len(v)/NCPU, u, c)
	}
	//等待所有CPU的任务完成
	for i := 0; i < NCPU; i++ {
		<-c
	}
}

func main() {
	var myv Vector = Vector{123.45, 65, 32, 54, 5, 4, 3, 333.22, 7,
		6, 12, 456.8, 54, 9000, 5444, 232}
	var myv1 Vector
	myv.DoAll(myv1)
	for _, i := range myv1 {
		fmt.Println(i)
	}

}
