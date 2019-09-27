package memcached

import (
	"database/sql"
	"errors"
)

/*ips_message_cfg表操作 begin*/
type IpsMessageCfg struct {
	db   *sql.DB
	rows map[string][]TableRow `库表内存数据`
}

//创建对象
func NewIpsMessageCfg(v_db *sql.DB) (cfg *IpsMessageCfg, err error) {
	cfg = &IpsMessageCfg{}
	cfg.db = v_db
	cfg.rows = make(map[string][]TableRow)
	return
}

//从数据库加载数据
func (p *IpsMessageCfg) LoadConfig() (err error) {
	var results []map[string]string
	results, err = SqlSelect(p.db, `SELECT msgid,msgname,lev,msgtmp,nodes FROM ips_message_cfg`)
	if err != nil {
		return err
	}

	for _, v := range results {
		p.rows[v["msgid"]] = append(p.rows[v["msgid"]], v)
	}

	return
}

//传入msgid，返回信息
func (p *IpsMessageCfg) Get(msgid string) (cfg TableRow, err error) {
	cfglist := p.rows[msgid]
	if len(cfglist) < 1 {
		err = errors.New("msgid[" + msgid + "]not found!")
		return
	}
	cfg = cfglist[0]
	return
}

/*ips_message_cfg表操作 end*/
