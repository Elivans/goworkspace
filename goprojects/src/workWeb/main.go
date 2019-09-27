package main

import (
	_ "workWeb/routers"

	"github.com/astaxie/beego"
	mylog "maywide.pkg/mylogger"
)

func main() {
	mylog.NewLogger("stdout")
	beego.BConfig.WebConfig.Session.SessionOn = true
	beego.BConfig.WebConfig.Session.SessionGCMaxLifetime = 300
	beego.BConfig.WebConfig.Session.SessionCookieLifeTime = 300
	beego.Run()
}
