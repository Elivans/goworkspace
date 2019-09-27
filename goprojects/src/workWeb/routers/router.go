package routers

import (
	"workWeb/controllers"

	"workWeb/controllers/work"

	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/", &controllers.LoginController{})
	beego.Router("/index", &controllers.IndexController{})
	beego.Router("/workinfo", &work.WorkInfoController{})
}
