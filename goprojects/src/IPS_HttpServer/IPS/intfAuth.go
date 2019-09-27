// auth
package IPS

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	. "IPS_HttpServer/IPS/memcached"

	"maywide.pkg/expr2"
	"maywide.pkg/mylogger"
)

type Auth struct {
	db_sos          *sql.DB
	bizdata         string              //请求报文
	resdata         string              //响应报文
	logger          *mylogger.MyLogger  //日志打印
	swpiog          []map[string]string //待发送的报文
	sqlvalMem       map[string][]string //SQL值缓存，防止同样的SQL执行多次
	ipsBusi2intfCfg TableRow            //当前处理到的配置
	voidnum         string              //当前报文的voidnum

	root     map[string]string            //一级节点存储，如报文中的serialno、city等
	second   map[string]map[string]string //二级节点存储，如报文中的devinfo、olddevinfo、msg
	servlist []map[string]string          //需要开户的用户列表
	prodlist []map[string]string          //产品列表
	recid    string                       //seq_swp_iog_id序列，每条指令更新一次
}

//创建对象
func NewAuth(db_sos *sql.DB, logger *mylogger.MyLogger, bizdata string) (auth *Auth, err error) {
	auth = &Auth{}
	auth.db_sos = db_sos
	auth.bizdata = bizdata
	auth.logger = logger
	auth.root = make(map[string]string)
	auth.second = make(map[string]map[string]string)
	auth.sqlvalMem = make(map[string][]string)
	return
}

//报文分析得到有用数据（msgnode、prodlist）
func (p *Auth) Unmarshal() (err error) {
	//json解析
	var reqMap map[string]interface{}
	err = json.Unmarshal([]byte(p.bizdata), &reqMap)
	if err != nil {
		p.logger.Println("json.Unmarshal faild!")
		p.logger.Println(err)
		return
	}

	for k, v := range reqMap {
		//得到root
		if _, ok := v.(string); ok {
			//如果value是字符就加入
			p.root[k] = v.(string)
		}
		//得到second
		if _, ok := v.(map[string]interface{}); ok {
			third := make(map[string]string)
			for kk, vv := range v.(map[string]interface{}) {
				if _, ok := vv.(string); ok {
					//如果value是字符就加入
					third[kk] = vv.(string)
				}
			}
			p.second[k] = third
		}

		//得到servlist
		if _, ok := v.([]interface{}); ok && k == "servlist" {
			for _, vv := range v.([]interface{}) {
				servone := make(map[string]string)
				if _, ok := vv.(map[string]interface{}); ok {
					for kkk, vvv := range vv.(map[string]interface{}) {
						if _, ok := vvv.(string); ok {
							//如果value是字符就加入
							servone[kkk] = vvv.(string)
						}
					}
				}
				p.servlist = append(p.servlist, servone)
			}
		}
		//得到prodlist
		if _, ok := v.([]interface{}); ok && k == "prodlist" {
			for _, vv := range v.([]interface{}) {
				prodone := make(map[string]string)
				if _, ok := vv.(map[string]interface{}); ok {
					for kkk, vvv := range vv.(map[string]interface{}) {
						if _, ok := vvv.(string); ok {
							//如果value是字符就加入
							prodone[kkk] = vvv.(string)
						}
					}
				}

				//p.prodlist = append(p.prodlist, prodone)
				var prodlist []map[string]string
				prodlist, err = p.UnpkgProd(prodone)
				if err != nil {
					return
				}
				for _, v := range prodlist {
					p.prodlist = append(p.prodlist, v)
				}

			}
		}
	}
	return nil
}

//组全产品拆分
func (p *Auth) UnpkgProd(prod map[string]string) (prodlist []map[string]string, err error) {
	if prod["prodtype"] != "1" {
		prodlist = append(prodlist, prod)
		return
	}

	var results []map[string]string

	results, err = SqlSelect(p.db_sos, `SELECT sservid,spid,permark`+
		` FROM biz_prod_mix WHERE servid=:1 AND pid=:2`, prod["servid"], prod["pid"])
	if err != nil {
		return
	}

	for _, row := range results {
		sprod := make(map[string]string) //子产品信息
		for k, v := range prod {         //先继承产品信息
			sprod[k] = v
		}
		sprod["prodtype"] = "0"
		sprod["servid"] = row["sservid"]
		sprod["pid"] = row["spid"]
		sprod["permark"] = row["permark"]
		prodlist = append(prodlist, sprod)
	}

	return
}

