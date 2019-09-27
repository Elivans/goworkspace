package memcached

import (
	"database/sql"
	"errors"
)

/*ips_node_cfg表操作 begin*/
type IpsNodeCfg struct {
	db   *sql.DB
	rows map[string][]TableRow `库表内存数据`
}

//创建对象
func NewIpsNodeCfg(v_db *sql.DB) (cfg *IpsNodeCfg, err error) {
	cfg = &IpsNodeCfg{}
	cfg.db = v_db
	cfg.rows = make(map[string][]TableRow)
	return
}

//从数据库加载数据
func (p *IpsNodeCfg) LoadConfig() (err error) {
	var results []map[string]string
	results, err = SqlSelect(p.db, `SELECT nodeid,node,name,`+
		`type,indata,sqlid,sqlvalueid `+
		` FROM ips_node_cfg`)
	if err != nil {
		return err
	}

	for _, v := range results {
		p.rows[v["node"]] = append(p.rows[v["node"]], v)
	}

	return
}

//传入node，返回节点信息
func (p *IpsNodeCfg) Get(node string) (cfg TableRow, err error) {
	cfglist := p.rows[node]
	if len(cfglist) < 1 {
		err = errors.New("node[" + node + "]not found!")
		return
	}
	cfg = cfglist[0]
	return
}

/*ips_node_cfg表操作 end*/
