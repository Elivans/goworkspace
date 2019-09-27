// handerDBApp
package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/larspensjo/config"
)

var HANDER_INF_HTTPAPP = make(map[string]string)

//HTTP转发类的绑定
func listenerHttpApp(cfg *config.Config) {
	section, err := cfg.SectionOptions("HTTP_APP")
	if err == nil {
		for _, v := range section {
			options, err := cfg.String("HTTP_APP", v)
			if err != nil {
				logger.Printf("[faild]read HTTP_APP %s faild\n", v)
				os.Exit(-1)
			}
			HANDER_INF_HTTPAPP[v] = options
			http.HandleFunc(v, HandlerHttpApp)
		}
	}

}

//请求处理
func HandlerHttpApp(w http.ResponseWriter, req *http.Request) {
	mylogger := logger.CloneLogger("HandlerHttpApp")

	defer req.Body.Close()
	reqdata, _ := ioutil.ReadAll(req.Body) //请求数据

	path := req.URL.Path
	host := req.RemoteAddr
	addr := HANDER_INF_HTTPAPP[path]

	mylogger.Printf("reqpath=%s", path)
	mylogger.Printf("reqip=%s", host)
	mylogger.Printf("post to=%s", addr)
	mylogger.Printf("request data=%s", string(reqdata))

	resp, err := http.Post(addr,
		"application/x-www-form-urlencoded",
		bytes.NewReader(reqdata))
	if err != nil {
		mylogger.Println("post faild")
		mylogger.Printf("post to %s err \n%s\n", addr, err.Error())
		w.Write([]byte("post faild"))
		return
	}

	defer resp.Body.Close()

	resbody, _ := ioutil.ReadAll(resp.Body)

	mylogger.Printf("res data=%s", string(resbody))

	w.Write(resbody)
}
