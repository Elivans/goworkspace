package memcached

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type ProdInfo struct {
	db          *sql.DB
	srvcodeMem  map[string][]string   //网元缓存数据，key=pid,value=srvcode数组
	prdPcodeMem map[string][]TableRow //产品缓存，key=pid,value=prd_pcode内容
	prdAttrMem  map[string][]TableRow //产品缓存，key=pid,value=prd_attr内容
}

//创建对象
func NewProdInfo(v_db *sql.DB) (cfg *ProdInfo, err error) {
	cfg = &ProdInfo{}
	cfg.db = v_db
	cfg.srvcodeMem = make(map[string][]string)
	cfg.prdPcodeMem = make(map[string][]TableRow)
	cfg.prdAttrMem = make(map[string][]TableRow)
	return
}

//根据pid取得prd_pcode信息
func (p *ProdInfo) GetPrdCode(v_pid string) (prdCode TableRow, err error) {
	prdCode = make(TableRow)
	if len(p.prdPcodeMem[v_pid]) > 0 {
		//缓存有，直接取
		prdCode = p.prdPcodeMem[v_pid][0]
		return
	}
	//缓存没有，从数据库取
	var values []map[string]string
	values, err = SqlSelect(p.db, `SELECT pcode,pname,pclass,psubclass,permark,`+
		`externalid,status,isbase,prodtype,bandwidth,uflag,city `+
		` FROM prd_pcode WHERE pid=:1`, v_pid)
	if err != nil {
		return
	}

	for _, v := range values {
		p.prdPcodeMem[v_pid] = append(p.prdPcodeMem[v_pid], v)
	}

	if len(p.prdPcodeMem[v_pid]) == 0 {
		err = errors.New("pid=" + v_pid + " not found from prod_pcode")
		return
	}
	prdCode = p.prdPcodeMem[v_pid][0]
	return
}

//根据pid取得prd_attr信息
func (p *ProdInfo) GetPrdAttr(v_pid string) (prdAttr TableRow, err error) {
	prdAttr = make(TableRow)
	if len(p.prdAttrMem[v_pid]) > 0 {
		//缓存有，直接取
		prdAttr = p.prdAttrMem[v_pid][0]
		return
	}
	//缓存没有，从数据库取
	var values []map[string]string
	values, err = SqlSelect(p.db, `SELECT attrcode,attrvalue `+
		` FROM prd_attr WHERE pid=:1`, v_pid)
	if err != nil {
		return
	}

	for _, v := range values {
		prdAttr[v["attrcode"]] = v["attrvalue"]
	}

	p.prdAttrMem[v_pid] = append(p.prdAttrMem[v_pid], prdAttr)
	return
}

//根据pid取得srvcode
func (p *ProdInfo) Pid2Srvcode(v_pid string, v_city string, v_areaid string,
	v_itype string, v_subtype string) (srvcodes []string, err error) {
	if len(p.srvcodeMem[v_pid+":"+v_city+":"+v_areaid+":"+v_itype+":"+v_subtype]) > 0 {
		//缓存有，直接取
		srvcodes = p.srvcodeMem[v_pid+":"+v_city+":"+v_areaid+":"+v_itype+":"+v_subtype]
		return
	}
	//缓存没有，从数据库取
	var prdCode TableRow
	prdCode, err = p.GetPrdCode(v_pid)
	if err != nil {
		return
	}
	fmt.Println("------------------prdCode---", prdCode)
	var values []map[string]string
	if prdCode["uflag"] == "Y" { //省级产品
		values, err = SqlSelect(p.db, `SELECT DISTINCT srvcode FROM swp_relation a, swp_service_info b ,swp_service_detail c `+
			` WHERE a.srvid=b.srvid AND b.srvid=c.srvid`+
			`   AND a.pid=:1 `+
			`   AND c.itype=:2 `+
			`   AND c.subtype=:3 `+
			`   AND checkbincludea(:4,areas,',')='Y'`, v_pid, v_itype, v_subtype, v_areaid)
		if err != nil {
			return
		}
	} else { //非省级产品
		values, err = SqlSelect(p.db, `SELECT DISTINCT srvcode FROM swp_relation a, swp_service_info b ,swp_service_detail c `+
			` WHERE a.srvid=b.srvid AND b.srvid=c.srvid`+
			`   AND a.pid=:1 `+
			`   AND c.itype=:2 `+
			`   AND c.subtype=:3 `, v_pid, v_itype, v_subtype)
		if err != nil {
			return
		}
	}
	for _, v := range values {
		p.srvcodeMem[v_pid+":"+v_city+":"+v_areaid+":"+v_itype+":"+v_subtype] = append(p.srvcodeMem[v_pid+":"+v_city+":"+v_areaid+":"+v_itype+":"+v_subtype], v["srvcode"])
	}

	srvcodes = p.srvcodeMem[v_pid+":"+v_city+":"+v_areaid+":"+v_itype+":"+v_subtype]
	return
}

