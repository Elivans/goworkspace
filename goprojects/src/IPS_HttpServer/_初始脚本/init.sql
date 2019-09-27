-------------------------------------------------;
--8.1.4 : 业务报文对应接口报文, 表结构;
-------------------------------------------------;
CREATE TABLE ips_busi2intf_cfg (
  recid                          NUMBER(16)       NOT NULL,
  cfgname                        VARCHAR2(100)    NOT NULL,
  city                           VARCHAR2(4)      NOT NULL,
  ruleexpr                       VARCHAR2(1024)   NOT NULL,
  nodes                          VARCHAR2(500)        NULL,
  busiopcode                     VARCHAR2(4)      NOT NULL,
  itype                          VARCHAR2(4)      NOT NULL,
  subtype                        VARCHAR2(8)      NOT NULL,
  opcode                         VARCHAR2(4)      NOT NULL,
  devno                          VARCHAR2(50)     NOT NULL,
  msgid                          NUMBER(16)       NOT NULL,
  orderno                        NUMBER(6)        NOT NULL,
  remark                         VARCHAR2(200)        NULL
);

ALTER TABLE ips_busi2intf_cfg ADD CONSTRAINTS pk_ips_busi2intf_cfg PRIMARY KEY (recid);
CREATE SEQUENCE seq_ips_busi2intf_cfg_id START WITH 1 CACHE 1000 ORDER;
COMMENT ON TABLE ips_busi2intf_cfg IS '业务报文对应接口报文';
COMMENT ON COLUMN ips_busi2intf_cfg.recid IS '序号';
COMMENT ON COLUMN ips_busi2intf_cfg.cfgname IS '配置名称';
COMMENT ON COLUMN ips_busi2intf_cfg.city IS '适用地市';
COMMENT ON COLUMN ips_busi2intf_cfg.ruleexpr IS '适用规则,如："$(AREAID)"=="200"&&"$(PERMARK)"=="2"&&regexp("^875810[0-9]{10}$","$(SERVINFO.KEYNO)")';
COMMENT ON COLUMN ips_busi2intf_cfg.nodes IS '规则用到的因子';
COMMENT ON COLUMN ips_busi2intf_cfg.busiopcode IS '业务操作码';
COMMENT ON COLUMN ips_busi2intf_cfg.itype IS '接口类型';
COMMENT ON COLUMN ips_busi2intf_cfg.subtype IS '接口子类型';
COMMENT ON COLUMN ips_busi2intf_cfg.opcode IS '接口操作码';
COMMENT ON COLUMN ips_busi2intf_cfg.devno IS '设备取值，使用因子，为了写swp_iog.devno';
COMMENT ON COLUMN ips_busi2intf_cfg.msgid IS '报文编号';
COMMENT ON COLUMN ips_busi2intf_cfg.orderno IS '报文顺序';
COMMENT ON COLUMN ips_busi2intf_cfg.remark IS '备注';

------------------------------------------------;
--8.1.5 : 接口报文配置表, 表结构;
-- ---------------------------------------------;
CREATE TABLE ips_message_cfg (
  msgid                          NUMBER(16)       NOT NULL,
  msgname                        VARCHAR2(100)    NOT NULL,
  level                          VARCHAR2(2)      NOT NULL,
  msgtmp                         VARCHAR2(4000)   NOT NULL,
  nodes                          VARCHAR2(500)        NULL,
  remark                         VARCHAR2(256)        NULL
);

ALTER TABLE ips_message_cfg ADD CONSTRAINTS pk_ips_message_cfg PRIMARY KEY (msgid);
CREATE SEQUENCE seq_ips_message_cfg_id START WITH 1 CACHE 1000 ORDER;
COMMENT ON TABLE ips_message_cfg IS '接口报文配置表';
COMMENT ON COLUMN ips_message_cfg.msgid IS '报文编号';
COMMENT ON COLUMN ips_message_cfg.msgname IS '报文名称';
COMMENT ON COLUMN ips_message_cfg.level IS '级别，00:客户级;10:用户级(业务报文要求有servlist节点);20:产品级(一个产品对应一个报文)(业务报文要求有pordlist节点);21:产品级(多个产品对应一个报文)(业务报文要求有pordlist节点)';
COMMENT ON COLUMN ips_message_cfg.msgtmp IS '报文模板';
COMMENT ON COLUMN ips_message_cfg.nodes IS '模板使用到的因子';
COMMENT ON COLUMN ips_message_cfg.remark IS '报文注释';

------------------------------------------------;
--8.1.6 : 模板因子配置表, 表结构;
------------------------------------------------;
CREATE TABLE ips_node_cfg (
  nodeid                         NUMBER(16)       NOT NULL,
  node                           VARCHAR2(32)     NOT NULL,
  name                           VARCHAR2(32)     NOT NULL,
  type                           VARCHAR2(1)      NOT NULL,
  indata                         VARCHAR2(100)        NULL,
  sqlid                          NUMBER(10)           NULL,
  sqlvalueid                     NUMBER(2)            NULL,
  memo                           VARCHAR2(512)        NULL
);

-------------------------------------------------;
--8.1.6 : 模板因子配置表 数据字典;
-------------------------------------------------;
COMMENT ON TABLE ips_node_cfg IS '模板因子配置表';
COMMENT ON COLUMN ips_node_cfg.nodeid IS '序号';
COMMENT ON COLUMN ips_node_cfg.node IS '节点因子';
COMMENT ON COLUMN ips_node_cfg.name IS '因子名字';
COMMENT ON COLUMN ips_node_cfg.type IS '因子类型,0:程序内部参数　1:SQL值';
COMMENT ON COLUMN ips_node_cfg.indata IS '内部值代号,Type为0时有效';
COMMENT ON COLUMN ips_node_cfg.sqlid IS 'SQLID,Type为1时有效';
COMMENT ON COLUMN ips_node_cfg.sqlvalueid IS 'SQL的第几个值,Type为1时有效';
COMMENT ON COLUMN ips_node_cfg.memo IS '备注';


------------------------------------------------;
-- 8.1.7 : 因子配置表, 表结构;
------------------------------------------------;
CREATE TABLE ips_nodesql_cfg (
  sqlid                          NUMBER(16)       NOT NULL,
  sqlname                        VARCHAR2(100)        NULL,
  dblink                         VARCHAR2(20)          NULL,
  sql                            VARCHAR2(1024)   NOT NULL,
  valuenum                       NUMBER(2)        NOT NULL,
  sqlparams                      VARCHAR2(200)        NULL
) initrans 20;

ALTER TABLE ips_nodesql_cfg ADD CONSTRAINTS pk_ips_nodesql_cfg PRIMARY KEY (sqlid);
CREATE SEQUENCE seq_ips_nodesql_cfg_id START WITH 1 CACHE 1000 ORDER;

COMMENT ON TABLE ips_nodesql_cfg IS '因子配置表';
COMMENT ON COLUMN ips_nodesql_cfg.sqlid IS '序号';
COMMENT ON COLUMN ips_nodesql_cfg.sqlname IS 'SQL名,SQL语句可使用内部值代号，见附录（内部值代号）';
COMMENT ON COLUMN ips_nodesql_cfg.dblink IS '数据源';
COMMENT ON COLUMN ips_nodesql_cfg.sql IS 'SQL语句';
COMMENT ON COLUMN ips_nodesql_cfg.valuenum IS '查询结果数量';
COMMENT ON COLUMN ips_nodesql_cfg.sqlparams IS 'SQL入参,只能是内部值';


