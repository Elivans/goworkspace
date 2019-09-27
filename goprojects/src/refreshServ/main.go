// refreshServ project main.go
package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-oci8"
)

var db *sql.DB

func init() {
	log.Println("init...")
	os.Setenv("NLS_LANG", "AMERICAN_AMERICA.UTF8")
	db, _ = sql.Open("oci8", "boss_sos/sZdgWk0w@gdboss")
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(1)
	db.Ping()
	db.Ping()
	log.Println("init end...")
}

func Sp(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	keyno, _ := r.Form["keyno"]
	fmt.Println(keyno[0])

	if len(keyno) < 1 {
		fmt.Fprintf(w, "刷新失败！")
		return
	}

	tx, _ := db.Begin()
	rows, err := tx.Query(`SELECT s.servid,str_link(p.pid) FROM sys_servst s, biz_product p
WHERE s.servid=p.servid
  AND s.servstatus = '2'
  AND p.pstatus='00'
  AND s.logicdevno=:1
  GROUP BY s.servid`, keyno[0])
	if err != nil {
		fmt.Println(err)
		fmt.Println("[faild] query  err")
		return
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
			return
		}
		rs, err := tx.Exec("call boss_adp.serv_control(:1,:2)", servid, pids)
		if err != nil {
			fmt.Println(err)
			fmt.Println("[faild] Exec  err", rs)
			return
		}
	}
	tx.Commit()

	if i < 1 {
		fmt.Fprintf(w, "无可刷新的数据!")
		return
	}

	fmt.Fprintf(w, "刷新成功！")

}

func main() {
	http.HandleFunc("/refresh", Sp)
	fmt.Println("listen on port 9090")
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