//内部取值结果
func (p *Auth) Indata2Value(indata string, serv TableRow, prod TableRow) (nodevalue string, rtcode string, err error) {
	tmp := strings.Split(indata, ".")
	if tmp[0] == "root" && len(tmp) >= 2 {
		nodevalue = p.root[tmp[1]]
	} else if tmp[0] == "second" && len(tmp) >= 3 {
		nodevalue = p.second[tmp[1]][tmp[2]]
	} else if tmp[0] == "serv" && len(tmp) >= 2 {
		nodevalue = serv[tmp[1]]
	} else if tmp[0] == "prod" && len(tmp) >= 2 {
		nodevalue = prod[tmp[1]]
	} else if tmp[0] == "prd_pcode" && len(tmp) >= 2 {
		prodCode, err2 := M_ProdInfo.GetPrdCode(prod["pid"])
		if err2 != nil {
			return
		}
		nodevalue = prodCode[tmp[1]]
	} else if tmp[0] == "prd_attr" && len(tmp) >= 2 {
		prodAttr, err2 := M_ProdInfo.GetPrdAttr(prod["pid"])
		if err2 != nil {
			return
		}
		nodevalue = prodAttr[strings.ToUpper(tmp[1])]
	} else if tmp[0] == "void_all" {
		nodevalue, p.voidnum, err = M_ProdInfo.GetVoid_all(prod["pids"], p.root["stime"], prod["etimes"], p.root["city"], p.root["areaid"],
			p.ipsBusi2intfCfg["itype"], p.ipsBusi2intfCfg["subtype"])
	} else if tmp[0] == "void_4k" {
		nodevalue, p.voidnum, err = M_ProdInfo.GetVoid_4k(prod["pids"], p.root["stime"], prod["etimes"], p.root["city"], p.root["areaid"],
			p.ipsBusi2intfCfg["itype"], p.ipsBusi2intfCfg["subtype"])
	} else if tmp[0] == "void_not4k" {
		nodevalue, p.voidnum, err = M_ProdInfo.GetVoid_not4k(prod["pids"], p.root["stime"], prod["etimes"], p.root["city"], p.root["areaid"],
			p.ipsBusi2intfCfg["itype"], p.ipsBusi2intfCfg["subtype"])
	} else if tmp[0] == "void_num" {
		nodevalue = p.voidnum
	} else if tmp[0] == "recid" {
		nodevalue = p.recid
	} else {
		rtcode = "0001"
		err = errors.New("未定义的内部值类型" + indata)
		return
	}
	return
}

//SQL取值结果
func (p *Auth) SQL2Value(sqlid string, sqlvalueid int, serv TableRow, prod TableRow) (nodevalue string, rtcode string, err error) {
	sqlcfg, _ := M_ipsNodesqlCfg.Get(sqlid)
	var sqlparams []string //入参所使用的内部值列表
	if len(strings.TrimSpace(sqlcfg["sqlparams"])) >= 1 {
		sqlparams = strings.Split(sqlcfg["sqlparams"], ",") //入参标识
	}

	var params []interface{}           //入参值
	paramsdata := ""                   //入参值拼串
	for _, indata := range sqlparams { //通过入参标识取到入参值
		nodevalue, rtcode, err = p.Indata2Value(indata, serv, prod)
		if err != nil {
			return
		}
		params = append(params, nodevalue)
		paramsdata = paramsdata + ":"
	}
	if len(p.sqlvalMem[sqlid+":"+paramsdata]) >= sqlvalueid {
		nodevalue = p.sqlvalMem[sqlid+":"+paramsdata][sqlvalueid-1]
		return
	}
	//从数据库取
	db := DB_LINK[sqlcfg["dblink"]] //数据库
	sqlstr := sqlcfg["sql"]

	valuenum, err := strconv.Atoi(sqlcfg["valuenum"])
	if err != nil {
		rtcode = "0001"
		err = errors.New("非整数valuenum＝" + sqlcfg["valuenum"])
	}
	valuesitf := make([]interface{}, valuenum)
	for i := 0; i < valuenum; i++ {
		var tmp string
		valuesitf[i] = &tmp
	}

	var row *sql.Row
	if len(params) > 0 {
		row = db.QueryRow(sqlstr, params...)
	} else {
		row = db.QueryRow(sqlstr)
	}
	err = row.Scan(valuesitf...)
	var values []string //查询结果
	for i := 0; i < valuenum; i++ {
		values = append(values, *valuesitf[i].(*string))
	}

	//加入缓存 key为sqlid+各入参值
	p.sqlvalMem[sqlid+":"+paramsdata] = values

	nodevalue = p.sqlvalMem[sqlid+":"+paramsdata][sqlvalueid-1]

	return
}

