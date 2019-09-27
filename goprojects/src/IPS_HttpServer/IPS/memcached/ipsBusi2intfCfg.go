package memcached

import (
	"database/sql"
	"errors"
)

/*ips_message_cfg表操作 begin*/
type IpsBusi2intfCfg struct {
	db   *sql.DB
	rows map[string][]TableRow `库表内存数据`
}

//创建对象
func NewIpsBusi2intfCfg(v_db *sql.DB) (cfg *IpsBusi2intfCfg, err error) {
	cfg = &IpsBusi2intfCfg{}
	cfg.db = v_db
	cfg.rows = make(map[string][]TableRow)
	return
}

//从数据库加载数据
func (p *IpsBusi2intfCfg) LoadConfig() (err error) {
	var results []map[string]string
	results, err = SqlSelect(p.db, `SELECT recid,cfgname,city,ruleexpr,nodes,busiopcode,`+
		`itype,subtype,opcode,devno,msgid,orderno `+
		` FROM ips_busi2intf_cfg ORDER BY orderno`)
	if err != nil {
		return err
	}

	for _, v := range results {
		p.rows[v["city"]+":"+v["busiopcode"]] = append(p.rows[v["city"]+":"+v["busiopcode"]], v)
	}

	return
}

//传入msgid，返回信息
func (p *IpsBusi2intfCfg) Get(city string, busiopcode string) (cfgs []TableRow, err error) {
	cfglist := p.rows[city+":"+busiopcode]
	if len(cfglist) < 1 {
		err = errors.New("city,busiopcode[" + city + "," + busiopcode +
			"]not found from ips_busi2intf_cfg!")
		return
	}
	cfgs = cfglist
	return
}

/*ips_message_cfg表操作 end*/
