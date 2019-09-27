// handerDBApp
package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/larspensjo/config"
)

var HANDER_INF_CMDAPP = make(map[string]string)

//执行本地命令的绑定
func listenerCmdApp(cfg *config.Config) {
	section, err := cfg.SectionOptions("CMD_APP")
	if err == nil {
		for _, v := range section {
			options, err := cfg.String("CMD_APP", v)
			if err != nil {
				logger.Printf("[faild]read CMD_APP %s faild\n", v)
				os.Exit(-1)
			}
			HANDER_INF_CMDAPP[v] = options
			http.HandleFunc(v, HandlerCmdApp)
		}
	}

}

//请求处理
func HandlerCmdApp(w http.ResponseWriter, req *http.Request) {
	mylogger := logger.CloneLogger("HandlerCmdApp")

	defer req.Body.Close()
	reqdata, _ := ioutil.ReadAll(req.Body) //请求数据

	path := req.URL.Path
	host := req.RemoteAddr
	cmdstr := HANDER_INF_CMDAPP[path]

	mylogger.Printf("reqpath=%s", path)
	mylogger.Printf("reqip=%s", host)
	mylogger.Printf("cmd=%s", HANDER_INF_CMDAPP[path])
	mylogger.Printf("request data=%s", string(reqdata))
	var tmpstr string
	if len(strings.TrimSpace(string(reqdata))) > 0 {
		tmpstr = cmdstr + " " + string(reqdata)
	} else {
		tmpstr = cmdstr
	}
	params := strings.Split(tmpstr, " ")
	cmdstr = params[0]

	mylogger.Printf("cmdstr=%s= data=%v", cmdstr, len(params))

	tmpstr = ""
	for i := 1; i < len(params); i++ {
		if i == 1 {
			tmpstr = params[i]
		} else {
			tmpstr = tmpstr + " " + params[i]
		}
	}
	params = strings.Split(strings.TrimSpace(tmpstr), " ")
	var cmd *exec.Cmd
	if len(params[0]) > 0 {
		cmd = exec.Command(cmdstr, params...)
	} else {
		cmd = exec.Command(cmdstr)
	}
	mylogger.Printf("exec=%v", cmd.Args)
	buf, err := cmd.Output()
	if err != nil {
		mylogger.Printf("%v", err.Error())
		w.Write([]byte("cmd faild"))
	}
	mylogger.Printf("res data=%s", string(buf))

	w.Write(buf)
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