//取节点值
func (p *Auth) GetNodeValue(node string, serv TableRow, prod TableRow) (nodevalue string, rtcode string, err error) {
	nodecfg, _ := M_ipsNodeCfg.Get(node)
	if nodecfg["type"] == "0" { //程序内部值
		indata := nodecfg["indata"]
		nodevalue, rtcode, err = p.Indata2Value(indata, serv, prod)
		if err != nil {
			return
		}
	} else if nodecfg["type"] == "1" { //SQL值
		sqlid := nodecfg["sqlid"]
		var sqlvalueid int
		sqlvalueid, err = strconv.Atoi(nodecfg["sqlvalueid"])
		if err != nil {
			rtcode = "0001"
			err = errors.New("非整数sqlvalueid＝" + nodecfg["sqlvalueid"])
		}
		nodevalue, rtcode, err = p.SQL2Value(sqlid, sqlvalueid, serv, prod)
		if err != nil {
			return
		}
	} else {
		err = errors.New("未定义的因子类型type=" + nodecfg["type"] + ",node=" + node)
	}

	return

}

//字条串中的因子替换为具体值
func (p *Auth) replaceValue(v_str string, v_nodeValue map[string]string) (str string) {
	str = v_str
	for k, v := range v_nodeValue { //因子标识替换为因子值
		str = strings.Replace(str, "$("+k+")", v, -1)
	}
	return
}

func (p *Auth) replaceValue2(str string, nodes string, serv TableRow, prod TableRow) (result string, rtcode string, err error) {
	if len(strings.TrimSpace(nodes)) < 2 {
		return
	}
	//先把节点的值取出来
	nodevalue := make(map[string]string)
	nodelist := strings.Split(nodes, ",")
	for _, node := range nodelist {
		nodevalue[node], rtcode, err = p.GetNodeValue(node, serv, prod)
		if err != nil {
			return
		}
	}
	//把值替换进去
	result = p.replaceValue(str, nodevalue)
	return
}

//规则是否满足
func (p *Auth) Ruleexpr(serv TableRow, prod TableRow, ipsMessageCfg TableRow) (isok bool, rtcode string, err error) {
	//是否满足规则
	ruleExpr := p.ipsBusi2intfCfg["ruleexpr"] //适用规则
	nodes := p.ipsBusi2intfCfg["nodes"]       //规则用到的因子
	var ruleExpr2 string                      //最终的规则
	ruleExpr2, rtcode, err = p.replaceValue2(ruleExpr, nodes, serv, prod)
	if err != nil {
		return
	}
	p.logger.Println(ruleExpr2)
	e, err2 := expr2.MustCompile2(ruleExpr2)
	if err2 != nil {
		//规则配置有异常要报错
		rtcode = "0001"
		err = errors.New("规则配置异常(" + err.Error() + ")")
		return
	}
	isok = e.Bool(expr2.V{})
	return
}

//单个处理
func (p *Auth) DoOne(serv TableRow, prod TableRow, ipsMessageCfg TableRow) (rtcode string, err error) {
	//取recid序列
	p.recid = ""
	db := DB_LINK["DB_SOS"] //数据库
	var row *sql.Row
	row = db.QueryRow("SELECT seq_swp_iog_id.nextval FROM dual")

	err = row.Scan(&p.recid)

	if err != nil || p.recid == "" {
		err = errors.New("get recid failed![" + err.Error() + "]")
		return
	}

	//设备取值
	var devno string
	devno, rtcode, err = p.GetNodeValue(p.ipsBusi2intfCfg["devno"], serv, prod)
	if err != nil {
		return
	}

	//是否满足规则
	//	var isok bool
	//	isok, rtcode, err = p.Ruleexpr(serv, prod, ipsMessageCfg)
	//	if err != nil {
	//		return
	//	}

	//	if isok != true {
	//		//不满足，不做处理
	//		return
	//	}
	//满足规则，继续
	msgtmp := ipsMessageCfg["msgtmp"] //报文模板
	nodes := ipsMessageCfg["nodes"]   //模板使用到的因子
	var reqdata string                //接口报文
	reqdata, rtcode, err = p.replaceValue2(msgtmp, nodes, serv, prod)
	if err != nil {
		return
	}

	//形成swp_iog数据
	iog := make(map[string]string)
	iog["recid"] = p.recid
	iog["itype"] = p.ipsBusi2intfCfg["itype"]
	iog["subtype"] = p.ipsBusi2intfCfg["subtype"]
	iog["priority"] = " "
	iog["opcode"] = p.ipsBusi2intfCfg["opcode"]
	iog["custid"] = p.root["custid"]
	iog["servid"] = prod["servid"]
	iog["devno"] = devno
	iog["optime"] = p.root["stime"]
	iog["etime"] = p.root["stime"]
	iog["serialno"] = p.root["serialno"]
	iog["nums"] = "0"
	iog["bizdata"] = p.bizdata
	iog["reqdata"] = reqdata
	iog["status"] = "S"
	iog["pri"] = p.root["pri"]
	iog["city"] = p.root["city"]
	iog["areaid"] = p.root["areaid"]

	p.swpiog = append(p.swpiog, iog)

	return
}

