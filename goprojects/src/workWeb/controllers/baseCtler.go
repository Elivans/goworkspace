package controllers

import (
	"fmt"

	"workWeb/db"

	"github.com/astaxie/beego"
)

type BaseController struct {
	beego.Controller
	User      string
	Name      string
	UseParams []string `使用到的参数列表`
}

func (this *BaseController) Params() {
	for _, gcode := range this.UseParams {
		params, _ := db.GetParam(gcode)
		htmlvalue := ""
		for _, param := range params {
			htmlvalue += `<option value="` + param.Mcode + `">` + param.Mname + `</option>`
			this.Data["Gcode_"+gcode+":"+param.Mcode] = param.Mname
		}
		this.Data["Gcode_"+gcode] = htmlvalue
	}

}

func (this *BaseController) MyGet() bool {
	fmt.Println(this.Ctx.Request.RequestURI)
	v := this.GetSession("username")
	if v == nil { //还未登录过，定向到登录页面
		this.TplName = "login.html"
		return false
	}
	this.User = v.(string)

	v = this.GetSession("name")
	this.Name = v.(string)

	memu0s, memuName, err := db.GetMemu(this.User, this.Ctx.Request.RequestURI)
	if err != nil {
		this.Ctx.WriteString("查询菜单失败！")
		return false
	}

	this.Params() //取参数

	this.Data["Memu"] = memu0s
	this.Data["Name"] = this.Name
	this.Data["MumeName"] = memuName
	this.Layout = "index.html"
	return true
}

func (this *BaseController) MyPost() bool {
	v := this.GetSession("username")
	if v == nil { //还未登录过，处理登录的请求
		Login(&this.Controller)
		return false
	}
	return true
}
