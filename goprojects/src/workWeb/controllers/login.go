package controllers

import (
	"fmt"
	"workWeb/db"

	"github.com/astaxie/beego"
)

func Login(c *beego.Controller) {
	username := c.GetString("username")
	password := c.GetString("password")
	fmt.Println("username=", username, "password=", password)
	//验证用户码是否正确todo
	isok, userinfo, err := db.Login(username, password)
	if !isok { //验证失败
		c.Ctx.WriteString(err.Error())
		return
	}

	//验证成功,把用户名写入session
	c.SetSession("username", username)
	c.SetSession("name", userinfo["name"])

	c.Ctx.WriteString("true")
}

type LoginController struct {
	beego.Controller
}

func (this *LoginController) Get() {
	this.DestroySession()
	this.TplName = "login.html"
}

func (this *LoginController) Post() {
	Login(&this.Controller)
}
