// 希尔排序（测试未通过）
package shellsort

func ShellSort(buf []int) {
	length := len(buf)
	incre := length
	// fmt.Println("buf: ", buf)
	for {
		incre /= 2
		for k := 0; k < incre; k++ { //根据增量分为若干子序列
			for i := k + incre; i < length; i += incre {
				for j := i; j > k; j -= incre {
					// fmt.Println("j: ", j, " data: ", buf[j], " j-incre: ", j-incre, " data: ", buf[j-incre])
					if buf[j] < buf[j-incre] {
						buf[j-incre], buf[j] = buf[j], buf[j-incre]
					} else {
						break
					}
				}
				// fmt.Println("middle: ", buf)
			}
			// fmt.Println("outer: ", buf)
		}
		// fmt.Println("outer outer: ", buf, " incre: ", incre)

		if incre == 1 {
			break
		}
	}
	// fmt.Println("after: ", buf)
}
