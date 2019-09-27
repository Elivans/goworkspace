// handerDBApp
package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/larspensjo/config"
)

var HANDER_INF_DBAPP = make(map[string]handerfun)

//数据库操作类的绑定
func listenerDBApp(cfg *config.Config) {
	section, err := cfg.SectionOptions("DB_APP")
	if err == nil {
		for _, v := range section {
			options, err := cfg.String("DB_APP", v)
			if err != nil {
				logger.Printf("[faild]read DB_APP %s faild\n", v)
				os.Exit(-1)
			}
			s := strings.Split(options, ",")
			var hfun handerfun
			hfun.Funname = s[0]
			hfun.Dblink = s[1]
			HANDER_INF_DBAPP[v] = hfun
			http.HandleFunc(v, HandlerDBApp)
		}
	}
}

//请求处理
func HandlerDBApp(w http.ResponseWriter, req *http.Request) {
	mylogger := logger.CloneLogger("HandlerDBApp")

	var resdata string //返回串

	defer req.Body.Close()
	reqdata, _ := ioutil.ReadAll(req.Body) //请求数据

	path := req.URL.Path

	var res response
	res.Rtcode = "000"
	res.Rtmessage = "处理成功"

	funname := HANDER_INF_DBAPP[path].Funname
	db := DB_LINK[HANDER_INF_DBAPP[path].Dblink]

	mylogger.Printf("reqpath=%s", path)
	mylogger.Printf("funname=%s", funname)
	mylogger.Printf("use db=%s", HANDER_INF_DBAPP[path].Dblink)
	mylogger.Printf("request data=%s", string(reqdata))

	tx, _ := db.Begin()
	result, err := funcs.Call(funname, tx, mylogger, string(reqdata)) //调用具体处理函数
	if err != nil {
		logger.Println("%v", err.Error())
		res.Rtcode = "EEE"
		res.Rtmessage = err.Error()
		bResdata, _ := json.Marshal(res)
		resdata = string(bResdata)
		tx.Rollback()
	} else {
		resdata = result[0].Interface().(string) //断言

		if result[1].Interface() != nil {
			logger.Println("call func faild")
			err = result[1].Interface().(error) //断言
			res.Rtcode = "EEE"
			res.Rtmessage = err.Error()
			bResdata, _ := json.Marshal(res)
			resdata = string(bResdata)
			tx.Rollback()
		}

	}

	mylogger.Printf("respons data=%s", resdata)
	_, err = w.Write([]byte(resdata))
	if err == http.ErrHandlerTimeout { //客户端离开了
		mylogger.Println("操作超时了")
		tx.Rollback()
	}
	tx.Commit()

	mylogger.Println("=============end!")
}
