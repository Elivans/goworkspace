//归并排序
package mergesort

func MergeSort(buf []int) {
	tmp := make([]int, len(buf))
	mergeSort(buf, 0, len(buf)-1, tmp)
}

func mergeSort(buf []int, first, last int, tmp []int) {
	if first < last {
		middle := (first + last) / 2
		mergeSort(buf, first, middle, tmp)        //左半部分排好序
		mergeSort(buf, middle+1, last, tmp)       //右半部分排好序
		mergeArray(buf, first, middle, last, tmp) //合并左右部分
	}
}

func mergeArray(buf []int, first, middle, end int, tmp []int) {
	// fmt.Printf("mergeArray buf: %v, first: %v, middle: %v, end: %v, tmp: %v\n",
	//     buf, first, middle, end, tmp)
	i, m, j, n, k := first, middle, middle+1, end, 0
	for i <= m && j <= n {
		if buf[i] <= buf[j] {
			tmp[k] = buf[i]
			k++
			i++
		} else {
			tmp[k] = buf[j]
			k++
			j++
		}
	}
	for i <= m {
		tmp[k] = buf[i]
		k++
		i++
	}
	for j <= n {
		tmp[k] = buf[j]
		k++
		j++
	}

	for ii := 0; ii < k; ii++ {
		buf[first+ii] = tmp[ii]
	}
	// fmt.Printf("sort: buf: %v\n", buf)
}
