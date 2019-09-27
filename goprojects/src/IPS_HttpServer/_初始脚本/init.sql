-------------------------------------------------;
--8.1.4 : ҵ���Ķ�Ӧ�ӿڱ���, ��ṹ;
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
COMMENT ON TABLE ips_busi2intf_cfg IS 'ҵ���Ķ�Ӧ�ӿڱ���';
COMMENT ON COLUMN ips_busi2intf_cfg.recid IS '���';
COMMENT ON COLUMN ips_busi2intf_cfg.cfgname IS '��������';
COMMENT ON COLUMN ips_busi2intf_cfg.city IS '���õ���';
COMMENT ON COLUMN ips_busi2intf_cfg.ruleexpr IS '���ù���,�磺"$(AREAID)"=="200"&&"$(PERMARK)"=="2"&&regexp("^875810[0-9]{10}$","$(SERVINFO.KEYNO)")';
COMMENT ON COLUMN ips_busi2intf_cfg.nodes IS '�����õ�������';
COMMENT ON COLUMN ips_busi2intf_cfg.busiopcode IS 'ҵ�������';
COMMENT ON COLUMN ips_busi2intf_cfg.itype IS '�ӿ�����';
COMMENT ON COLUMN ips_busi2intf_cfg.subtype IS '�ӿ�������';
COMMENT ON COLUMN ips_busi2intf_cfg.opcode IS '�ӿڲ�����';
COMMENT ON COLUMN ips_busi2intf_cfg.devno IS '�豸ȡֵ��ʹ�����ӣ�Ϊ��дswp_iog.devno';
COMMENT ON COLUMN ips_busi2intf_cfg.msgid IS '���ı��';
COMMENT ON COLUMN ips_busi2intf_cfg.orderno IS '����˳��';
COMMENT ON COLUMN ips_busi2intf_cfg.remark IS '��ע';

------------------------------------------------;
--8.1.5 : �ӿڱ������ñ�, ��ṹ;
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
COMMENT ON TABLE ips_message_cfg IS '�ӿڱ������ñ�';
COMMENT ON COLUMN ips_message_cfg.msgid IS '���ı��';
COMMENT ON COLUMN ips_message_cfg.msgname IS '��������';
COMMENT ON COLUMN ips_message_cfg.level IS '����00:�ͻ���;10:�û���(ҵ����Ҫ����servlist�ڵ�);20:��Ʒ��(һ����Ʒ��Ӧһ������)(ҵ����Ҫ����pordlist�ڵ�);21:��Ʒ��(�����Ʒ��Ӧһ������)(ҵ����Ҫ����pordlist�ڵ�)';
COMMENT ON COLUMN ips_message_cfg.msgtmp IS '����ģ��';
COMMENT ON COLUMN ips_message_cfg.nodes IS 'ģ��ʹ�õ�������';
COMMENT ON COLUMN ips_message_cfg.remark IS '����ע��';

------------------------------------------------;
--8.1.6 : ģ���������ñ�, ��ṹ;
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
--8.1.6 : ģ���������ñ� �����ֵ�;
-------------------------------------------------;
COMMENT ON TABLE ips_node_cfg IS 'ģ���������ñ�';
COMMENT ON COLUMN ips_node_cfg.nodeid IS '���';
COMMENT ON COLUMN ips_node_cfg.node IS '�ڵ�����';
COMMENT ON COLUMN ips_node_cfg.name IS '��������';
COMMENT ON COLUMN ips_node_cfg.type IS '��������,0:�����ڲ�������1:SQLֵ';
COMMENT ON COLUMN ips_node_cfg.indata IS '�ڲ�ֵ����,TypeΪ0ʱ��Ч';
COMMENT ON COLUMN ips_node_cfg.sqlid IS 'SQLID,TypeΪ1ʱ��Ч';
COMMENT ON COLUMN ips_node_cfg.sqlvalueid IS 'SQL�ĵڼ���ֵ,TypeΪ1ʱ��Ч';
COMMENT ON COLUMN ips_node_cfg.memo IS '��ע';


------------------------------------------------;
-- 8.1.7 : �������ñ�, ��ṹ;
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

COMMENT ON TABLE ips_nodesql_cfg IS '�������ñ�';
COMMENT ON COLUMN ips_nodesql_cfg.sqlid IS '���';
COMMENT ON COLUMN ips_nodesql_cfg.sqlname IS 'SQL��,SQL����ʹ���ڲ�ֵ���ţ�����¼���ڲ�ֵ���ţ�';
COMMENT ON COLUMN ips_nodesql_cfg.dblink IS '����Դ';
COMMENT ON COLUMN ips_nodesql_cfg.sql IS 'SQL���';
COMMENT ON COLUMN ips_nodesql_cfg.valuenum IS '��ѯ�������';
COMMENT ON COLUMN ips_nodesql_cfg.sqlparams IS 'SQL���,ֻ�����ڲ�ֵ';


