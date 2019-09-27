package memcached

import (
	"database/sql"
	"errors"
)

type SwpInfo struct {
	db   *sql.DB
	rows map[string][]TableRow `库表内存数据`
}

//创建对象
func NewSwpInfo(v_db *sql.DB) (cfg *SwpInfo, err error) {
	cfg = &SwpInfo{}
	cfg.db = v_db
	cfg.rows = make(map[string][]TableRow)
	return
}

//从数据库加载数据
func (p *SwpInfo) LoadConfig() (err error) {
	var results []map[string]string
	results, err = SqlSelect(p.db, `SELECT itype,subtype,reqid FROM swp_info WHERE isuse = 'Y'`)
	if err != nil {
		return err
	}

	for _, v := range results {
		p.rows[v["itype"]+":"+v["subtype"]] = append(p.rows[v["itype"]+":"+v["subtype"]], v)
	}

	return
}

//传入sqlid，返回节点信息
func (p *SwpInfo) Get(itype string, subtype string) (cfg TableRow, err error) {
	cfglist := p.rows[itype+":"+subtype]
	if len(cfglist) < 1 {
		err = errors.New("swp_info not found! itype=" + itype + ",subtype" + subtype)
		return
	}
	cfg = cfglist[0]
	return
}

func (p *SwpInfo) Get_reqid(itype string, subtype string) (reqid string, err error) {
	cfg, err2 := p.Get(itype, subtype)
	if err2 != nil {
		err = err2
		return
	}
	reqid = cfg["reqid"]
	return
}
