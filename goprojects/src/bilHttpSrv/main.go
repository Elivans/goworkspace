// main
package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"maywide.pkg/env"
	"maywide.pkg/mylogger"

	"github.com/larspensjo/config"
	_ "github.com/mattn/go-oci8"
)

var logger *mylogger.MyLogger

type response struct {
	Rtcode    string
	Rtmessage string
}

type handerfun struct {
	Funname string
	Dblink  string
}

var DB_LINK = make(map[string]*sql.DB)

var TimeOut int64

//数据库连接初始化
func initDB(cfg *config.Config) {
	dbnls, err := env.Ivalue(cfg, "APP_ENV", "NLS_LANG")
	if err != nil {
		logger.Println("[faild]read DB_ENV NLS_LANG faild")
		os.Exit(-1)
	}
	os.Setenv("NLS_LANG", dbnls)

	section, err := cfg.SectionOptions("DB_LNK")
	if err == nil {
		for _, v := range section {
			dbstr, err := env.Ivalue(cfg, "DB_LNK", v)
			if err != nil {
				logger.Printf("[faild]read DB %s faild\n", v)
				os.Exit(-1)
			}

			sMaxOpenConns, err := env.Ivalue(cfg, "DB_LIMIT", v+"_MAXOPEN")
			if err != nil {
				logger.Printf("[faild]read DB_LIMIT %s faild\n", v+"_MAXOPEN")
				os.Exit(-1)
			}
			sMaxIdleConns, err := env.Ivalue(cfg, "DB_LIMIT", v+"_MAXIDLE")
			if err != nil {
				logger.Printf("[faild]read DB_LIMIT %s faild\n", v+"_MAXIDLE")
				os.Exit(-1)
			}
			iMaxOpenConns, err := strconv.Atoi(sMaxOpenConns)
			if err != nil {
				logger.Printf("[faild]%s MaxOpenConns is not int\n", sMaxOpenConns)
				os.Exit(-1)
			}
			iMaxIdleConns, err := strconv.Atoi(sMaxIdleConns)
			if err != nil {
				logger.Printf("[faild]%s MaxIdleConns is not int\n", sMaxIdleConns)
				os.Exit(-1)
			}
			db, _ := sql.Open("oci8", dbstr)
			db.SetMaxOpenConns(iMaxOpenConns)
			db.SetMaxIdleConns(iMaxIdleConns)
			DB_LINK[v] = db
			err = DB_LINK[v].Ping()
			if err != nil {
				logger.Printf("[faild]%s ping faild\n", v)
				os.Exit(-1)
			}
		}
	}

}

//对配置文件中INT_APP下的配置进行监听
func listener(cfg *config.Config) {
	//服务监听端口
	port, err := cfg.String("APP_ENV", "PORT")
	if err != nil {
		logger.Println("[faild]read APP_ENV PORT faild")
		os.Exit(-1)
	}
	ipport := ":" + port

	//超时设置
	rtimout, err := cfg.String("APP_ENV", "READ_TIMEOUT")
	if err != nil {
		logger.Println("[faild]read APP_ENV READ_TIMEOUT faild")
		os.Exit(-1)
	}
	wtimout, err := cfg.String("APP_ENV", "WRITE_TIMEOUT")
	if err != nil {
		logger.Println("[faild]read APP_ENV WRITE_TIMEOUT faild")
		os.Exit(-1)
	}
	irtimeout, _ := strconv.ParseInt(rtimout, 10, 64)
	iwtimeout, _ := strconv.ParseInt(wtimout, 10, 64)
	//irwtimeout := irtimeout + iwtimeout
	TimeOut = iwtimeout

	srv := &http.Server{
		Addr:         ipport,
		Handler:      http.TimeoutHandler(http.DefaultServeMux, time.Second*time.Duration(iwtimeout-1), "server timeout"),
		ReadTimeout:  time.Duration(irtimeout) * time.Second,
		WriteTimeout: time.Duration(iwtimeout) * time.Second,
	}

	//数据库操作类的绑定
	listenerDBApp(cfg)
	//HTTP转发类的绑定
	listenerHttpApp(cfg)
	//执行本地命令的绑定
	listenerCmdApp(cfg)
	//普通函数绑定
	listenerCommApp(cfg)
	//Socket转发绑定
	listenerSockApp(cfg)

	logger.Println(srv.ListenAndServe())
	//http.ListenAndServe(ipport, nil) //开始监听
}

func main() {

	arg_num := len(os.Args)
	if arg_num != 3 {
		fmt.Printf("Usage: %s bilHttpSrv.ini stdout\n", os.Args[0])
		os.Exit(-1)
	}

	//打开日志文件
	logger, _ = mylogger.NewLogger(os.Args[2])

	//读取配置文件
	cfg, _ := env.Iputenv(os.Args[1], "ENV")

	logger.Println("bossDB start ...")

	BindFun()
	initDB(cfg)
	listener(cfg)

}
