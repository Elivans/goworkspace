// hello.go project main.go
package main

import (
	"io"
	//	"log"
	//	"io/ioutil"
	"net/http"

	"maywide.pkg/database"
	//"maywide.pkg/database/pool"

	"encoding/json"

	//	"strings"

	//	"fmt"

	mylog "maywide.pkg/mylogger"
)

type Request struct {
	Custid string `json:"custid"`
}

type ServidArrear struct {
	Servid string `json:"servid"`
	Fees   string `json:"fees"`
}

type Response struct {
	Retcode string         `json:"retcode"`
	Retmsg  string         `json:"retmsg"`
	Rrrear  []ServidArrear `json:"arrear"`
}

var res Response

var resmsg []byte

func goarrearHandler(w http.ResponseWriter, r *http.Request) {
	//	body, _ := ioutil.ReadAll(r.Body)

	//	body_str := string(body)
	//	mylog.Println(body_str)

	//	var req Request
	//	err := json.Unmarshal(body, &req)
	//	if err != nil {
	//		mylog.Println(err)
	//	}
	//	mylog.Println(req)
	//	resmsg := getFees(req.Custid)
	resmsg := getFees(string("2234565"))

	io.WriteString(w, string(resmsg))
}

func getFees(custid string) []byte {
	DSN := "boss_crm/boss_crm@10.205.28.28:1521/boss"
	database.Driver = "oracle"
	database.Addr = "10.205.28.86:9527"
	database.Show = true

	mylog.Println(DSN)

	db := database.Connect(DSN, 3)

	mylog.Println("connect db ", db.Sid, db.Sqlcode, db.Sqlerrm)

	sqlstr := "SELECT SERVID,SUM(FEES) FEES FROM BOSS_BIL.BIL_ARREAR_TOTAL WHERE CUSTID = " + custid + " GROUP BY SERVID"

	getmsg, err := db.SelectMap(sqlstr)
	if err != nil {
		mylog.Println("select mcode: ", err.Error())
		res.Retcode = "0001"
		res.Retmsg = "返回失败"
	}

	mylog.Println("MAP:", getmsg)

	arrear := make([]ServidArrear, 0)

	var varr ServidArrear

	for key, value := range getmsg {
		varr.Servid = key
		varr.Fees = value
		arrear = append(arrear, varr)
	}
	if len(arrear) == 0 {
		res.Retcode = "0002"
		res.Retmsg = "查无数据"
	} else {
		res.Retcode = "0000"
		res.Retmsg = "返回成功"
	}

	res.Rrrear = arrear

	resmsg, err = json.Marshal(res)

	if err != nil {
		mylog.Println("JSON: ", err.Error())
	}

	mylog.Println("JSON:", string(resmsg))
	return resmsg
}
func main() {
	mylog.NewLogger("stdout")

	http.HandleFunc("/", goarrearHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		mylog.Println("ListenAndServe: ", err.Error())
	}
}