//报文转换
func (p *Auth) DoTranslate() (rtcode string, err error) {
	ipsBusi2intfCfgs, _ := M_ipsBusi2intfCfg.Get(p.root["city"], p.root["opcode"])
	for _, ipsBusi2intfCfg := range ipsBusi2intfCfgs {
		p.ipsBusi2intfCfg = ipsBusi2intfCfg
		ipsMessageCfg, _ := M_ipsMessageCfg.Get(ipsBusi2intfCfg["msgid"])
		if ipsMessageCfg["lev"] == "00" {
			//是否满足规则
			var isok bool
			isok, rtcode, err = p.Ruleexpr(make(TableRow), make(TableRow), ipsMessageCfg)
			if err != nil {
				return
			}

			if isok != true {
				//不满足，不做处理
				continue
			}
			//00:客户级
			rtcode, err = p.DoOne(make(TableRow), make(TableRow), ipsMessageCfg)
			if err != nil {
				return
			}

		} else if ipsMessageCfg["lev"] == "10" {
			//10:用户级(业务报文要求有servlist节点)
			for _, serv := range p.servlist {
				//是否满足规则
				var isok bool
				isok, rtcode, err = p.Ruleexpr(serv, make(TableRow), ipsMessageCfg)
				if err != nil {
					return
				}

				if isok != true {
					//不满足，不做处理
					continue
				}
				rtcode, err = p.DoOne(serv, make(TableRow), ipsMessageCfg)
				if err != nil {
					return
				}
			}

		} else if ipsMessageCfg["lev"] == "20" {
			//20:产品级(一个产品对应一个报文)(业务报文要求有pordlist节点)
			for _, prod := range p.prodlist {
				//是否满足规则
				var isok bool
				isok, rtcode, err = p.Ruleexpr(make(TableRow), prod, ipsMessageCfg)
				if err != nil {
					return
				}

				if isok != true {
					//不满足，不做处理
					continue
				}
				rtcode, err = p.DoOne(make(TableRow), prod, ipsMessageCfg)
				if err != nil {
					return
				}
			}
		} else if ipsMessageCfg["lev"] == "21" {
			//21:产品级(多个产品对应一个报文)(业务报文要求有pordlist节点)
			var prods []TableRow
			for _, prod := range p.prodlist {
				//是否满足规则
				var isok bool
				isok, rtcode, err = p.Ruleexpr(make(TableRow), prod, ipsMessageCfg)
				if err != nil {
					return
				}

				if isok != true {
					//不满足，不做处理
					continue
				}
				isfound := false
				//servid,permark相同的合成一条，增加pids,etime节点，其它信息以第一条为准
				for k, v := range prods {
					if prod["servid"] == v["servid"] && prod["permark"] == v["permark"] {
						prods[k]["pids"] = prods[k]["pids"] + "," + prod["pid"]
						prods[k]["etimes"] = prods[k]["etimes"] + "," + prod["etime"]
						isfound = true
						break
					}
				}
				if isfound == false {
					onep := make(TableRow)
					for k, v := range prod {
						onep[k] = v
					}
					onep["pids"] = onep["pid"]
					onep["etimes"] = onep["etime"]
					prods = append(prods, onep)
				}
			}
			//组合后每一条处理（prod比普通的多了pids,etimes节点）
			for _, prod := range prods {
				p.logger.Println("======================1")
				rtcode, err = p.DoOne(make(TableRow), prod, ipsMessageCfg)
				if err != nil {
					return
				}
				p.logger.Println("======================2")
			}

		} else {
			err = errors.New("lev=" + ipsMessageCfg["lev"] + " 不支持！")

		}
	}

	return
}

