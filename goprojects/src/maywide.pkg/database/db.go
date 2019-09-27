// db-api for myslsnr[9527] or oralsnr[9527]
package database

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/axgle/mahonia"
	mylog "maywide.pkg/mylogger"
)

// default addr
var Addr string = "127.0.0.1:9527"
var Driver string = "mysql"
var Show bool = true

// 0 unchange, 1 lower, 2 title
var ColumnCase int = 0

type DB struct {
	Driver   string
	ADDR     string
	DSN      string
	Sid      string
	Sqlcode  int
	Sqlerrm  string
	Lasttime time.Time
	Errfunc  interface{}
	status   bool
	addr     *net.TCPAddr
	conn     *net.TCPConn
	data     *msg
}

type Rows struct {
	Curid  int
	Cols   []string
	status bool
	dbi    *DB
}

type Columns struct {
	Tabname string
	Prikey  string
	Cols    []string
}

type msg struct {
	sid     string
	dsn     string
	name    string
	reqSql  string
	resSql  string
	inVar   []string
	outVar  []string
	colName []string
	rows    int
	curId   int
	sqlCode string
	sqlErrm string
	req     string
	res     string
	conn    *net.TCPConn
}

func (m *msg) fieldAdd(dest string, name string, value string) string {
	head := fmt.Sprintf("%02d%04d", len(name), len(value))
	return dest + head + name + value
}

func (m *msg) fieldValue(src string, name string) string {
	l := len(src)
	i := 0
	e := 0

	for {
		e = i + 2
		namelen := src[i:e]

		i = e
		e = i + 4
		vallen := src[i:e]

		nl, _ := strconv.Atoi(namelen)
		vl, _ := strconv.Atoi(vallen)

		if nl <= 0 {
			break
		}

		i = e
		e = i + nl
		code := src[i:e]

		i = e
		e = i + vl
		value := src[i:e]
		if code == name {
			return value
		}

		i = e
		if i >= l-1 {
			break
		}
	}
	return ""
}

func (m *msg) comm() bool {
	m.req = fmt.Sprintf("%08d", len(m.req)) + m.req
	//mylog.Log("req=", m.req)
	n, err := m.conn.Write([]byte(m.req))
	if err != nil {
		m.sqlCode = "-10"
		m.sqlErrm = "conn.Write:" + err.Error()
		return false
	}

	if n < len(m.req) {
		newreq := m.req[n:len(m.req)]
		n, err = m.conn.Write([]byte(newreq))
		if err != nil {
			m.sqlCode = "-11"
			m.sqlErrm = "conn.Write:" + err.Error()
			return false
		}
	} else if n > len(m.req) {
		m.sqlCode = "-12"
		mylog.Log("m.req=", m.req)
		m.sqlErrm = "conn.Write: " + fmt.Sprint(len(m.req)) + ", effected: " + fmt.Sprint(n)
		return false
	}

	// get response head
	res := make([]byte, 8)
	n, err = m.conn.Read(res)
	if err != nil {
		m.sqlCode = "-21"
		m.sqlErrm = "conn.Read.head:" + err.Error()
		return false
	}

	// socket error
	if n <= 0 {
		m.sqlCode = "-23"
		m.sqlErrm = "conn.Read.head: socket was terminated."
		return false
	}

	// check length
	if n != 8 {
		m.sqlCode = "-22"
		m.sqlErrm = "conn.Read.head: 8, effected: " + fmt.Sprint(n)
		return false
	}

	// read data checking
	reslen, _ := strconv.Atoi(string(res))
	if reslen <= 16 {
		m.sqlCode = "-24"
		m.sqlErrm = "conn.Read.head: error [" + string(res) + "]"
		return false
	}

	// read data
	resdata := make([]byte, reslen)
	n, err = m.conn.Read(resdata)
	if err != nil {
		m.sqlCode = "-31"
		m.sqlErrm = "conn.Read.data:" + err.Error()
		return false
	}

	// socket error
	if n <= 0 {
		m.sqlCode = "-33"
		m.sqlErrm = "conn.Read.data: socket was terminated."
		return false
	}

	// check length
	if n != reslen {
		m.sqlCode = "-32"
		m.sqlErrm = "conn.Read.data: 8, effected: " + fmt.Sprint(n)
		return false
	}

	m.res = string(resdata)
	return true
}

