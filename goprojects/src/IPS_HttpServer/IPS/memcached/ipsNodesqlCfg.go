package memcached

import (
	"database/sql"
	"errors"
)

/*ips_nodesql_cfg表操作 begin*/
type IpsNodesqlCfg struct {
	db   *sql.DB
	rows map[string][]TableRow `库表内存数据`
}

//创建对象
func NewIpsNodesqlCfg(v_db *sql.DB) (cfg *IpsNodesqlCfg, err error) {
	cfg = &IpsNodesqlCfg{}
	cfg.db = v_db
	cfg.rows = make(map[string][]TableRow)
	return
}

//从数据库加载数据
func (p *IpsNodesqlCfg) LoadConfig() (err error) {
	var results []map[string]string
	results, err = SqlSelect(p.db, `SELECT sqlid,sqlname,dblink,sql,valuenum,sqlparams`+
		` FROM ips_nodesql_cfg`)
	if err != nil {
		return err
	}

	for _, v := range results {
		p.rows[v["sqlid"]] = append(p.rows[v["sqlid"]], v)
	}

	return
}

//传入sqlid，返回节点信息
func (p *IpsNodesqlCfg) Get(sqlid string) (cfg TableRow, err error) {
	cfglist := p.rows[sqlid]
	if len(cfglist) < 1 {
		err = errors.New("sqlid[" + sqlid + "]not found!")
		return
	}
	cfg = cfglist[0]
	return
}

/*ips_nodesql_cfg表操作 end*/