func (p *Auth) Send() (err error) {
	fmt.Println(p.swpiog)
	err = SendMsg(p.root["isrealtime"], p.swpiog)
	return
}

func (p *Auth) Check() (err error) {
	if p.root["serialno"] == "" {
		err = errors.New("serialno不能为空")
		return
	}
	if p.root["city"] == "" {
		err = errors.New("city不能为空")
		return
	}
	if p.root["areaid"] == "" {
		err = errors.New("areaid不能为空")
		return
	}
	if p.root["custid"] == "" {
		err = errors.New("custid不能为空")
		return
	}
	if p.root["opcode"] == "" {
		err = errors.New("opcode不能为空")
		return
	}

	if p.root["isrealtime"] == "" {
		p.root["isrealtime"] = "N"
	}
	if p.root["stime"] == "" {
		p.root["stime"] = time.Now().Format("20060102150405")
	}
	if p.root["pri"] == "" {
		p.root["pri"] = "1"
	}

	return
}

/*授权接口，全报文样例（一套设备）
{
  "serialno":"seri123456",
  "city":"DG",
  "areaid":"200",
  "custid":"c1234567",
  "opcode":"KK",
  "devinfo":{"keyno":"123","stbno":"123","cm":"7C08D9F045B2","keyno4k":"456"},
  "olddevinfo":{"keyno":"123","stbno":"123","cm":"7C08D9F045B2","keyno4k":"456"},
  "isrealtime":"Y",               //Y实时调接口返回处理结果,N提交到接口队列排队处理，默认N
  "stime":"yyyymmddhh24miss",     //非实时时此节点有效，为预约发送时间，为空默认当前时间
  "pri":"1",					  //非实时时此节点有效，优先级(1-99)，数字越大，优化级越高
  "servlist":[                    //设备相关的用户列表
    {"servid":"s123456","permark":"1"},
    {"servid":"s123457","permark":"3"}
  ],
  "prodlist":[                    //需要发授权的产品信息
    {"servid":"s123456","permark":"1","pid":"p123","etime":"20301231235959","prodtype","1"},
    {"servid":"s123456","permark":"1","pid":"p124","etime":"20301231235959","prodtype","0"},
    {"servid":"s123457","permark":"3","pid":"p125","etime":"20201231235959","prodtype","0"}
  ],
  "msg": {
	"title":"消息头",
	"content":"消息体",
	"round":"60", //循环次数
	"time":"3600",//持续时间
	"flag":"1" //OSD滚动方式
   }

}
*/
//{"serialno":"seri123456","city":"FS","areaid":"681","custid":"c1234567","opcode":"KK","devinfo":{"keyno":"123","stbno":"123","cm":"7C08D9F045B2"},"olddevinfo":{"keyno":"123","stbno":"123","cm":"7C08D9F045B2"},"isrealtime":"Y","stime":"20170628010101","pri":"1","prodlist":[{"servid":"s123456","pid":"p123","permark":"1","etime":"20301231235959","prodtype":"1"}]}
func IPS_Auth(logger *mylogger.MyLogger, bizdata string) (resdata string, err error) {
	rtcode := "9999"
	mylogger := logger.CloneLogger("IPS_Auth")
	auth, _ := NewAuth(DB_LINK["DB_SOS"], mylogger, bizdata) //创建实例
	err = auth.Unmarshal()                                   //分析报文
	if err != nil {
		resdata = code2msg1("0001", err.Error())
		logger.Printf("resdata=[%v]", resdata)
		return
	}
	err = auth.Check() //检查报文
	if err != nil {
		resdata = code2msg1("0001", err.Error())
		return
	}

	//报文转换...
	rtcode, err = auth.DoTranslate()
	if err != nil {
		resdata = code2msg1(rtcode, err.Error())
		return
	}

	err = auth.Send()
	if err != nil {
		resdata = code2msg1("0003", err.Error())
		return
	}

	//返回报文todo
	//resbyte, _ := json.Marshal(auth.swpiog)
	//resdata = string(resbyte)
	//resdata = auth.swpiog[0]["reqdata"]

	//	for k, v := range auth.swpiog {
	//		resdata = resdata + "--------------" + strconv.Itoa(k) + "\n"
	//		resdata = resdata + v["reqdata"] + "\n"
	//	}

	resdata = code2msg1("0000", "")

	return
}
