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
读取文件；
将读文件的内容，调用用户 Map 函数，生产对于的 KeyValue 值；
最后按照 KeyValue 里面的 Key 进行分区，将内容写入到文件里面，以便于后面的 Reduce 过程执行；
*/
func doMap(
	jobName string, // // the name of the MapReduce job
	mapTaskNumber int, // which map task this is
	inFile string,
	nReduce int, // the number of reduce task that will be run
	mapF func(file string, contents string) []KeyValue,
) {

	//setp 1 read file
	contents, err := ioutil.ReadFile(inFile)
	if err != nil {
		log.Fatal("do map error for inFile ", err)
	}
	//setp 2 call user user-map method ,to get kv
	kvResult := mapF(inFile, string(contents))
	/**
	 *   setp 3 use key of kv generator nReduce file ,partition
	 *      a. create tmpFiles
	 *      b. create encoder for tmpFile to write contents
	 *      c. partition by key, then write tmpFile
	 */
	var tmpFiles []*os.File = make([]*os.File, nReduce)
	var encoders []*json.Encoder = make([]*json.Encoder, nReduce)
	for i := 0; i < nReduce; i++ {
		tmpFileName := reduceName(jobName, mapTaskNumber, i)
		tmpFiles[i], err = os.Create(tmpFileName)
		if err != nil {
			log.Fatal(err)
		}
		defer tmpFiles[i].Close()
		encoders[i] = json.NewEncoder(tmpFiles[i])
		if err != nil {
			log.Fatal(err)
		}
	}
	for _, kv := range kvResult {
		hashKey := int(ihash(kv.Key)) % nReduce
		err := encoders[hashKey].Encode(&kv)
		if err != nil {
			log.Fatal("do map encoders ", err)
		}
	}
}
