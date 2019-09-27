// netsrv
package netsrv

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/axgle/mahonia"
)

type Netsrv struct {
	ADDR       string // address ip:port
	addr       *net.TCPAddr
	conn       *net.TCPConn
	checkid    uint64
	Reqtimeout int // default timeout of request
	Restimeout int // default timeout of response
	Buffer     int // default buffer for socket read
	resid      int
	status     bool
}

var connectid uint64 = 0

func FiledAdd(dest string, name string, value string) string {
	head := fmt.Sprintf("%02d%04d", len(name), len(value))
	return dest + head + name + value
}

func FiledValue(src string, name string) string {
	l := len(src)
	i := 0
	e := 0

	for {
		e = i + 2
		namelen := src[i:e]

		i = e
		e = i + 4
		vallen := src[i:e]

		nl, _ := strconv.Atoi(namelen)
		vl, _ := strconv.Atoi(vallen)

		if nl <= 0 || vl <= 0 {
			break
		}

		i = e
		e = i + nl
		code := src[i:e]

		i = e
		e = i + vl
		value := src[i:e]

		if code == name {
			return value
		}

		i = e
		if i >= l-1 {
			break
		}
	}
	return ""
}

func NewNetsrv(v_hostport string) (*Netsrv, error) {
	mnet := &Netsrv{}
	mnet.ADDR = v_hostport
	mnet.Buffer = 4096
	mnet.Reqtimeout = 15
	mnet.Restimeout = 60
	mnet.status = false
	connectid++
	err := mnet.Connect()
	return mnet, err
}

func (mnet *Netsrv) Connect() error {
	if mnet.status == true {
		return nil
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp4", mnet.ADDR)
	if err != nil {
		return errors.New("Connect.ResolveTCPAddr: " + err.Error())
	}

	mnet.addr = tcpAddr
	mnet.conn, err = net.DialTCP("tcp", nil, mnet.addr)
	if err != nil {
		return errors.New("Connect.DialTCP: " + err.Error())
	}

	mnet.conn.SetKeepAlive(true)

	msg := ""
	msg = FiledAdd(msg, "TYPE", "INIT")
	msg = FiledAdd(msg, "NAME", "GOLANG")

	_, err = mnet.conn.Write([]byte(msg))
	if err != nil {
		mnet.conn.Close()
		return errors.New("Connect.DialTCP: " + err.Error())
	}

	mnet.conn.SetReadDeadline(time.Now().Add(time.Second * 2))
	// get login status
	res := make([]byte, 64)
	reslen, err := mnet.conn.Read(res)
	if err != nil {
		mnet.Close()
		return errors.New("Connect.Init: " + err.Error())
	}

	if reslen <= 0 {
		mnet.Close()
		return errors.New("Connect.Socket was close by exceptions.")
	}

	if err = mnet.conn.SetReadDeadline(time.Time{}); err != nil {
		mnet.conn.Close()
		return errors.New("Connect.Set.Timeout Failed: " + err.Error())
	}

	status := FiledValue(string(res), "CODE")
	if status == "" {
		mnet.conn.Close()
		return errors.New("Connect.Init Failed: " + string(res))
	}

	mnet.status = true
	mnet.checkid = uint64(os.Getpid()) + connectid*10000 + 200000000
	return nil
}

func (mnet *Netsrv) Close() {
	mnet.conn.Close()
	mnet.status = false
}

func (mnet *Netsrv) SendSec(msgname string, msgtext string) error {
	if mnet.status == false {
		mnet.Connect()
		if mnet.status == false {
			return errors.New("SendSec: Netsrv.Socket was not created.")
		}
	}

	msg := ""
	msg = FiledAdd(msg, "TYPE", "SVR_SEND_SEC")
	msg = FiledAdd(msg, "MSGID", msgname)
	msg = FiledAdd(msg, "MSGTEXT", mahonia.NewEncoder("gbk").ConvertString(msgtext))
	msg = FiledAdd(msg, "LENGTH", fmt.Sprint(len(msgtext)))
	msg = FiledAdd(msg, "PRIORITY", "1")
	msg = FiledAdd(msg, "TIMEOUT", fmt.Sprint(mnet.Reqtimeout))

	_, err := mnet.conn.Write([]byte(msg))
	if err != nil {
		return errors.New("SendSec: " + err.Error())
	}

	// set timeout and read
	mnet.conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(mnet.Reqtimeout+2)))

	// get login status
	res := make([]byte, 64)
	reslen, err := mnet.conn.Read(res)
	if err != nil {
		return errors.New("SendSec: " + err.Error())
	}

	if reslen <= 0 {
		mnet.Close()
		return errors.New("SendSec.Socket was close by exceptions.")
	}

	// reset timeout status
	mnet.conn.SetReadDeadline(time.Time{})

	status := FiledValue(string(res), "CODE")
	if status != "0000" {
		status := FiledValue(string(res), "VALUE")
		return errors.New("netsrv.Return.Code=" + status)
	}
	return nil
}

