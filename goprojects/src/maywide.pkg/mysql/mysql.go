package mysql

/*
#cgo CFLAGS: -IC:/golang/mingw64/include
#cgo LDFLAGS: -LC:/golang/mingw64/lib -lcymysql
#include <boss.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"github.com/axgle/mahonia"
)

type DB struct {
	NAME     string
	isclose  bool
	DSN      string
	debug    int
	linkaddr *_Ctype_char
	Sqlcode  int
	Sqlerrm  string
	LowerCol int
	Errfunc  interface{}
}

type Rows struct {
	Curid    int
	Cols     []string
	isclose  bool
	linkaddr *_Ctype_char
}

type Columns struct {
	Tabname string
	Prikey  string
	Cols    []string
}

var db_mutex sync.Mutex

// open trace for cgo
func Topen(filename string) {
	if filename == "" || filename == "/dev/null" {
		return
	}
	name := C.CString(filename)
	defer C.free(unsafe.Pointer(name))
	C.topen(name)
	return
}

// set database tag
func SetDatabase(linkaddr *_Ctype_char) {
	C.GO_SETDB(linkaddr)
	return
}

// Trace for CheckError
func Println(arg ...interface{}) {
	s := fmt.Sprint(arg...)
	p := C.CString(s + "\n")
	defer C.free(unsafe.Pointer(p))
	C.GO_LOG(p)
	return
}

// putenv for cgo
func Iputenv(filename string, secname string) {
	name := C.CString(filename)
	defer C.free(unsafe.Pointer(name))

	secs := C.CString(secname)
	defer C.free(unsafe.Pointer(secs))

	C.iputenv(name, secs)
	return
}

// lock for global variables inside api
func Lock() {
	db_mutex.Lock()
}

// unlock for global variables outside api
func Unlock() {
	db_mutex.Unlock()
}

// lock for global variables inside api
func (dbi *DB) Lock() {
	db_mutex.Lock()
}

// unlock for global variables outside api
func (dbi *DB) Unlock() {
	db_mutex.Unlock()
}

func (dbi *DB) Debug(on int) *DB {
	C.GO_TRACE(C.int(on))
	dbi.debug = on
	return dbi
}

// commit data into database
func (dbi *DB) Commit() {
	dbi.Lock()
	defer dbi.Unlock()
	SetDatabase(dbi.linkaddr)
	C.DB_COMMIT()
	return
}

// commit data into database
func (dbi *DB) Rollback() {
	dbi.Lock()
	defer dbi.Unlock()
	SetDatabase(dbi.linkaddr)
	C.DB_ROLLBACK()
	return
}

// get last sql status
func (dbi *DB) GetError(msg ...string) int {
	if len(msg) == 0 {
		nbyte := make([]byte, 512)
		pchar := C.CString(string(nbyte))
		defer C.free(unsafe.Pointer(pchar))
		pvar := (*_Ctype_char)(unsafe.Pointer(pchar))
		code := C.DB_SQLERRM(pvar)
		dbi.Sqlcode = int(code)
		dbi.Sqlerrm = C.GoString(pvar)
	} else {
		if msg[0] == "1403" {
			dbi.Sqlcode = 1403
			dbi.Sqlerrm = "No data found"
		}
	}

	if dbi.Sqlcode != 0 && dbi.Errfunc != nil {
		inparam := make([]reflect.Value, 2)
		inparam[0] = reflect.ValueOf(dbi.Sqlcode)
		inparam[1] = reflect.ValueOf(dbi.Sqlerrm)
		reflect.ValueOf(dbi.Errfunc).Call(inparam)
	}
	return dbi.Sqlcode
}

// connect to oracle by dsn, retry [repeat] times when failed
func Connect(dsn string, repeat int) *DB {
	Lock()
	defer Unlock()
	os.Setenv("NLS_LANG", "AMERICAN_AMERICA.UTF8")
	dbi := new(DB)
	dbi.DSN = dsn
	dbi.NAME = ""
	dbi.linkaddr = C.CString(dbi.NAME)
	dbi.LowerCol = 1
	login := C.CString(dsn)
	defer C.free(unsafe.Pointer(login))
	// set database address
	SetDatabase(dbi.linkaddr)
	// connect to oracle
	i := C.DB_CONNECT(login, C.int(repeat))
	if i < 0 {
		fmt.Println("Cannot connect to oracle")
		os.Exit(1)
	}
	dbi.isclose = false
	dbi.Debug(1)
	return dbi
}

// connect to oracle by dsn, retry [repeat] times when failed
func (dbi *DB) Connect() *DB {
	dbi.Lock()
	defer dbi.Unlock()
	os.Setenv("NLS_LANG", "AMERICAN_AMERICA.UTF8")
	// dsn is not set
	if dbi.DSN == "" {
		Println("dbi.dsn was not set")
		dbi.isclose = true
		return dbi
	}
	// set database address
	dbi.linkaddr = C.CString(dbi.NAME)
	SetDatabase(dbi.linkaddr)

	login := C.CString(dbi.DSN)
	defer C.free(unsafe.Pointer(login))
	// connect to oracle
	i := C.DB_CONNECT(login, C.int(-1))
	if i < 0 {
		fmt.Println("Cannot connect to oracle")
		os.Exit(1)
	}
	dbi.isclose = false
	dbi.LowerCol = 1
	dbi.Debug(1)
	return dbi
}

// disconnect from oracle
func (dbi *DB) Disconnect() {
	dbi.Lock()
	defer dbi.Unlock()
	if dbi.isclose {
		return
	}
	SetDatabase(dbi.linkaddr)
	C.DB_RELEASE()
	C.free(unsafe.Pointer(dbi.linkaddr))
	dbi.isclose = true
	return
}

// disconnect from oracle
func (dbi *DB) Close() {
	dbi.Disconnect()
}

// check database time for status
func (dbi *DB) Ping() error {
	dbi.Lock()
	defer dbi.Unlock()
	SetDatabase(dbi.linkaddr)
	if dbi.isclose {
		return nil
	}
	i := C.DB_CHECK()
	if i < 0 {
		dbi.GetError()
		return errors.New("Select sysdate from database failed")
	}
	return nil
}

// variablity to slice for cgo
func (dbi *DB) Prepare(sqls string, arg ...string) {
	psql := C.CString(sqls)
	defer C.free(unsafe.Pointer(psql))

	inlen := len(arg)
	if inlen == 0 {
		C.GO_PREPARE(psql, nil, C.int(0))
		return
	}

	// array content for input string
	invar := make([](*_Ctype_char), 0)
	for i := 0; i < inlen; i++ {
		pchar := C.CString(mahonia.NewEncoder("gbk").ConvertString(arg[i]))
		defer C.free(unsafe.Pointer(pchar))
		pvar := (*_Ctype_char)(unsafe.Pointer(pchar))
		// add *char into *char-list
		invar = append(invar, pvar)
	}
	C.GO_PREPARE(psql, (**_Ctype_char)(unsafe.Pointer(&invar[0])), C.int(inlen))
	invar = nil
}

// execute dml/ddl sql
func (dbi *DB) Exec(sqls string, arg ...string) (int, error) {
	dbi.Lock()
	defer dbi.Unlock()
	SetDatabase(dbi.linkaddr)
	if dbi.isclose {
		return -1, errors.New("database was not opened.")
	}

	dbi.Prepare(sqls, arg...)
	psql := C.CString(sqls)
	defer C.free(unsafe.Pointer(psql))
	dbi.Sqlcode = 0
	dbi.Sqlerrm = "OKAY"
	rows := C.DB_EXECUTE(psql)
	if int(rows) < 0 {
		dbi.GetError()
		return -1, errors.New("Cannot execute sql: " + sqls)
	}
	return int(rows), nil
}

// select one-row-with-all-columns by all binded variables
func (dbi *DB) SelectRow(sqls string, arg ...string) (map[string]string, error) {
	dbi.Lock()
	defer dbi.Unlock()
	SetDatabase(dbi.linkaddr)
	if dbi.isclose {
		return nil, errors.New("database was not opened.")
	}

	dbi.Prepare(sqls, arg...)
	// maximun output columns
	inlen := 128
	// array content for input string
	invar := make([](*_Ctype_char), 0)
	for i := 0; i < inlen; i++ {
		nbyte := make([]byte, 4096)
		pchar := C.CString(string(nbyte))
		defer C.free(unsafe.Pointer(pchar))
		pvar := (*_Ctype_char)(unsafe.Pointer(pchar))
		// add *char into *char-list
		invar = append(invar, pvar)
	}

	psql := C.CString(sqls)
	defer C.free(unsafe.Pointer(psql))

	rows := C.GO_SELECT(psql, (**_Ctype_char)(unsafe.Pointer(&invar[0])), C.int(inlen))
	if rows <= 0 {
		if rows == -1403 {
			dbi.GetError("1403")
		} else {
			dbi.GetError()
		}
		invar = nil
		return nil, errors.New("Cannot select columns values")
	}

	if int(rows) > 128 {
		invar = nil
		return nil, errors.New("Oracle.select.exception found.")
	}

	reslist := make([]string, int(rows))
	for i := 0; i < int(rows); i++ {
		reslist[i] = mahonia.NewDecoder("gbk").ConvertString(C.GoString(invar[i]))
	}

	namerows := C.GO_COLUMNS(-1, (**_Ctype_char)(unsafe.Pointer(&invar[0])), rows)
	if namerows <= 0 {
		dbi.GetError()
		invar = nil
		reslist = nil
		return nil, errors.New("Cannot select columns names")
	}

	colvar := make(map[string]string)
	for i := 0; i < int(rows); i++ {
		colvar[strings.ToLower(C.GoString(invar[i]))] = reslist[i]
	}
	invar = nil
	reslist = nil
	return colvar, nil
}

// open cursor for select statement
func (dbi *DB) Cursor(sqls string, arg ...string) (*Rows, error) {
	dbi.Lock()
	defer dbi.Unlock()
	SetDatabase(dbi.linkaddr)
	if dbi.isclose {
		return nil, errors.New("database was not opened.")
	}

	dbi.Prepare(sqls, arg...)
	psql := C.CString(sqls)
	defer C.free(unsafe.Pointer(psql))

	curid := C.GO_CURSOR(psql)
	if int(curid) < 0 {
		dbi.GetError()
		return nil, errors.New(fmt.Sprint("Cannot open cursor for sql, curid=", int(curid)))
	}

	rows := new(Rows)
	rows.Curid = int(curid)
	rows.isclose = false
	rows.linkaddr = dbi.linkaddr

	//get all column name into rows
	inlen := 128
	// array content for input string
	invar := make([](*_Ctype_char), 0)
	for i := 0; i < inlen; i++ {
		nbyte := make([]byte, 36)
		pchar := C.CString(string(nbyte))
		defer C.free(unsafe.Pointer(pchar))
		pvar := (*_Ctype_char)(unsafe.Pointer(pchar))
		// add *char into *char-list
		invar = append(invar, pvar)
	}

	colnums := C.GO_COLUMNS(curid, (**_Ctype_char)(unsafe.Pointer(&invar[0])), C.int(inlen))
	if int(colnums) < 0 {
		dbi.GetError()
		invar = nil
		return nil, errors.New("Cannot fetch columns name.")
	}

	reslist := make([]string, int(colnums))
	for i := 0; i < int(colnums); i++ {
		if dbi.LowerCol == 1 {
			reslist[i] = strings.ToLower(C.GoString(invar[i]))
		} else {
			reslist[i] = C.GoString(invar[i])
		}
	}

	rows.Cols = reslist
	reslist = nil
	invar = nil
	return rows, nil
}

// close cursor
func (rows *Rows) Close() {
	Lock()
	defer Unlock()
	if rows.isclose {
		return
	}
	SetDatabase(rows.linkaddr)
	rows.isclose = true
	C.GO_CLOSE(C.int(rows.Curid))
	Println("sql.cursor.id[", rows.Curid, "].release.manually.\n")
}

// open cursor for select statement
func (rows *Rows) Next() ([]string, error) {
	Lock()
	defer Unlock()
	if rows.isclose {
		return nil, errors.New("Cursor had been closed" + fmt.Sprintf("[%d].", rows.Curid))
	}

	if len(rows.Cols) == 0 {
		rows.Close()
		return nil, errors.New("Cursor has no columns" + fmt.Sprintf("[%d].", rows.Curid))
	}
	SetDatabase(rows.linkaddr)

	//get all column name into rows
	inlen := len(rows.Cols)
	// array content for input string
	invar := make([](*_Ctype_char), 0)
	for i := 0; i < inlen; i++ {
		nbyte := make([]byte, 4096)
		pchar := C.CString(string(nbyte))
		defer C.free(unsafe.Pointer(pchar))
		pvar := (*_Ctype_char)(unsafe.Pointer(pchar))
		// add *char into *char-list
		invar = append(invar, pvar)
	}

	colnums := C.GO_NEXT(C.int(rows.Curid), (**_Ctype_char)(unsafe.Pointer(&invar[0])), C.int(inlen))
	if int(colnums) == 0 {
		invar = nil
		rows.isclose = true
		return nil, nil
	} else if int(colnums) < 0 {
		invar = nil
		rows.isclose = true
		return nil, errors.New("Cannot fetch columns name.")
	}

	reslist := make([]string, int(colnums))
	for i := 0; i < int(colnums); i++ {
		reslist[i] = mahonia.NewDecoder("gbk").ConvertString(C.GoString(invar[i]))
	}
	invar = nil
	return reslist, nil
}

// select limited rows from sql
func (dbi *DB) SelectRows(nums int, sqls string, arg ...string) ([]map[string]string, error) {
	if dbi.isclose {
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
			rows.isclose = true
			return nil, errors.New("Cannot fetch records for sql: " + sqls)
		} else if colvar == nil && rowerr == nil {
			// no records found
			rows.isclose = true
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
	if dbi.isclose {
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
			rows.isclose = true
			return nil, errors.New("Cannot fetch records for sql: " + sqls)
		} else if colvar == nil && rowerr == nil {
			// no records found
			rows.isclose = true
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
	if dbi.isclose {
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
			rows.isclose = true
			return nil, errors.New("Cannot fetch records for sql: " + sqls)
		} else if colvar == nil && rowerr == nil {
			// no records found
			rows.isclose = true
			break
		}

		resvar[colvar[0]] = colvar[1]
	}
	return resvar, nil
}

// return list of array for select-statement
func (dbi *DB) SelectList(sqls string, arg ...string) ([]string, error) {
	if dbi.isclose {
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
			rows.isclose = true
			return nil, errors.New("Cannot fetch records for sql: " + sqls)
		} else if colvar == nil && rowerr == nil {
			// no records found
			rows.isclose = true
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
	dbi.Lock()
	defer dbi.Unlock()
	SetDatabase(dbi.linkaddr)
	colvar := ""
	if dbi.isclose {
		return colvar, errors.New("database was not opened.")
	}

	dbi.Prepare(sqls, arg...)

	// maximun output columns
	inlen := 128
	// array content for input string
	invar := make([](*_Ctype_char), 0)
	for i := 0; i < inlen; i++ {
		nbyte := make([]byte, 4096)
		pchar := C.CString(string(nbyte))
		defer C.free(unsafe.Pointer(pchar))
		pvar := (*_Ctype_char)(unsafe.Pointer(pchar))
		// add *char into *char-list
		invar = append(invar, pvar)
	}

	psql := C.CString(sqls)
	defer C.free(unsafe.Pointer(psql))

	rows := C.GO_SELECT(psql, (**_Ctype_char)(unsafe.Pointer(&invar[0])), C.int(inlen))
	if rows <= 0 {
		dbi.GetError()
		invar = nil
		return colvar, errors.New("Cannot select columns values")
	}

	if int(rows) > 128 {
		invar = nil
		return colvar, errors.New("select.exception found.")
	}

	colvar = C.GoString(invar[0])
	invar = nil
	return colvar, nil
}

// get columns info for table
func (dbi *DB) TableColumns(tabname string) *Columns {
	cols := new(Columns)
	cols.Tabname = strings.ToLower(tabname)

	var err error
	cols.Cols, err = dbi.SelectList("SELECT lower(column_name) FROM information_schema.columns WHERE table_name=UPPER(:1) AND table_schema=database()", cols.Tabname)
	if err != nil || cols.Cols == nil {
		return nil
	}
	cols.Prikey = cols.Cols[0]
	return cols
}

// get columns info for table
func (dbi *DB) TableExists(tabname string) bool {
	tabNums, _ := dbi.Get("SELECT 1 FROM information_schema.tables WHERE table_name=UPPER(:1) AND table_schema=database()", tabname)
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