func (m *msg) commPrepare() bool {
	if len(m.name) < 2 {
		return false
	}

	m.req = m.fieldAdd("", "name", m.name)
	if m.name == "DB_CONNECT" {
		m.req = m.fieldAdd(m.req, "login", m.dsn)
		return true
	} else if m.name == "DB_RELEASE" || m.name == "DB_COMMIT" || m.name == "DB_ROLLBACK" {
		return true
	}

	// auto add parameters for socket request
	if len(m.reqSql) > 6 {
		m.req = m.fieldAdd(m.req, "sql", m.reqSql)
	}

	// for cursor
	if m.curId > 0 {
		m.req = m.fieldAdd(m.req, "curid", fmt.Sprint(m.curId))
	}

	// for bind variables
	if len(m.inVar) > 0 {
		m.req = m.fieldAdd(m.req, "cols", fmt.Sprint(len(m.inVar)))
		for i := 0; i < len(m.inVar); i++ {
			m.req = m.fieldAdd(m.req, fmt.Sprint("value", i), m.inVar[i])
		}
	} else {
		m.req = m.fieldAdd(m.req, "cols", "0")
	}
	return true
}

func (m *msg) commParse() bool {
	code := m.fieldValue(m.res, "code")
	if m.name == "DB_SELECT" || m.name == "DB_PREPARE" ||
		m.name == "DB_SELECT_FIRST" || m.name == "DB_OPEN_CURSOR" {
		m.resSql = m.fieldValue(m.res, "sql")
		// show sql after sql parse or execution
		if Show && len(m.resSql) > 0 {
			mylog.Log("["+m.sid+"]SQL=", m.resSql)
		}
	}

	if code == "" {
		m.sqlCode = "-51"
		m.sqlErrm = "FATAL: response has no [code]."
		return false
	}
	m.sqlCode = m.fieldValue(m.res, "sqlcode")
	m.sqlErrm = m.fieldValue(m.res, "sqlerrm")

	if m.sqlCode != "0" {
		if m.sqlCode[0:1] != "-" {
			m.sqlCode = "-" + m.sqlCode
		}
		mylog.Println("ERR" + m.sqlCode + ": " + m.sqlErrm)
	}

	// return value
	codeValue, _ := strconv.Atoi(code)

	m.curId = 0
	if m.name == "DB_OPEN_CURSOR" {
		m.curId = codeValue
		code := m.fieldValue(m.res, "cols")
		codeValue, _ = strconv.Atoi(code)
	} else if m.name == "DB_EXECUTE" {
		m.rows = codeValue
	}

	if m.name == "DB_FETCH_RECORD" || m.name == "DB_SELECT" || m.name == "DB_SELECT_FIRST" {
		m.outVar = nil
		for i := 0; i < codeValue; i++ {
			vName := fmt.Sprint("value", i)
			vStr := m.fieldValue(m.res, vName)
			m.outVar = append(m.outVar, mahonia.NewDecoder("gbk").ConvertString(vStr))
		}
	} else if m.name == "DB_COLUMNS" || m.name == "DB_OPEN_CURSOR" {
		m.colName = nil
		for i := 0; i < codeValue; i++ {
			vName := fmt.Sprint("value", i)
			vStr := m.fieldValue(m.res, vName)
			if ColumnCase == 1 {
				vStr = strings.ToLower(vStr)
			} else if ColumnCase == 2 {
				vStr = strings.ToTitle(strings.ToLower(vStr))
			}
			m.colName = append(m.colName, vStr)
		}
	}
	return true
}

func (dbi *DB) parseError() {
	dbi.Sqlcode, _ = strconv.Atoi(dbi.data.sqlCode)
	dbi.Sqlerrm = dbi.data.sqlErrm

	if dbi.Sqlcode != 0 && dbi.Errfunc != nil {
		inparam := make([]reflect.Value, 2)
		inparam[0] = reflect.ValueOf(dbi.Sqlcode)
		inparam[1] = reflect.ValueOf(dbi.Sqlerrm)
		reflect.ValueOf(dbi.Errfunc).Call(inparam)
	}
}

