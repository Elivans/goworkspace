// 堆排序(测试未通过)
package heapsort
/*
func HeapSort(buf []int) {
	n := len(buf)

	for i := (n - 1) / 2; i >= 0; i-- {
		minHeapFixdown(buf, i, n)
	}

	for i := n - 1; i > 0; i-- {
		buf[0], buf[i] = buf[i], buf[0]
		minHeapFixdown(buf, 0, i)
	}
}
*/
func minHeapFixdown(buf []int, i, n int) {
	j := 2*i + 1
	for j < n {
		if j+1 < n && buf[j+1] < buf[j] {
			j++
		}

		if buf[i] <= buf[j] {
			break
		}
		buf[i], buf[j] = buf[j], buf[i]

		i = j
		j = 2*i + 1
	}
}


// 堆排序
func HeapSort(buf []int) {
    temp, n := 0, len(buf)

    for i := (n - 1) / 2; i >= 0; i-- {
        minHeapFixdown(buf, i, n)
    }

    for i := n - 1; i > 0; i-- {
        temp = buf[0]
        buf[0] = buf[i]
        buf[i] = temp
        minHeapFixdown(buf, 0, i)
    }
}
