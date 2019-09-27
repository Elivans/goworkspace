// callDT
package main

import (
	"fmt"
	"net"
	"time"

	"net/http"
	"strings"
)

func callDT(url string, phonenumber string, msg string) (err error) {
	reqdata := fmt.Sprintf("{phonenumber:\"%s\",channel:\"0\",smstext:\"%s\"}",
		phonenumber, msg)
	logger.Println("[callDT]post to", url, reqdata)

	c := http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(25 * time.Second)
				c, err := net.DialTimeout(netw, addr, time.Second*20)
				if err != nil {
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
		},
	}
	resp, err := c.Post(url,
		"application/x-www-form-urlencoded",
		strings.NewReader(reqdata))
	if err != nil {
		logger.Println("[faild]callDT Post err", err)
		return
	}
	logger.Println("[callDT]Post end")

	defer resp.Body.Close()
	//	body, err := ioutil.ReadAll(resp.Body)
	//	if err != nil {
	//		logger.Println("[faild]callDT ReadAll err", err)
	//		return
	//	}
	//	logger.Println(string(body))
	//fmt.Println(string(body))
	return
}

//扫描数据发送到短厅
func sendToDT() {
	logger.Println("[begin] sendToDT ")
	rows, err := db.Query(`SELECT a.recid, a.phonenumber, a.smstext, a.city,b.serv_url 
							 FROM swp_smsreceived a,
							      boss_crm.DT_SMSSERVICE b,
								(select code rtcode from boss_crm.DT_ENSURE_SMS_CODE
                                  union
                                 select rtcode from boss_crm.dt_sms_crm) c 
							WHERE a.city=b.city 
							  AND a.status = 0  
							  AND a.smstext=c.rtcode AND a.smsreptime>SYSDATE-4/24`)
	if err != nil {
		logger.Println("[faild] query swp_smsreceived err")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var recid string
		var phonenumber string
		var smstext string
		var city string
		var serv_url string
		if err = rows.Scan(&recid, &phonenumber, &smstext, &city, &serv_url); err != nil {
			return
		}
		//调接口
		err = callDT(serv_url, phonenumber, smstext)
		if err != nil {
			continue
		}
		logger.Println("[sendToDT]callDT end")
		tx, _ := db.Begin()
		logger.Println("[sendToDT]db.Begin")
		_, err = tx.Exec("UPdate swp_smsreceived set status ='2' where recid=:1", recid)
		if err != nil {
			logger.Println("[faild]upd swp_smsreceived err", err)
			tx.Rollback()
			continue
		}
		tx.Commit()
	}
	logger.Println("[end] sendToDT ")
}

func main2() {
	//发到短厅
	callDT("http://10.205.9.8:8888/", "15989101900", "FT")
}
