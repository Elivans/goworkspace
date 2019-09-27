package controllers

import (
	"fmt"
)

type IndexController struct {
	BaseController
}

func (this *IndexController) Get() {
	if !this.MyGet() {
		return
	}
	fmt.Println(this.User)
	this.TplName = "user/add.html"
}

func (this *IndexController) Post() {
	if !this.MyPost() {
		return
	}
}