func (mnet *Netsrv) RecvSec(msgname string) (string, error) {
	if mnet.status == false {
		mnet.Connect()
		if mnet.status == false {
			return "", errors.New("RecvSec: Netsrv.Socket was not created.")
		}
	}
	msg := ""
	msg = FiledAdd(msg, "TYPE", "SVR_RECV_SEC")
	msg = FiledAdd(msg, "MSGID", msgname)
	msg = FiledAdd(msg, "CHECKID", fmt.Sprint(mnet.checkid))
	msg = FiledAdd(msg, "TIMEOUT", fmt.Sprint(mnet.Restimeout))

	_, err := mnet.conn.Write([]byte(msg))
	if err != nil {
		return "", errors.New("RecvSec: " + err.Error())
	}

	// set timeout and read
	mnet.conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(mnet.Restimeout+2)))

	// get login status
	res := make([]byte, mnet.Buffer)
	reslen, err := mnet.conn.Read(res)
	if err != nil {
		return "", errors.New("RecvSec:" + err.Error())
	}

	if reslen <= 0 {
		mnet.Close()
		return "", errors.New("RecvSec.Socket was close by exceptions.")
	}

	// reset timeout status
	mnet.conn.SetReadDeadline(time.Time{})

	status := FiledValue(string(res), "CODE")
	// fmt.Printf("mnet.conn.Send.Status=%s\n", status)
	if status != "0000" {
		status := FiledValue(string(res), "VALUE")
		return "", errors.New("RecvSec.Return.Code=" + status)
	}

	message := mahonia.NewDecoder("gbk").ConvertString(FiledValue(string(res), "MSGTEXT"))
	return message, nil
}

func (mnet *Netsrv) SendSync(name string, msgname string, msgtext string) error {
	if mnet.status == false {
		mnet.Connect()
		if mnet.status == false {
			return errors.New("SendSec: Netsrv.Socket was not created.")
		}
	}
	mnet.resid = -1
	msg := ""
	msg = FiledAdd(msg, "TYPE", name)
	msg = FiledAdd(msg, "MSGID", msgname)
	msg = FiledAdd(msg, "MSGTEXT", mahonia.NewEncoder("gbk").ConvertString(msgtext))
	msg = FiledAdd(msg, "LENGTH", fmt.Sprint(len(msgtext)))
	msg = FiledAdd(msg, "PRIORITY", "1")
	msg = FiledAdd(msg, "TIMEOUT", fmt.Sprint(mnet.Reqtimeout))

	_, err := mnet.conn.Write([]byte(msg))
	if err != nil {
		return errors.New("SendSec: " + err.Error())
	}

	// set timeout and read
	mnet.conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(mnet.Reqtimeout+2)))

	// get login status
	res := make([]byte, 64)
	reslen, err := mnet.conn.Read(res)
	if err != nil {
		return errors.New("SendSec: " + err.Error())
	}

	if reslen <= 0 {
		mnet.Close()
		return errors.New("SendSec.Socket was close by exceptions.")
	}

	// reset timeout status
	mnet.conn.SetReadDeadline(time.Time{})

	status := FiledValue(string(res), "CODE")
	if status != "0000" {
		status := FiledValue(string(res), "VALUE")
		return errors.New("netsrv.Return.Code=" + status)
	}

	status = FiledValue(string(res), "VALUE")
	mnet.resid, _ = strconv.Atoi(status)
	return nil
}

func (mnet *Netsrv) RecvSync(name string) (string, error) {
	if mnet.status == false {
		mnet.Connect()
		if mnet.status == false {
			return "", errors.New("RecvSec: Netsrv.Socket was not created.")
		}
	}

	if mnet.resid < 0 {
		return "", errors.New("SyncResponse: No responseid was set, SyncRequest maybe failure.")
	}

	msg := ""
	msg = FiledAdd(msg, "TYPE", name)
	msg = FiledAdd(msg, "MSGID", fmt.Sprint(mnet.resid))
	msg = FiledAdd(msg, "CHECKID", fmt.Sprint(mnet.checkid))
	msg = FiledAdd(msg, "TIMEOUT", fmt.Sprint(mnet.Restimeout))

	_, err := mnet.conn.Write([]byte(msg))
	if err != nil {
		return "", errors.New("RecvSec: " + err.Error())
	}

	// set timeout and read
	mnet.conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(mnet.Restimeout+2)))

	// get login status
	res := make([]byte, mnet.Buffer)
	reslen, err := mnet.conn.Read(res)
	if err != nil {
		return "", errors.New("RecvSec:" + err.Error())
	}

	if reslen <= 0 {
		mnet.Close()
		return "", errors.New("RecvSec.Socket was close by exceptions.")
	}

	// reset timeout status
	mnet.conn.SetReadDeadline(time.Time{})

	status := FiledValue(string(res), "CODE")
	if status != "0000" {
		status := FiledValue(string(res), "VALUE")
		return "", errors.New("RecvSec.Return.Code=" + status)
	}

	message := mahonia.NewDecoder("gbk").ConvertString(FiledValue(string(res), "MSGTEXT"))
	mnet.Close()
	return message, nil
}

func (mnet *Netsrv) SyncRequest(msgname string, msgtext string) error {
	return mnet.SendSync("SVR_SEND_SYNC", msgname, msgtext)
}

func (mnet *Netsrv) SyncResponse() (string, error) {
	return mnet.RecvSync("RECV_SYNC2")
}

func (mnet *Netsrv) SecRequest(msgname string, msgtext string) error {
	return mnet.SendSync("SVR_SEND_SEC", msgname, msgtext)
}

func (mnet *Netsrv) SecResponse() (string, error) {
	mnet.resid = 0
	return mnet.RecvSync("RECV_SEC")
}
