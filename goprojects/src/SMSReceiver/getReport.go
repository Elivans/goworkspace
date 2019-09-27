// getReport
package main

import (
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"strings"
)

func getReport(url string, spcode string, loginname string, password string, wait chan int) {
	logger.Println("[begin] getReport ")
	defer c_end(wait)
	urlstr := url + "?SpCode=" + spcode + "&LoginName=" + loginname + "&Password=" + password
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
	resp, err := c.Get(urlstr)
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()
	if err != nil {
		logger.Printf("[getReport][faild] HTTP GET '%s', '%s', '%s'\n", spcode, loginname, password)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Printf("[getReport][faild] ReadAll '%s', '%s', '%s'\n", spcode, loginname, password)
		return
	}

	oneReport := strings.Split(string(body), ";")
	//logger.Printf("[getReport]%s\n", string(body))

	for i := 0; i < len(oneReport); i++ {

		tmpstr := strings.Split(oneReport[i], ",")
		if len(tmpstr) < 3 {
			continue
		}
		serial := tmpstr[0]
		phonenumber := tmpstr[1]
		result := tmpstr[2]
		logger.Printf("[getReport]serial=%s phonenumber=%s result=%s\n", serial, phonenumber, result)

		status := "3"
		if result == "0" {
			status = "4"
		}

		tx, _ := db.Begin()
		_, err = tx.Exec("update swp_smssend_log set status = :1 where sendserial = :2", status, serial)
		if err != nil {
			logger.Println("[getReport][faild]upd swp_smssend_log err", err)
			tx.Rollback()
			continue
		}
		tx.Commit()
	}
	logger.Println("[end] getReport ")

}
