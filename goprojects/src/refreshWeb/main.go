package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/astaxie/beego"
	_ "github.com/mattn/go-oci8"
)

var db *sql.DB

func init() {
	os.Setenv("NLS_LANG", "AMERICAN_AMERICA.UTF8")
	//db, _ = sql.Open("oci8", "boss_sos/boss_sos@10.205.28.19:1521/boss")
	db, _ = sql.Open("oci8", "boss_sos/sZdgWk0w@gdboss")
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(1)
	db.Ping()
	db.Ping()
}

func refresh(v_custid string, v_markno string, v_keyno string) error {
	var custid string
	if len(v_custid) < 1 && len(v_markno) < 1 && len(v_keyno) < 1 {
		return errors.New("请输入信息")
	}

	if len(v_custid) > 0 {
		custid = v_custid
	} else if len(v_markno) > 0 {
		row := db.QueryRow("SELECT custid FROM sys_cust WHERE markno = :1", v_markno)
		err := row.Scan(&custid)
		if err != nil {
			return err
		}
	} else if len(v_keyno) > 0 {
		row := db.QueryRow("SELECT custid FROM sys_servst WHERE logicdevno = :1 AND rownum=1", v_keyno)
		err := row.Scan(&custid)
		if err != nil {
			return err
		}
	}

	fmt.Println("custid=", custid)

	if len(custid) < 1 {
		return errors.New("未找到客户信息！")
	}

	tx, _ := db.Begin()
	rows, err := tx.Query(`SELECT s.servid,str_link(p.pid) FROM sys_servst s, biz_product p
WHERE s.servid=p.servid
  AND s.servstatus = '2'
  AND p.pstatus='00'
  AND s.custid=:1
  GROUP BY s.servid`, custid)
	if err != nil {
		fmt.Println(err)
		fmt.Println("[faild] query  err")
		return err
	}
	defer rows.Close()
	i := 0
	for rows.Next() {
		i = i + 1
		var (
			servid string
			pids   string
		)
		if err = rows.Scan(&servid, &pids); err != nil {
			fmt.Println(err.Error())
			return err
		}
		rs, err := tx.Exec("call boss_adp.serv_control(:1,:2)", servid, pids)
		if err != nil {
			fmt.Println(err)
			fmt.Println("[faild] Exec  err", rs)
			return err
		}
	}
	tx.Commit()

	if i < 1 {
		return errors.New("无可刷新的数据")

	}

	return nil

}

type MainController struct {
	beego.Controller
}

func (this *MainController) Get() {
	this.TplName = "index.tpl"
}
func (this *MainController) Post() {
	fmt.Println(this.Input())
	custid := this.Input().Get("custid")
	markno := this.Input().Get("markno")
	keyno := this.Input().Get("keyno")

	err := refresh(custid, markno, keyno)
	message := "刷新成功！"
	if err != nil {
		message = "刷新失败![" + err.Error() + "]"
	}

	this.Data["Custid"] = custid
	this.Data["Markno"] = markno
	this.Data["Keyno"] = keyno
	this.Data["Message"] = message
	this.TplName = "index.tpl"

	return
}

func main() {
	beego.Router("/", &MainController{})
	beego.Run()
}
