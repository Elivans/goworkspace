package main

import "flag"
import "fmt"
import "bufio"
import "io"
import "os"
import "strconv"
import "time"
import "math/rand"

import "algorithms/bubblesort"
import "algorithms/qsort"
import "algorithms/heapsort"
import "algorithms/insertsort"
import "algorithms/mergesort"
import "algorithms/selectsort"
import "algorithms/shellsort"

var infile *string = flag.String("i", "infile", "File contains values for sorting")
var outfile *string = flag.String("o", "outfile", "File to receive sorted values")
var algorithm *string = flag.String("a", "qsort", "Sort algorithm")

//逐行读取文件内容，并解析为int类型数据，再添加到int类型的数组切片中
func readValues(infile string) (values []int, err error) {
	file, err := os.Open(infile)
	if err != nil {
		fmt.Println("Failed to open the input file ", infile)
		return
	}
	defer file.Close()

	br := bufio.NewReader(file)
	values = make([]int, 0)

	for {
		line, isPrefix, err1 := br.ReadLine()

		if err1 != nil {
			if err1 != io.EOF {
				err = err1
			}
			break
		}

		if isPrefix {
			fmt.Println("A too long line, seems unexpected.")
			return
		}

		str := string(line)
		value, err1 := strconv.Atoi(str)

		if err1 != nil {
			err = err1
			return
		}

		values = append(values, value)
	}
	return
}

//排序后的数组切片写入文件
func writeValues(values []int, outfile string) error {
	file, err := os.Create(outfile)
	if err != nil {
		fmt.Println("Failed to create the output file ", outfile)
		return err
	}
	defer file.Close()

	for _, value := range values {
		str := strconv.Itoa(value)
		file.WriteString(str + "\n")
	}
	return nil
}

func main() {
	//获取并解析命令行输入
	flag.Parse()

	if infile != nil {
		fmt.Println("infile =", *infile, "outfile =", *outfile, "algorithm =",
			*algorithm)
	}

	values, err := readValues(*infile)

	const (
		num      = 100000
		rangeNum = 100000
	)
	randSeed := rand.New(rand.NewSource(time.Now().Unix() + time.Now().UnixNano()))
	for i := 0; i < num; i++ {
		values = append(values, randSeed.Intn(rangeNum))
	}

	if err == nil {
		fmt.Println("Values(before sort):", values)
		t1 := time.Now()
		switch *algorithm {
		case "qsort":
			qsort.QuickSort(values)
		case "bubblesort":
			bubblesort.BubbleSort(values)
		case "heapsort":
			heapsort.HeapSort(values)
		case "insertsort":
			insertsort.InsertSort(values)
		case "mergesort":
			mergesort.MergeSort(values)
		case "selectsort":
			selectsort.SelectSort(values)
		case "shellsort":
			shellsort.ShellSort(values)
		default:
			fmt.Println("Sorting algorithm", *algorithm,
				"is either unknown or unsupported.")
		}

		//t2 := time.Now()

		//fmt.Println("Values(after sort):", values)
		//fmt.Println("The sorting process costs", t2.Sub(t1), "to complete.")
		fmt.Println("The sorting process costs", time.Since(t1), "to complete.")
		
		writeValues(values, *outfile)

	} else {
		fmt.Println(err)
	}
}
