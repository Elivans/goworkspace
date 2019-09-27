// handerDBApp
package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/larspensjo/config"
)

var HANDER_INF_COMMAPP = make(map[string]string)

//执行本地函数的绑定
func listenerCommApp(cfg *config.Config) {
	section, err := cfg.SectionOptions("COMM_APP")
	if err == nil {
		for _, v := range section {
			options, err := cfg.String("COMM_APP", v)
			if err != nil {
				logger.Printf("[faild]read COMM_APP %s faild\n", v)
				os.Exit(-1)
			}
			HANDER_INF_COMMAPP[v] = options
			http.HandleFunc(v, HandlerCommApp)
		}
	}

}

//请求处理
func HandlerCommApp(w http.ResponseWriter, req *http.Request) {
	mylogger := logger.CloneLogger("HandlerCommApp")

	var resdata string //返回串

	defer req.Body.Close()
	reqdata, _ := ioutil.ReadAll(req.Body) //请求数据

	path := req.URL.Path

	var res response
	res.Rtcode = "000"
	res.Rtmessage = "处理成功"

	funname := HANDER_INF_COMMAPP[path]

	mylogger.Printf("reqpath=%s", path)
	mylogger.Printf("funname=%s", funname)
	mylogger.Printf("request data=%s", string(reqdata))

	result, err := funcs.Call(funname, mylogger, string(reqdata)) //调用具体处理函数
	if err != nil {
		mylogger.Printf("[faild] %v", err.Error())
		res.Rtcode = "EEE"
		res.Rtmessage = err.Error()
		bResdata, _ := json.Marshal(res)
		resdata = string(bResdata)
	} else {
		resdata = result[0].Interface().(string) //断言

	}

	mylogger.Printf("respons data=%v", resdata)
	_, err = w.Write([]byte(resdata))
	if err == http.ErrHandlerTimeout {
		mylogger.Println("操作超时了")
	}
}

//func main() {

//	cmd := exec.Command("C:\\Windows\\System32\\cmd.exe")
//	fmt.Println(cmd.Args)
//	buf, err := cmd.Output()

//	if err != nil {
//		fmt.Fprintf(os.Stderr, err.Error())
//		return
//	}
//	fmt.Fprintf(os.Stdout, "Result: %s", buf)
//}
