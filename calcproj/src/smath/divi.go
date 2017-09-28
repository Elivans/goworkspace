//一些简单的数学函数
package smath

import "errors"

//将两个整数相除，返回除后的结果
func Divi(a float32, b float32) (ret float32, err error) {
  if b == 0 {
    err = errors.New("Divisor should not be Zero!")
    return
  }
  return a / b, nil
}


