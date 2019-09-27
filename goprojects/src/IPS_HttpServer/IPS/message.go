// auth
package IPS

import (
	"encoding/json"
)

var Message = map[string]string{
	"9999": "处理失败!",
	"0000": "处理成功!",
	"0001": "报文解析失败!",
	"0002": "未匹配到任何接口报文!",
	"0003": "报文发送失败!",
}

type rtMsg1 struct {
	Code    string
	Message string
}

func code2msg1(code string, otherMsg string) (msg string) {
	rtmsg := make(map[string]string)
	rtmsg["code"] = code
	rtmsg["message"] = Message[code]
	if len(otherMsg) > 0 {
		rtmsg["message"] += "[" + otherMsg + "]"
	}
	bRes, _ := json.Marshal(rtmsg)
	msg = string(bRes)
	return
}
