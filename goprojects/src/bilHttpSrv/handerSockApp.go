// handerDBApp
package main

import (
	"io/ioutil"
	"net"
	"net/http"
	"os"

	"github.com/larspensjo/config"
)

var HANDER_INF_SOCKAPP = make(map[string]string)

//Scoket转发类的绑定
func listenerSockApp(cfg *config.Config) {
	section, err := cfg.SectionOptions("SOCK_APP")
	if err == nil {
		for _, v := range section {
			options, err := cfg.String("SOCK_APP", v)
			if err != nil {
				logger.Printf("[faild]read SOCK_APP %s faild\n", v)
				os.Exit(-1)
			}
			HANDER_INF_SOCKAPP[v] = options
			http.HandleFunc(v, HandlerSockApp)
		}
	}

}

//请求处理
func HandlerSockApp(w http.ResponseWriter, req *http.Request) {
	mylogger := logger.CloneLogger("HandlerSockApp")

	defer req.Body.Close()
	reqdata, _ := ioutil.ReadAll(req.Body) //请求数据

	path := req.URL.Path
	host := req.RemoteAddr
	addr := HANDER_INF_SOCKAPP[path]

	mylogger.Printf("reqpath=%s", path)
	mylogger.Printf("reqip=%s", host)
	mylogger.Printf("conn to=%s", addr)
	mylogger.Printf("request data=%s", string(reqdata))

	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		mylogger.Println("socket faild")
		mylogger.Printf("[faild]ResolveTCPAddr err \n%s\n", err.Error())
		w.Write([]byte("socket faild"))
		return
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		mylogger.Println("socket faild")
		mylogger.Printf("[faild]DialTCP err\n%s\n", err.Error())
		w.Write([]byte("socket faild"))
		return
	}

	conn.Write([]byte(reqdata))
	var resbody []byte
	conn.Read(resbody)
	mylogger.Printf("res data=%s", string(resbody))

	w.Write(resbody)
}
