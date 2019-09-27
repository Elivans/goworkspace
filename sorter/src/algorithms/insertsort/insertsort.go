// 插入排序
package insertsort

func InsertSort(buf []int) {
	for i := 1; i < len(buf); i++ {
		for j := i; j > 0; j-- {
			if buf[j] < buf[j-1] {
				buf[j-1], buf[j] = buf[j], buf[j-1]
			} else {
				break
			}
		}
	}
}