func Connect(dsn string, repeat int) *DB {
	dbi := &DB{}
	dbi.ADDR = Addr
	dbi.DSN = dsn
	dbi.status = false

	i := 0
	for {
		err := dbi.Connect()
		if err == nil {
			break
		} else {
			if Show == false {
				mylog.Println(dbi.Sqlerrm)
			}
			time.Sleep(time.Second * 3)
		}
		i++
		if repeat != -1 && i >= repeat {
			break
		}
	}
	return dbi
}

func (dbi *DB) Connect() error {
	if dbi.status {
		return nil
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp4", dbi.ADDR)
	if err != nil {
		return errors.New("Connect.ResolveTCPAddr: " + err.Error())
	}

	dbi.addr = tcpAddr
	dbi.conn, err = net.DialTCP("tcp", nil, dbi.addr)
	if err != nil {
		return errors.New("Connect.DialTCP: " + err.Error())
	}

	dbi.conn.SetKeepAlive(true)
	dbi.data = new(msg)
	// create msg struct and handle it
	dbi.createOperation("DB_CONNECT")
	// combine data to be sent
	dbi.data.commPrepare()
	// send and receive
	if dbi.data.comm() == false {
		dbi.Sqlcode, _ = strconv.Atoi(dbi.data.sqlCode)
		dbi.Sqlerrm = dbi.data.sqlErrm
		return errors.New(dbi.Sqlerrm)
	}

	if dbi.data.commParse() == false {
		dbi.Sqlcode, _ = strconv.Atoi(dbi.data.sqlCode)
		dbi.Sqlerrm = dbi.data.sqlErrm
		return errors.New(dbi.Sqlerrm)
	}
	dbi.parseError()

	if dbi.Sqlcode == 0 {
		dbi.status = true
		if Driver == "mysql" {
			dbi.Sid, _ = dbi.Get("SELECT CONNECTION_ID()")
		} else if Driver == "oracle" {
			dbi.Sid, _ = dbi.Get("SELECT sid FROM v$mystat WHERE rownum=1")
		}
	} else {
		return errors.New(dbi.Sqlerrm)
	}
	return nil
}

// define operations
func (dbi *DB) createOperation(name string, sql ...string) *DB {
	dbi.data.dsn = dbi.DSN
	dbi.data.name = name
	dbi.data.reqSql = ""
	if len(sql) > 0 {
		dbi.data.reqSql = sql[0]
	}
	dbi.data.sid = dbi.Sid
	dbi.data.conn = dbi.conn
	dbi.data.reqSql = ""
	dbi.data.resSql = ""
	dbi.data.curId = 0
	dbi.data.req = ""
	dbi.data.inVar = nil
	dbi.Lasttime = time.Now()
	return dbi
}

// disconnect from database
func (dbi *DB) Disconnect() {
	if dbi.status {
		dbi.createOperation("DB_RELEASE")
		dbi.data.commPrepare()
		dbi.data.comm()
		dbi.data.commParse()
		dbi.conn.Close()
		dbi.status = false
		dbi.data.conn = nil
		dbi.data = nil
		dbi.addr = nil
	}
}

// disconnect from database
func (dbi *DB) Close() {
	dbi.conn.Close()
	dbi.status = false
	//dbi.data.conn = nil
	dbi.data = nil
	dbi.addr = nil
}

// commit data into database
func (dbi *DB) Commit() {
	if dbi.status {
		dbi.createOperation("DB_COMMIT")
		dbi.data.commPrepare()
		dbi.data.comm()
		dbi.data.commParse()
		if Show {
			mylog.Println("[" + dbi.Sid + "]SQL=COMMIT;")
		}
	}
}

// commit data into database
func (dbi *DB) Rollback() {
	if dbi.status {
		dbi.createOperation("DB_ROLLBACK")
		dbi.data.commPrepare()
		dbi.data.comm()
		dbi.data.commParse()
		if Show {
			mylog.Println("[" + dbi.Sid + "]SQL=ROLLBACK;")
		}
	}
}

func (dbi *DB) Ping() error {
	if dbi.status == false {
		dbi.Connect()
		if dbi.Sqlcode != 0 {
			return errors.New(dbi.Sqlerrm)
		}
	}

	if Driver == "mysql" {
		_, err := dbi.Get("SELECT NOW()")
		return err
	} else if Driver == "oracle" {
		_, err := dbi.Get("SELECT SYSDATE FROM DUAL")
		return err
	}
	return nil
}

// prepare sql for cursor or executions
func (dbi *DB) Prepare(sqls string, arg ...string) {
	if dbi.status == false {
		return
	}

	dbi.createOperation("DB_PREPARE")
	// bind variables
	dbi.data.reqSql = sqls
	dbi.data.inVar = arg
	// data
	dbi.data.commPrepare()
	// communicate
	if dbi.data.comm() == false {
		dbi.Close()
	}
	// parse
	dbi.data.commParse()
}

// execute dml/ddl sql
func (dbi *DB) Exec(sqls string, arg ...string) (int, error) {
	if dbi.status == false {
		return -1, errors.New("database was not opened.")
	}

	dbi.Prepare(sqls, arg...)
	// start operation
	dbi.createOperation("DB_EXECUTE")
	dbi.data.reqSql = sqls
	// data
	dbi.data.commPrepare()
	// communicate
	if dbi.data.comm() == false {
		dbi.Close()
		return -1, errors.New(dbi.Sqlerrm)
	}
	// parse
	dbi.data.commParse()
	dbi.parseError()
	if dbi.Sqlcode != 0 {
		return -1, errors.New(dbi.Sqlerrm)
	}
	return dbi.data.rows, nil
}

// select one-row-with-all-columns by all binded variables
func (dbi *DB) SelectRow(sqls string, arg ...string) (map[string]string, error) {
	if dbi.status == false {
		return nil, errors.New("database was not opened.")
	}

	// start select
	dbi.createOperation("DB_SELECT")
	dbi.data.reqSql = sqls
	dbi.data.inVar = arg
	// data
	dbi.data.commPrepare()
	// communicate
	if dbi.data.comm() == false {
		mylog.Println(dbi.data.sqlCode, ":", dbi.data.sqlErrm)
		dbi.Close()
		return nil, errors.New(dbi.Sqlerrm)
	}
	// parse
	dbi.data.commParse()
	dbi.parseError()
	if dbi.Sqlcode != 0 {
		return nil, errors.New(dbi.Sqlerrm)
	}

	// get column name list
	dbi.createOperation("DB_COLUMNS")
	// data
	dbi.data.curId = -1
	dbi.data.commPrepare()
	// communicate
	if dbi.data.comm() == false {
		dbi.Close()
		return nil, errors.New(dbi.Sqlerrm)
	}
	// parse
	dbi.data.commParse()
	// no need to change dbi.sqlcode

	colvar := make(map[string]string)
	for i := 0; i < len(dbi.data.colName); i++ {
		colvar[strings.ToLower(dbi.data.colName[i])] = dbi.data.outVar[i]
	}
	return colvar, nil
}

// open cursor for select statement
func (dbi *DB) Cursor(sqls string, arg ...string) (*Rows, error) {
	if dbi.status == false {
		return nil, errors.New("database was not opened.")
	}
	dbi.Prepare(sqls, arg...)

	// start cursor
	dbi.createOperation("DB_OPEN_CURSOR")
	dbi.data.reqSql = sqls
	// data
	dbi.data.commPrepare()
	// communicate
	if dbi.data.comm() == false {
		dbi.Close()
		return nil, errors.New(dbi.Sqlerrm)
	}
	// parse
	dbi.data.commParse()
	dbi.parseError()

	if dbi.Sqlcode < 0 {
		return nil, errors.New(dbi.Sqlerrm)
	}

	if dbi.data.curId <= 0 {
		return nil, errors.New("No cursor created.")
	}

	rows := new(Rows)
	rows.Curid = dbi.data.curId
	rows.Cols = make([]string, len(dbi.data.colName))
	rows.status = true
	rows.dbi = dbi

	for i := 0; i < len(dbi.data.colName); i++ {
		rows.Cols[i] = strings.ToLower(dbi.data.colName[i])
	}
	return rows, nil
}

// open cursor for select statement
func (rows *Rows) Next() ([]string, error) {
	if rows.status == false {
		return nil, errors.New("Cursor not opened.")
	}

	// start cursor
	rows.dbi.createOperation("DB_FETCH_RECORD")
	rows.dbi.data.curId = rows.Curid
	// data
	rows.dbi.data.commPrepare()
	// communicate
	if rows.dbi.data.comm() == false {
		rows.dbi.Close()
		rows.status = false
		return nil, errors.New(rows.dbi.Sqlerrm)
	}
	// parse
	rows.dbi.data.commParse()
	rows.dbi.parseError()

	if rows.dbi.Sqlcode < 0 {
		return nil, errors.New(rows.dbi.Sqlerrm)
	}

	return rows.dbi.data.outVar, nil
}

// close cursor
func (rows *Rows) Close() {
	// start cursor
	rows.dbi.createOperation("DB_CLOSE_CURSOR")
	rows.dbi.data.curId = rows.Curid
	// data
	rows.dbi.data.commPrepare()
	// communicate
	if rows.dbi.data.comm() == false {
		rows.dbi.Close()
		rows.status = false
	}
	// parse
	rows.dbi.data.commParse()
	rows.dbi.parseError()
	rows.status = false
}

// select limited rows from sql
func (dbi *DB) SelectRows(nums int, sqls string, arg ...string) ([]map[string]string, error) {
	if dbi.status == false {
		return nil, errors.New("database was not opened.")
	}

	rows, err := dbi.Cursor(sqls, arg...)
	if err != nil {
		return nil, errors.New("Cannot open cursor for sql: " + sqls)
	}

	var resvar []map[string]string
	limit := 0
	for {
		colvar, rowerr := rows.Next()
		if rowerr != nil {
			rows.Close()
			return nil, errors.New("Cannot fetch records for sql: " + sqls)
		} else if colvar == nil && rowerr == nil {
			// no records found, lsnr had been release automatically
			rows.status = false
			break
		}

		line := make(map[string]string)
		for i := 0; i < len(rows.Cols); i++ {
			line[rows.Cols[i]] = colvar[i]
		}
		resvar = append(resvar, line)
		limit++
		if nums > 0 && limit >= nums {
			rows.Close()
			break
		}
	}
	return resvar, nil
}

// select all rows from sql
func (dbi *DB) Select(sqls string, arg ...string) ([]map[string]string, error) {
	if dbi.status == false {
		return nil, errors.New("database was not opened.")
	}

	rows, err := dbi.Cursor(sqls, arg...)
	if err != nil {
		return nil, errors.New("Cannot open cursor for sql: " + sqls)
	}

	var resvar []map[string]string
	for {
		colvar, rowerr := rows.Next()
		if rowerr != nil {
			rows.Close()
			return nil, errors.New("Cannot fetch records for sql: " + sqls)
		} else if colvar == nil && rowerr == nil {
			// no records found
			rows.status = false
			break
		}

		line := make(map[string]string)
		for i := 0; i < len(rows.Cols); i++ {
			line[rows.Cols[i]] = colvar[i]
		}
		resvar = append(resvar, line)
	}
	return resvar, nil
}

// select all rows from sql into map by column value, two columns only
func (dbi *DB) SelectMap(sqls string, arg ...string) (map[string]string, error) {
	if dbi.status == false {
		return nil, errors.New("database was not opened.")
	}

	rows, err := dbi.Cursor(sqls, arg...)
	if err != nil {
		return nil, errors.New("Cannot open cursor for sql: " + sqls)
	}

	if len(rows.Cols) != 2 {
		rows.Close()
		return nil, errors.New("Error: only two columns allowed.")
	}

	resvar := make(map[string]string)
	for {
		colvar, rowerr := rows.Next()
		if rowerr != nil {
			rows.Close()
			return nil, errors.New("Cannot fetch records for sql: " + sqls)
		} else if colvar == nil && rowerr == nil {
			// no records found
			rows.status = false
			break
		}

		resvar[colvar[0]] = colvar[1]
	}
	return resvar, nil
}

// return list of array for select-statement
func (dbi *DB) SelectList(sqls string, arg ...string) ([]string, error) {
	if dbi.status == false {
		return nil, errors.New("database was not opened.")
	}

	rows, err := dbi.Cursor(sqls, arg...)
	if err != nil {
		return nil, errors.New("Cannot open cursor for sql: " + sqls)
	}

	resvar := make([]string, 0)
	for {
		colvar, rowerr := rows.Next()
		if rowerr != nil {
			rows.Close()
			return nil, errors.New("Cannot fetch records for sql: " + sqls)
		} else if colvar == nil && rowerr == nil {
			// no records found
			rows.status = false
			break
		}

		line := colvar[0]
		for i := 1; i < len(rows.Cols); i++ {
			line = line + "," + colvar[i]
		}
		resvar = append(resvar, line)
	}

	if len(resvar) == 0 {
		return nil, nil
	}
	return resvar, nil
}

// select one value only
func (dbi *DB) Get(sqls string, arg ...string) (string, error) {
	if dbi.status == false {
		return "", errors.New("database was not opened.")
	}

	// start select
	dbi.createOperation("DB_SELECT")
	dbi.data.reqSql = sqls
	// data
	dbi.data.commPrepare()
	// communicate
	if dbi.data.comm() == false {
		dbi.Close()
		return "", errors.New(dbi.Sqlerrm)
	}
	// parse
	dbi.data.commParse()
	dbi.parseError()

	if dbi.Sqlcode != 0 {
		return "", errors.New(dbi.Sqlerrm)
	}

	if len(dbi.data.outVar) > 0 {
		return dbi.data.outVar[0], nil
	}
	return "", nil
}

// get columns info for table
func (dbi *DB) TableColumns(tabname string) *Columns {
	cols := new(Columns)
	cols.Tabname = strings.ToLower(tabname)

	sqls := "SELECT lower(column_name) FROM information_schema.columns WHERE table_name=UPPER(:1) AND table_schema=database()"
	if Driver == "oracle" {
		sqls = "SELECT lower(column_name) FROM user_tab_columns WHERE table_name=UPPER(:1) ORDER BY column_id"
	}

	var err error
	cols.Cols, err = dbi.SelectList(sqls, cols.Tabname)
	if err != nil || cols.Cols == nil {
		return nil
	}

	if Driver == "mysql" {
		cols.Prikey = cols.Cols[0]
	} else if Driver == "oracle" {
		sqls = "SELECT lower(column_name) FROM user_ind_columns WHERE table_name=UPPER(:1) AND index_name IN (SELECT constraint_name FROM user_constraints WHERE table_name=UPPER(:2) AND constraint_type='P') AND ROWNUM=1"
		cols.Prikey, err = dbi.Get(sqls, cols.Tabname, cols.Tabname)
		if err != nil {
			cols.Prikey = cols.Cols[0]
		}
	}
	return cols
}

// get columns info for table
func (dbi *DB) TableExists(tabname string) bool {
	sqls := "SELECT 1 FROM information_schema.tables WHERE table_name=UPPER(:1) AND table_schema=database()"
	if Driver == "oracle" {
		sqls = "SELECT 1 FROM user_tables WHERE table_name=UPPER(:1)"
	}

	tabNums, _ := dbi.Get(sqls, tabname)
	if tabNums == "1" {
		return true
	}
	return false
}

type Table struct {
	db        *DB
	tablename string
	param     []string
	columnstr string
	where     string
	pk        string
	orderby   string
	limit     string
	join      string
}

func (dbi *DB) Table(tabname string) *Table {
	m := new(Table)
	m.tablename = tabname
	m.columnstr = "*"
	m.db = dbi
	return m
}

func (m *Table) Table(tablename string) *Table {
	m.tablename = tablename
	return m
}

func (m *Table) FindRows(rows int) []map[string]string {
	if m.db == nil {
		return nil
	}

	if len(m.param) == 0 {
		m.columnstr = "*"
	} else {
		if len(m.param) == 1 {
			m.columnstr = m.param[0]
		} else {
			m.columnstr = strings.Join(m.param, ",")
		}
	}

	query := fmt.Sprintf("SELECT %v FROM %v %v %v %v %v", m.columnstr, m.tablename, m.join, m.where, m.orderby, m.limit)
	result, err := m.db.SelectRows(rows, query)
	if err != nil {
		return nil
	}
	return result
}

func (m *Table) FindAll() []map[string]string {
	return (m.FindRows(0))
}

func (m *Table) FindOne() []map[string]string {
	return (m.Limit(1).FindRows(1))
}

func (m *Table) Where(param string) *Table {
	m.where = fmt.Sprintf(" WHERE %v", param)
	return m
}

func (m *Table) SetPk(pk string) *Table {
	m.pk = pk
	return m
}

func (m *Table) OrderBy(param string) *Table {
	m.orderby = fmt.Sprintf("ORDER BY %v", param)
	return m
}

func (m *Table) Limit(size ...int) *Table {
	var end int
	start := size[0]
	if len(size) > 1 {
		end = size[1]
		m.limit = fmt.Sprintf("LIMIT %d,%d", start, end)
		return m
	}
	m.limit = fmt.Sprintf("LIMIT %d", start)
	return m
}

func (m *Table) LeftJoin(table, condition string) *Table {
	m.join = fmt.Sprintf("LEFT JOIN %v ON %v", table, condition)
	return m
}

func (m *Table) RightJoin(table, condition string) *Table {
	m.join = fmt.Sprintf("RIGHT JOIN %v ON %v", table, condition)
	return m
}

func (m *Table) Join(table, condition string) *Table {
	m.join = fmt.Sprintf("INNER JOIN %v ON %v", table, condition)
	return m
}

func (m *Table) FullJoin(table, condition string) *Table {
	m.join = fmt.Sprintf("FULL JOIN %v ON %v", table, condition)
	return m
}

func (m *Table) Insert(param map[string]interface{}) int {
	if m.db == nil {
		return -2
	}

	var keys []string
	var values []string

	for key, value := range param {
		keys = append(keys, strings.ToLower(key))
		switch value.(type) {
		case int, int64, int32:
			values = append(values, strconv.Itoa(value.(int)))
		case string:
			values = append(values, value.(string))
		case float32, float64:
			values = append(values, strconv.FormatFloat(value.(float64), 'f', -1, 64))
		}
	}
	valueList := "'" + strings.Join(values, "','") + "'"
	fieldList := "`" + strings.Join(keys, "`,`") + "`"
	sql := fmt.Sprintf("INSERT INTO %v (%v) VALUES (%v)", m.tablename, fieldList, valueList)
	rows, err := m.db.Exec(sql)
	if err != nil {
		return -1
	}
	return rows
}

func (m *Table) Delete(param string) int {
	if m.db == nil {
		return -2
	}

	sql := fmt.Sprintf("DELETE FROM %v WHERE %v", m.tablename, param)
	rows, err := m.db.Exec(sql)
	if err != nil {
		return -1
	}
	return rows
}

func (m *Table) Update(param map[string]interface{}) int {
	if m.db == nil {
		return -2
	}
	var setValue []string
	for key, value := range param {
		switch value.(type) {
		case int, int64, int32:
			set := fmt.Sprintf("%v = %v", key, value.(int))
			setValue = append(setValue, set)
		case string:
			set := fmt.Sprintf("%v = '%v'", key, value.(string))
			setValue = append(setValue, set)
		case float32, float64:
			set := fmt.Sprintf("%v = '%v'", key, strconv.FormatFloat(value.(float64), 'f', -1, 64))
			setValue = append(setValue, set)
		}

	}
	setData := strings.Join(setValue, ",")
	sql := fmt.Sprintf("UPDATE %v SET %v %v", m.tablename, setData, m.where)
	result, err := m.db.Exec(sql)
	if err != nil {
		return -1
	}
	return result
}
