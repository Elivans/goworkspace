package main

import (
	"os"
	"time"

	"maywide.pkg/database/pool"
	"maywide.pkg/env"
	table "maywide.pkg/mgtv/table"
	mylog "maywide.pkg/mylogger"
)

func dbProcess(pid int) {
	for {
		// get from pool
		db := pool.GetDB()
		// set db for table.GetDB()
		table.SetDB(db)
		//mylog.Log("session[", pid, "]: database=", db.Sid)

		var prv_sysparam table.TablePrv_sysparam
		prv_sysparam.Paramid = "120716"
		prv_sysparam.Select("gcode,mcode")
		mylog.Log("session[", pid, "] gcode=", prv_sysparam.Gcode, ", mcode=", prv_sysparam.Mcode)
		time.Sleep(time.Second * time.Duration(1))

		// put into the pool
		pool.PutDB(db)

		// some session release
		if pid <= 2 {
			break
		}
	}
}

func main() {
	mylog.NewLogger("stdout")
	env.Iputenv("../mgtv.ini", "ENV")

	// connect pool created
	pool.Driver = "oracle"
	pool.Addr = "10.205.28.86:9527"
	pool.DSN = env.Value("DB", "ORACLE")
	pool.MinPoolSize = 5
	pool.MaxPoolSize = 5
	pool.Show = true

	mylog.Println(pool.DSN)

	// start 3 session
	for i := 0; i < 8; i++ {
		go dbProcess(i)
	}

	// clear idle connection per 30 seconds
	pool.AutoClear(30)

	mylog.Println("finish")
	os.Exit(0)
}
