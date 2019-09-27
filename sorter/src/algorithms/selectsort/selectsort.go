// 选择排序
package selectsort

func SelectSort(buf []int) {
	for i := 0; i < len(buf)-1; i++ {
		min := i
		for j := i; j < len(buf); j++ {
			if buf[min] > buf[j] {
				min = j
			}
		}
		if min != i {
			buf[i], buf[min] = buf[min], buf[i]
		}
	}
}