//取得<void>节点，CA授权有用到
func (p *ProdInfo) GetVoid_all(v_pids string, v_stime string, v_etimes string, v_city string, v_areaid string,
	v_itype string, v_subtype string) (void string, voidnum string, err error) {

	pidlist := strings.Split(v_pids, ",")
	etimelist := strings.Split(v_etimes, ",")
	srvCodes := make(map[string]string) //key为srvcode,value为etime
	for k, v := range pidlist {
		if len(v) <= 0 {
			continue
		}
		srvCodelist, err2 := p.Pid2Srvcode(v, v_city, v_areaid, v_itype, v_subtype)
		if err2 != nil {
			err = err2
			return
		}
		for _, vv := range srvCodelist {
			if len(srvCodes[vv]) == 0 || (len(srvCodes[vv]) > 0 && srvCodes[vv] < etimelist[k]) {
				//相同srvCode取时间大的
				srvCodes[vv] = etimelist[k]
			}
		}
	}
	void = ""
	for k, v := range srvCodes {
		stime := v_stime[0:4] + "-" + v_stime[4:6] + "-" + v_stime[6:8] +
			" " + v_stime[8:10] + ":" + v_stime[10:12] + ":" + v_stime[12:14]
		etime := v[0:4] + "-" + v[4:6] + "-" + v[6:8] +
			" " + v[8:10] + ":" + v[10:12] + ":" + v[12:14]
		void = void + "<caid>" + k + "</caid><stime>" + stime + "</stime><etime>" +
			etime + "</etime>"
	}
	voidnum = strconv.Itoa(len(srvCodes))
	return

}

//取得<void>节点，CA授权有用到(4K产品的)
func (p *ProdInfo) GetVoid_4k(v_pids string, v_stime string, v_etimes string, v_city string, v_areaid string,
	v_itype string, v_subtype string) (void string, voidnum string, err error) {
	pidlist := strings.Split(v_pids, ",")
	etimelist := strings.Split(v_etimes, ",")
	srvCodes := make(map[string]string) //key为srvcode,value为etime
	for k, pid := range pidlist {
		if len(pid) <= 0 {
			continue
		}
		prdAttr, err2 := p.GetPrdAttr(pid) //产品扩展属性
		if err2 != nil {
			err = err2
			return
		}
		if prdAttr["IS_4K_PROD"] != "Y" { //非4k产品，继续
			continue
		}
		srvCodelist, err2 := p.Pid2Srvcode(pid, v_city, v_areaid, v_itype, v_subtype)
		if err2 != nil {
			err = err2
			return
		}
		for _, vv := range srvCodelist {
			if len(srvCodes[vv]) == 0 || (len(srvCodes[vv]) > 0 && srvCodes[vv] < etimelist[k]) {
				//相同srvCode取时间大的
				srvCodes[vv] = etimelist[k]
			}
		}
	}
	void = ""
	for k, v := range srvCodes {
		stime := v_stime[0:4] + "-" + v_stime[4:6] + "-" + v_stime[6:8] +
			" " + v_stime[8:10] + ":" + v_stime[10:12] + ":" + v_stime[12:14]
		etime := v[0:4] + "-" + v[4:6] + "-" + v[6:8] +
			" " + v[8:10] + ":" + v[10:12] + ":" + v[12:14]
		void = void + "<caid>" + k + "</caid><stime>" + stime + "</stime><etime>" +
			etime + "</etime>"
	}
	voidnum = strconv.Itoa(len(srvCodes))
	return

}

//取得<void>节点，CA授权有用到(非4K产品的)
func (p *ProdInfo) GetVoid_not4k(v_pids string, v_stime string, v_etimes string, v_city string, v_areaid string,
	v_itype string, v_subtype string) (void string, voidnum string, err error) {
	pidlist := strings.Split(v_pids, ",")
	etimelist := strings.Split(v_etimes, ",")
	srvCodes := make(map[string]string) //key为srvcode,value为etime
	for k, pid := range pidlist {
		if len(pid) <= 0 {
			continue
		}
		prdAttr, err2 := p.GetPrdAttr(pid) //产品扩展属性
		if err2 != nil {
			err = err2
			return
		}
		if prdAttr["IS_4K_PROD"] == "Y" { //4k产品，继续
			continue
		}
		srvCodelist, err2 := p.Pid2Srvcode(pid, v_city, v_areaid, v_itype, v_subtype)
		if err2 != nil {
			err = err2
			return
		}
		for _, vv := range srvCodelist {
			if len(srvCodes[vv]) == 0 || (len(srvCodes[vv]) > 0 && srvCodes[vv] < etimelist[k]) {
				//相同srvCode取时间大的
				srvCodes[vv] = etimelist[k]
			}
		}
	}
	void = ""
	for k, v := range srvCodes {
		stime := v_stime[0:4] + "-" + v_stime[4:6] + "-" + v_stime[6:8] +
			" " + v_stime[8:10] + ":" + v_stime[10:12] + ":" + v_stime[12:14]
		etime := v[0:4] + "-" + v[4:6] + "-" + v[6:8] +
			" " + v[8:10] + ":" + v[10:12] + ":" + v[12:14]
		void = void + "<caid>" + k + "</caid><stime>" + stime + "</stime><etime>" +
			etime + "</etime>"
	}
	voidnum = strconv.Itoa(len(srvCodes))
	return

}
