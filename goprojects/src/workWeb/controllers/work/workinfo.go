package work

import (
	"fmt"

	. "workWeb/controllers"
)

//查询
func (this *WorkInfoController) query() {
	//pagenum := this.Input().Get("pagenum") //页码

	var records []map[string]string
	one := make(map[string]string)
	one["workid"] = "T2017092500001"
	one["workname"] = "统一对账平台开发"

	records = append(records, one)
	this.Data["Records"] = records
	return
}

type WorkInfoController struct {
	BaseController
}

func (this *WorkInfoController) Get() {
	this.UseParams = append(this.UseParams, "work_status") //页面用到的参数
	this.UseParams = append(this.UseParams, "work_type")   //页面用到的参数
	if !this.MyGet() {
		return
	}
	fmt.Println(this.User)
	this.TplName = "work/workinfo.html"
}

func (this *WorkInfoController) Post() {
	if !this.MyPost() {
		return
	}
	method := this.Input().Get("method") //调用方法
	if method == "query" {
		this.query()
		fmt.Println("qqqqqqqqqq")
	}

	this.Get()
}
