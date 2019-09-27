// 程序运行时加载的内容
package memcached

import (
	"database/sql"

	"github.com/larspensjo/config"
	"maywide.pkg/mylogger"
)

var DB_LINK map[string]*sql.DB
var M_cfg *config.Config

var M_ipsNodeCfg *IpsNodeCfg
var M_ipsNodesqlCfg *IpsNodesqlCfg
var M_ipsMessageCfg *IpsMessageCfg
var M_ipsBusi2intfCfg *IpsBusi2intfCfg
var M_ProdInfo *ProdInfo
var M_SwpInfo *SwpInfo

type TableRow map[string]string

//加载所有参数配置信息
func LoadDBConfig(dblink map[string]*sql.DB, v_logger *mylogger.MyLogger, v_cfg *config.Config) (err error) {
	M_cfg = v_cfg
	DB_LINK = dblink
	//logger := v_logger.CloneLogger("LoadDBConfig")

	M_ipsNodeCfg, _ = NewIpsNodeCfg(DB_LINK["DB_SOS"])
	M_ipsNodesqlCfg, _ = NewIpsNodesqlCfg(DB_LINK["DB_SOS"])
	M_ipsMessageCfg, _ = NewIpsMessageCfg(DB_LINK["DB_SOS"])
	M_ipsBusi2intfCfg, _ = NewIpsBusi2intfCfg(DB_LINK["DB_SOS"])
	M_ProdInfo, _ = NewProdInfo(DB_LINK["DB_SOS"])
	M_SwpInfo, _ = NewSwpInfo(DB_LINK["DB_SOS"])
	//开始加载一次
	err = M_ipsNodeCfg.LoadConfig()
	if err != nil {
		return
	}
	err = M_ipsNodesqlCfg.LoadConfig()
	if err != nil {
		return
	}
	err = M_ipsMessageCfg.LoadConfig()
	if err != nil {
		return
	}
	err = M_ipsBusi2intfCfg.LoadConfig()
	if err != nil {
		return
	}
	err = M_SwpInfo.LoadConfig()
	if err != nil {
		return
	}
	//	//以后每5分钟加载一次
	//	go func() {
	//		for range time.Tick(time.Second * 600) {
	//			err = M_ipsNodeCfg.LoadConfig()
	//			if err != nil {
	//				logger.Println("Load ips_node_cfg faild...")
	//			}
	//			err = M_ipsNodesqlCfg.LoadConfig()
	//			if err != nil {
	//				logger.Println("Load ips_nodesql_cfg faild...")
	//			}
	//			err = M_ipsMessageCfg.LoadConfig()
	//			if err != nil {
	//				logger.Println("Load ips_message_cfg faild...")
	//			}
	//			err = M_ipsBusi2intfCfg.LoadConfig()
	//			if err != nil {
	//				logger.Println("Load ips_busi2intf_cfg faild...")
	//			}
	//		}

	//	}()
	return
}
