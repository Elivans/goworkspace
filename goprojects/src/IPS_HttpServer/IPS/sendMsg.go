// save2db
package IPS

import (
	. "IPS_HttpServer/IPS/memcached"
	"database/sql"
	"encoding/xml"
	"errors"
	"reflect"

	"time"

	"maywide.pkg/netsrv"
)

func Insert_data(v_dbs interface{}, v_tabName string, v_recid string, v_data string, v_optime string) (err error) {

	MAX_LEN := 2000
	execsql := `INSERT INTO ` + v_tabName + `(pkid, recid, data, optime)` +
		`VALUES(seq_` + v_tabName + `_id.nextval, :1, :2, to_date(:3,'yyyymmddhh24miss'))`
	data_rune := []rune(v_data)
	isdone := false
	for {
		var data string
		if len(data_rune) <= MAX_LEN {
			isdone = true
			ll := len(data_rune)
			data = string(data_rune[0:ll])
		} else {
			data = string(data_rune[0:MAX_LEN])
			data_rune = data_rune[MAX_LEN:]
		}
		if reflect.TypeOf(v_dbs).Elem().Name() == "DB" {
			_, err = v_dbs.(*sql.DB).Exec(execsql, v_recid, data, v_optime)
		} else if reflect.TypeOf(v_dbs).Elem().Name() == "Tx" {
			_, err = v_dbs.(*sql.Tx).Exec(execsql, v_recid, data, v_optime)
		} else {
			//fmt.Println("db.type.name =", reflect.TypeOf(v_dbs).Elem().Name())
			return errors.New("Un-recognized db object")
		}

		if err != nil {
			return
		}
		if isdone {
			break
		}
	}
	return

}

func Insert_swpIog(tx *sql.Tx, values map[string]string) (err error) {
	values["optime"] = time.Now().Format("20060102150405")
	_, err = tx.Exec(`INSERT INTO swp_iog(recid, itype, subtype, priority, opcode, custid`+
		`, servid, devno, optime, etime, serialno, nums`+
		`, status,pri)`+
		`VALUES(:1, :2, :3, :4, :5, :6`+
		`, :7, :8, to_date(:9,'yyyymmddhh24miss'), to_date(:10,'yyyymmddhh24miss'), :11, :12`+
		`, :13, :14)`,
		values["recid"], values["itype"], values["subtype"], values["priority"],
		values["opcode"], values["custid"], values["servid"], values["devno"],
		values["optime"], values["etime"], values["serialno"], values["nums"],
		values["status"], values["pri"])

	if err != nil {
		return
	}

	err = Insert_data(tx, "swp_iog_bizdata", values["recid"], values["bizdata"], values["optime"])
	if err != nil {
		return
	}

	err = Insert_data(tx, "swp_iog_reqdata", values["recid"], values["reqdata"], values["optime"])
	if err != nil {
		return
	}

	return

}

func Insert_swpLog(tx *sql.Tx, values map[string]string) (err error) {
	values["optime"] = time.Now().Format("20060102150405")
	_, err = tx.Exec(`INSERT INTO swp_log(recid, itype, subtype, priority, opcode, custid`+
		`, servid, devno, optime, etime, serialno, nums`+
		`, isokay,rtcode,areaid)`+
		`VALUES(:1, :2, :3, :4, :5, :6`+
		`, :7, :8, to_date(:9,'yyyymmddhh24miss'), to_date(:10,'yyyymmddhh24miss'), :11, :12`+
		`, :13, :14, :15)`,
		values["recid"], values["itype"], values["subtype"], values["priority"],
		values["opcode"], values["custid"], values["servid"], values["devno"],
		values["optime"], values["etime"], values["serialno"], values["nums"],
		values["isokay"], values["rtcode"], values["areaid"])

	if err != nil {
		err = errors.New("ins swp_log failed!" + err.Error())
		return
	}
	err = Insert_data(tx, "swp_iog_bizdata", values["recid"], values["bizdata"], values["optime"])
	if err != nil {
		err = errors.New("ins swp_iog_bizdata failed!" + err.Error())
		return
	}

	err = Insert_data(tx, "swp_iog_reqdata", values["recid"], values["reqdata"], values["optime"])
	if err != nil {
		err = errors.New("ins swp_iog_reqdata failed!" + err.Error())
		return
	}

	err = Insert_data(tx, "swp_iog_rtdata", values["recid"], values["rtdata"], values["optime"])
	if err != nil {
		err = errors.New("ins swp_iog_rtdata failed!" + err.Error())
		return
	}
	return

}

func SendNetSrv(swpiog map[string]string) (rtdata string, err error) {
	ipaddr, err2 := M_cfg.String("NET_SRV", "ADDR")
	if err2 != nil {
		err = err2
		return
	}
	reqid, err3 := M_SwpInfo.Get_reqid(swpiog["itype"], swpiog["subtype"])
	if err3 != nil {
		err = err3
		return
	}

	var m1 *netsrv.Netsrv
	m1, err = netsrv.NewNetsrv(ipaddr)
	if err != nil {
		return
	}
	m1.SyncRequest(reqid, swpiog["reqdata"])
	rtdata, err = m1.SyncResponse()
	if err != nil {
		return
	}
	return

}

type Response struct {
	Status  string `xml:"status"`
	Output  Output `xml:"output"`
	Code    string `xml:"code"`
	Message string `xml:"message"`
}

type Output struct {
	Serialno string `xml:"serialno"`
	Recid    string `xml:"recid"`
}

func SendMsg(isrtime string, swpiogs []map[string]string) (err error) {
	if isrtime == "Y" {
		//实时发送，调接口
		for _, iog := range swpiogs {
			iog["isokay"] = "N"
			var rtdata string
			rtdata, err = SendNetSrv(iog)
			if err != nil {
				err = errors.New(iog["subtype"] + "[" + err.Error() + "];")
				iog["rtcode"] = "MMM"
				iog["rtdata"] = err.Error()
			} else {
				var result Response
				err = xml.Unmarshal([]byte(rtdata), &result)
				if err != nil {
					err = errors.New(iog["subtype"] + "[" + err.Error() + "];")
				}
				rtcode := result.Code
				if len(rtcode) > 0 {
					if rtcode == "000" || rtcode == "0000" {
						iog["isokay"] = "Y"
					}
				} else {
					rtcode = " "
				}
				iog["rtcode"] = rtcode
				iog["rtdata"] = rtdata
			}

			//写log
			tx, err2 := DB_LINK["DB_SOS"].Begin()
			defer tx.Rollback()
			if err2 != nil {
				err = err2
				return
			}
			err2 = Insert_swpLog(tx, iog)
			if err2 != nil {
				err = err2
				return
			}
			tx.Commit()
		}
	} else {
		//非实时，写待发送表
		tx, err2 := DB_LINK["DB_SOS"].Begin()
		defer tx.Rollback()
		if err2 != nil {
			err = err2
			return
		}
		for _, iog := range swpiogs {
			err = Insert_swpIog(tx, iog)
			if err != nil {
				return
			}
		}
		tx.Commit()
	}

	return
}
