package db

import (
	"errors"
	"fmt"

	"github.com/astaxie/beego"
	"maywide.pkg/database"
)

var dbdsn string

func init() {
	database.Addr = beego.AppConfig.String("DBAddr")
	database.Driver = beego.AppConfig.String("DBDriver")
	if "true" == beego.AppConfig.String("DBShow") {
		database.Show = true
	} else {
		database.Show = false
	}
	dbdsn = beego.AppConfig.String("DBLOGIN")
}

func Login(user string, password string) (isok bool, userinfo map[string]string, err error) {
	db := database.Connect(dbdsn, -1)
	defer db.Disconnect()
	var values []map[string]string
	values, err = db.SelectRows(0, `select * 
from prv_operator where loginname=:1 and passwd=:2`,
		user, password)

	if err != nil {
		fmt.Println(err.Error())
		isok = false
		return
	}
	fmt.Println(values)
	if len(values) != 1 {
		err = errors.New("登录失败，请检查用户名密码")
		return
	}
	isok = true
	userinfo = values[0]
	return
}
