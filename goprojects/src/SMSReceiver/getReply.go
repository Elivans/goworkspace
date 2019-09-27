// getReply
package main

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"strings"

	"github.com/axgle/mahonia"
)

//确认回复
func httpReplyConfirm(url string, spcode string, loginname string, password string, id string) {
	urlstr := url + "?SpCode=" + spcode + "&LoginName=" + loginname + "&Password=" + password + "&id=" + id
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
		logger.Printf("[faild httpReplyConfirm] HTTP GET '%s', '%s', '%s'\n", spcode, loginname, password)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Printf("[faild httpReplyConfirm] ReadAll '%s', '%s', '%s'\n", spcode, loginname, password)
		return
	}
	logger.Printf("[httpReplyConfirm]  '%s', '%s', '%s'[%s]\n", spcode, loginname, password, string(body))

}

//获取回复
func httpGetReply(url string, city string, areaid string, spcode string, loginname string, password string, wait chan int) (err error) {

	defer c_end(wait)

	urlstr := url + "?SpCode=" + spcode + "&LoginName=" + loginname + "&Password=" + password
	logger.Println(urlstr)
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
		logger.Printf("[faild] HTTP GET '%s', '%s', '%s'\n", spcode, loginname, password)
		return
	}

	//defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Printf("[faild] ReadAll '%s', '%s', '%s'\n", spcode, loginname, password)
		return
	}

	utf8Str := mahonia.NewDecoder("gbk").ConvertString(string(body))

	logger.Println(utf8Str)

	tmpstr := strings.Split(utf8Str, "&")
	if len(tmpstr) < 4 {
		logger.Printf("[faild] data format err '%s', '%s', '%s'\n", spcode, loginname, password)
		return
	}

	confirm_time := strings.Split(tmpstr[1], "=")[1]
	id := strings.Split(tmpstr[2], "=")[1]

	replys := strings.Split(tmpstr[3], "=")
	if len(replys) < 2 {
		logger.Printf("[faild] data format err '%s', '%s', '%s'\n", spcode, loginname, password)
		return
	}
	for i := 2; i < len(replys); i++ {
		replys[1] = replys[1] + replys[i]
	}

	logger.Println("\n\n\replys=\n%v", replys[1])
	var recDatas []RecData
	err = json.Unmarshal([]byte(replys[1]), &recDatas)

	if err != nil {
		logger.Printf("[faild] json.Unmarshal '%s', '%s', '%s'[%v]\n", spcode, loginname, password, err)
		return
	}

	tx, _ := db.Begin()
	defer tx.Rollback()
	for i := 0; i < len(recDatas); i++ {
		//recDatas[i].Content = mahonia.NewEncoder("utf-8").ConvertString(recDatas[i].Content)
		insStr := "INSERT INTO swp_smsreceived(recid, city, areaid, LASTID, callmdn, phonenumber, smstext, smsreptime, platreptime, inqtime, constatus, status)  VALUES(SEQ_GD_SMSREC_RECID.NEXTVAL, :1, :2, :3, :4, :5, :6, to_date(:7, 'YYYY-MM-DD HH24:MI:SS'), to_date(:8, 'YYYY-MM-DD HH24:MI:SS'), SYSDATE, '1', '0')"
		_, err = tx.Exec(insStr,
			city, areaid, id, recDatas[i].Callmdn, recDatas[i].Mdn, recDatas[i].Content, recDatas[i].Reply_time, confirm_time)
		if err != nil {
			logger.Printf("[faild] ins swp_smsreceived [%v]\n", err)
			logger.Printf("[faild]  [%v]\n", recDatas[i])
			logger.Printf("[Content]  [%v]\n", []byte(recDatas[i].Content))
			tx.Rollback()
			return
		}

	}
	//最后释放tx内部的连接
	tx.Commit()
	//logger.Println("[end] httpGet ")

	httpReplyConfirm(confirmUrl, spcode, loginname, password, id) //确认回复
	return

}
