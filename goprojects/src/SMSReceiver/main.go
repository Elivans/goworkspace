// SMSRevcier project main.go
package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	_ "net/http/pprof"

	"github.com/larspensjo/config"
	_ "github.com/mattn/go-oci8"
	"maywide.pkg/env"
	"maywide.pkg/mylogger"
)

type RecData struct {
	Callmdn    string
	Mdn        string
	Content    string
	Reply_time string
	Id         string
}

type RecDatas struct {
	Data []RecData
}

var db *sql.DB
var logger *mylogger.MyLogger

var replyUrl string
var reportUrl string
var confirmUrl string

var c_wait = make(chan int, 1)

//数据库连接初始化
func initParam(cfg *config.Config) {

	url, err := env.Ivalue(cfg, "GDSMS", "REPLYURL")
	if err != nil {
		logger.Println("[faild]read GDSMS REPLYURL faild")
		os.Exit(-1)
	}
	replyUrl = url

	confirmUrl, err = env.Ivalue(cfg, "GDSMS", "CONFIRMURL")
	if err != nil {
		logger.Println("[faild]read GDSMS CONFIRMURL faild")
		os.Exit(-1)
	}

	reportUrl, err = env.Ivalue(cfg, "GDSMS", "REPORTURL")
	if err != nil {
		logger.Println("[faild]read GDSMS REPORTURL faild")
		os.Exit(-1)
	}

	dbstr, err := env.Ivalue(cfg, "DB", "LOGIN")
	if err != nil {
		logger.Println("[faild]read DB LOGIN faild")
		os.Exit(-1)
	}

	db, _ = sql.Open("oci8", dbstr)
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(1)
	err = db.Ping()
	if err != nil {
		logger.Printf("[faild]%s ping faild\n", dbstr)
		os.Exit(-1)
	}

}

func c_end(wait chan int) {
	wait <- 1
}

func process() {
	logger.Println("[begin] process ")
	rows, err := db.Query("select distinct city,companyno, username, password from swp_sms_config")
	if err != nil {
		logger.Println("[faild] query swp_sms_config err")
		os.Exit(-1)
	}
	defer rows.Close()

	nums := 0
	for rows.Next() {
		var city string
		var companyno string
		var username string
		var password string
		if err = rows.Scan(&city, &companyno, &username, &password); err != nil {
			return
		}
		//logger.Printf("geting '%s', '%s', '%s'\n", city, companyno, username, password)
		go httpGetReply(replyUrl, city, "*", companyno, username, password, c_wait) //取回复
		go getReport(reportUrl, companyno, username, password, c_wait)              //取回执
		nums = nums + 1
	}
	//等待所有的完成
	for i := 0; i < nums; i++ {
		//logger.Println("[end] one ")
		<-c_wait
		<-c_wait
	}

	logger.Println("[end] process ")

}

func main() {
	arg_num := len(os.Args)
	if arg_num != 3 {
		fmt.Printf("Usage: %s SMSReceiver.ini stdout\n", os.Args[0])
		os.Exit(-1)
	}

	//打开日志文件
	logger, _ = mylogger.NewLogger(os.Args[2])

	//读取配置文件
	cfg, _ := env.Iputenv(os.Args[1], "ENV")

	logger.Println("SMSReceiver start ...")

	initParam(cfg)

	go func() {
		http.HandleFunc("/goroutines", func(w http.ResponseWriter, r *http.Request) {
			num := strconv.FormatInt(int64(runtime.NumGoroutine()), 10) + "\n" + strconv.FormatInt(int64(db.Stats().OpenConnections), 10) + "\n"
			w.Write([]byte(num))
		})
		http.ListenAndServe("localhost:6060", nil)
		//glog.Info("goroutine stats and pprof listen on 6060")
	}()

	for {
		process() //读取回复写表
		//sendToDT() //扫描数据发到短厅
		time.Sleep(5 * time.Second)
	}
	//	c_wait = make(chan int, 1)
	//	httpGet(replyUrl, "GZ", "*", "210789", "guangzhou_boss", "123456", c_wait)
	//	time.Sleep(1 * time.Second)

	logger.Println("SMSReceiver end ...")
}
