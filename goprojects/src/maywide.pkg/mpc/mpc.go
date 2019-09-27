package mpc

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"github.com/redis"
)

type Mpc struct {
	Prior    int    // default 0, last
	Sendout  int    // default send timeout
	Recvout  int    // default read timeout
	Debug    bool   // default 0
	Sendname string // last send to list
	Recvname string // last recv from list
	client   redis.Client
}

func NewMpc(addr string, pass string) *Mpc {
	mpc := &Mpc{Prior: 0, Sendout: 10,
		Recvout: 60, Debug: false}
	mpc.client.Addr = addr
	mpc.client.Password = pass
	// clear all hand-check list
	hand := fmt.Sprintf("HAND_[12]%08d*", os.Getpid())
	keys, _ := mpc.client.Keys(hand)
	for i := 0; i < len(keys); i++ {
		mpc.client.Del(keys[i])
		fmt.Println(keys[i], "was deleted.")
	}
	return mpc
}

func (mpc *Mpc) SyncSend(name string, msg string, chanid int) error {
	if mpc.Prior < 0 || mpc.Prior > 9 {
		return errors.New("priority is outside [0-9].")
	}

	sndid := fmt.Sprintf("1%08d%d", os.Getpid(), chanid)
	redismsg := []byte(msg + "\x11\x01\x01\x01\x11" + sndid + "\x11\x01\x01\x01\x11")
	key := fmt.Sprint("SYNC_", name, "_", mpc.Prior)

	err := mpc.client.Lpush(key, redismsg)
	if err != nil {
		return err
	}

	if mpc.Debug {
		fmt.Printf("[%d] SyncSend[%s]=%s\n", chanid, key, string(redismsg))
	}

	keys := make([]string, 1)
	keys[0] = "HAND_" + sndid
	_, value, _ := mpc.client.Brpop(keys, uint(mpc.Sendout))
	if value == nil {
		mpc.client.Lpop(key)
		fmt.Printf("[%d] SyncSend.Hand=%s:%d\n", chanid, keys[0], mpc.Sendout)
		return errors.New("Send timeout exception")
	}

	if mpc.Debug {
		fmt.Printf("[%d] SyncSend.Receiver=%s\n", chanid, string(value))
	}

	err = mpc.client.Lpush("HAND_"+string(value), []byte(sndid))
	if err != nil {
		fmt.Printf("[%d] SyncSend.Handsend.Failed=%s", chanid, err.Error())
		return err
	}
	return nil
}

func (mpc *Mpc) SyncRecv(name string, chanid int) (string, error) {
	rcvid := fmt.Sprintf("2%08d%d", os.Getpid(), chanid)

	keys := make([]string, 10)
	for i := 0; i < 10; i++ {
		keys[i] = fmt.Sprint("SYNC_", name, "_", 9-i)
	}

	key, value, err := mpc.client.Brpop(keys, uint(mpc.Recvout))
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	if key == nil {
		fmt.Println("Timeout exception")
		return "", nil
	}

	mpc.Recvname = *key
	frommsg := string(value)

	s := bytes.Index(value, []byte("\x11\x01\x01\x01\x11"))
	e := bytes.LastIndex(value, []byte("\x11\x01\x01\x01\x11"))
	if e <= s {
		fmt.Println("key=", *key)
		fmt.Println("msg=", frommsg)
		return frommsg, nil
	}

	checkid := string(value[s+5 : e])
	frommsg = string(value[0:s])

	if mpc.Debug {
		fmt.Printf("[%d] SyncRecv=%s\n", chanid, frommsg)
	}

	// fmt.Printf("SyncRecv.Send.Hand=%s\n", "HAND_"+checkid)
	err = mpc.client.Lpush("HAND_"+checkid, []byte(rcvid))
	if err != nil {
		fmt.Printf("[%d] SyncRecv.Handrecv.Failed=%s\n", chanid, err.Error())
		return "", err
	}

	hnds := make([]string, 1)
	hnds[0] = "HAND_" + rcvid
	_, value, _ = mpc.client.Brpop(hnds, uint(2))
	if value == nil {
		mpc.client.Rpop("HAND_" + checkid)
		return "", errors.New("Receive timeout exception.")
	}

	if checkid != string(value) {
		fmt.Println("sorry: checkid=", checkid, ", handid=", string(value))
		return "", errors.New("Hand check failure, restart is recommended.")
	}

	if mpc.Debug {
		fmt.Printf("[%d] SyncRecv.key=%s", chanid, *key)
		fmt.Printf("[%d] SyncRecv.Sender=%s", chanid, string(value))
	}
	return frommsg, nil
}
