// test
package example_int

import (
	"bilHttpSrv/dbobj/bil"
	"database/sql"
	"encoding/json"
	"time"

	"maywide.pkg/mylogger"
)

//
func TestDB(tx *sql.Tx, logger *mylogger.MyLogger, reqdata string) (resdata string, err error) {
	logger.Println("test start...")
	var tb = &bil.Testdb{}
	err = json.Unmarshal([]byte(reqdata), tb)
	if err != nil {
		return
	}
	//for i := 0; i < 100; i++ {
	//time.Sleep(1 * time.Second)
	//logger.Println(i)
	err = tb.Insert(tx, tb)
	if err != nil {
		return
	}
	//}
	_, err = tx.Exec("INSERT INTO custid_test SELECT * FROM sys_cust")

	resdata = "{\"rtcode\":\"000\";\"message\":\"success\"}"
	return
}

func TestDB2(tx *sql.Tx, logger *mylogger.MyLogger, reqdata string) (resdata string, err error) {
	logger.Println("test2 start...")
	var tb = &bil.Testdb{}
	err = json.Unmarshal([]byte(reqdata), tb)
	if err != nil {
		return
	}
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		err = tb.Insert(tx, tb)
		if err != nil {
			return
		}
	}

	resdata = "{\"rtcode\":\"000\";\"message\":\"success\"}"
	return
}
