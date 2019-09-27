package main

import (
	//"crypto/rand"
	"crypto/tls"
	"io"
	"log"
	//"net"
	//"time"
)




/*
读取 doMap 过程中产生的中间文件；
按照读取相同文件中的 Key 进新按照字典顺序进行排序；
遍历读取的 KeyValue，并且调用用户的 Reduce 方法，将计算的结果继续写入到文件中；
*/
func doReduce(
	jobName string, // the name of the whole MapReduce job
	reduceTaskNumber int, // which reduce task this is
	nMap int, // the number of map tasks that were run ("M" in the paper)
	reduceF func(key string, values []string) string,
) {
	
	// file.Close()
	//setp 1,read map generator file ,same key merge put map[string][]string
	kvs := make(map[string][]string)
	for i := 0; i < nMap; i++ {
		fileName := reduceName(jobName, i, reduceTaskNumber)
		file, err := os.Open(fileName)
		if err != nil {
			log.Fatal("doReduce1: ", err)
		}
		dec := json.NewDecoder(file)
		for {
			var kv KeyValue
			err = dec.Decode(&kv)
			if err != nil {
				break
			}
			_, ok := kvs[kv.Key]
			if !ok {
				kvs[kv.Key] = []string{}
			}
			kvs[kv.Key] = append(kvs[kv.Key], kv.Value)
		}
		file.Close()
	}
	var keys []string
	for k := range kvs {
		keys = append(keys, k)
	}
	//setp 2 sort by keys
	sort.Strings(keys)
	//setp 3 create result file
	p := mergeName(jobName, reduceTaskNumber)
	file, err := os.Create(p)
	if err != nil {
		log.Fatal("doReduce2: ceate ", err)
	}
	enc := json.NewEncoder(file)
	//setp 4 call user reduce each key of kvs
	for _, k := range keys {
		res := reduceF(k, kvs[k])
		enc.Encode(KeyValue{k, res})
	}
	file.Close()
}