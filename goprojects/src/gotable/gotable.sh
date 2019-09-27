#!/usr/bin/ksh

if [ $# -lt 3 ]; then
  echo "Usage: <login> <tabname> <keyname> <,type>"
  echo " type: iog, means insert <tab>_iog sync"
  echo "     : rol, means insert r_<table> sync"
  echo "     : rni, means rol and iog at the same time."
  echo "     : log, means log mode, <table>_yymm redirect"
  echo " default: normal table only"
  exit 1;
fi;

login=$1
tabname=$2
keyname=$3

# 全小写
tab=$(echo $tabname | tr [A-Z] [a-z])
# 首字母大写
Tab=$(echo $tabname | tr [A-Z] [a-z] | sed -nr 's/^(.)/\u\1/p')
key=$(echo $keyname | tr [A-Z] [a-z])
Key=$(echo $keyname | tr [A-Z] [a-z] | sed -nr 's/^(.)/\u\1/p')

TABTYPE=0
if [ $# -gt 3 ]; then
  case $4 in
    "iog"  ) TABTYPE=1; break;;
    "rol"  ) TABTYPE=2; break;;
    "rni"  ) TABTYPE=3; break;;
    "log"  ) TABTYPE=4; break;;
    *      ) TABTYPE=0; break;;
  esac
fi;

# 自动检测是oracle连接还是mysql连接0-mysql,1-oracle
DB=0

echo $login | grep tcp > /dev/null 2>&1
if [ $? -eq 0 ]; then
  login=$(echo "$login" | sed 's/tcp(//;s/)//;s/:/\//;')
fi;

mdbsql -l "$login" "select now()" >/dev/null 2>&1
if [ $? -eq 0 ]; then
  DB=0
else
  orasql -l "$login" "select sysdate from dual" >/dev/null 2>&1
  if [ $? -eq 0 ]; then
    DB=1;
  else
    echo "无法连接数据库"
    exit 1;
  fi;
fi;

if [ $DB -eq 0 ]; then
  CMD=mdbsql
  BCP=mdbbcp
  LINK=GROUP_CONCAT
  DICT=information_schema.columns
  COND="AND table_schema=database()"
  COLID=ordinal_position
else
  CMD=orasql
  BCP=cybcp
  LINK=str_link
  DICT=user_tab_columns
  COND=""
  COLID=column_id
fi;

# 获取字段名列表
column_list=$($CMD -l "$login" "SELECT $LINK(column_name) FROM (SELECT concat(upper(substr(column_name,1,1)),substr(lower(column_name),2)) as column_name FROM $DICT
  WHERE table_name=upper('$tab') $COND ORDER BY $COLID) t")
if [ $? -ne 0 ]; then
  echo "无法获取表的字段列表信息1"
  exit 1;
fi;

select_list=$($CMD -l "$login" "SELECT $LINK(column_name) FROM (SELECT lower(column_name) as column_name FROM $DICT
  WHERE table_name=upper('$tab') $COND ORDER BY $COLID) t")

inputlist=$($CMD -l "$login" "SELECT $LINK(column_name) FROM (SELECT '?' as column_name FROM $DICT
  WHERE table_name=upper('$tab') $COND ORDER BY $COLID) t")
if [ $? -ne 0 ]; then
  echo "无法获取表的字段列表信息2"
  exit 1;
fi;

struct_list=$($BCP -l "$login" "SELECT lower(column_name) as column_name FROM $DICT
  WHERE table_name=upper('$tab') $COND ORDER BY $COLID" 2>/dev/null)
if [ $? -ne 0 ]; then
  echo "无法获取表的字段列表信息3"
  exit 1;
fi;

update_list=$($CMD -l "$login" "SELECT $LINK(column_name) FROM (SELECT concat(upper(substr(column_name,1,1)),substr(concat(lower(column_name),'=?'),2)) as column_name FROM $DICT
  WHERE table_name=upper('$tab') $COND AND column_name<>upper('$key') ORDER BY $COLID) t")
if [ $? -ne 0 ]; then
  echo "无法获取表的字段列表信息4"
  echo "SELECT $LINK(column_name) FROM (SELECT concat(upper(substr(column_name,1,1)),substr(concat(lower(column_name),'=?'),2)) as column_name FROM $DICT
  WHERE table_name=upper('$tab') $COND AND column_name<>upper('$key') ORDER BY $COLID) t"
  exit 1;
fi;

uprowlist=$($CMD -l "$login" "SELECT $LINK(column_name) FROM (SELECT concat('r.',lower(column_name)) as column_name FROM $DICT
  WHERE table_name=upper('$tab') $COND AND column_name<>upper('$key') ORDER BY $COLID) t")
if [ $? -ne 0 ]; then
  echo "无法获取表的字段列表信息5"
  exit 1;
fi;

# init param
paramlist=$(echo ,$column_list | sed 's/,/, \&r\./g' | cut -c3-)
varcolist=$(echo ,$select_list | sed 's/,/, r./g' | cut -c3-)

PARAMLIST=$(echo ,$select_list | sed 's/,/, \&r\./g' | cut -c3-)

echo "package table"
echo '
import (
  "database/sql"
  "reflect"
  "strings"
  "fmt"
  "log"
  "os"'
if [ $TABTYPE -eq 4 ]; then
echo '  "time"
  "strconv"'
fi;
echo ')
'

echo "//table $tab struct
type table$Tab struct {"

for i in ${struct_list}; do
  printf "  %-20s sql.NullString\n" $i
done;

if [ $TABTYPE -eq 4 ]; then
  echo "  YYMM                 string"
fi;

echo '}
'
echo "//table $tab struct
type Table$Tab struct {"

for i in ${struct_list}; do
  Col=$(echo $i | tr [A-Z] [a-z] | sed -nr 's/^(.)/\u\1/p')
  printf "  %-20s string\n" $Col
done;

if [ $TABTYPE -eq 4 ]; then
  echo "YYMM  string"
fi;
echo "  RELA1              string"
echo "  RELA2              string"
echo '}
'
echo "type Map$Tab map[string]Table$Tab

type Class$Tab struct {
  List Map$Tab
  Keys map[int]string
}

var ${Tab}_serialno string
var ${Tab}_db *sql.DB
var ${Tab}_Log = log.New(os.Stdout, \"\", log.Ldate|log.Ltime)
"
echo "//List init before use
func New$Tab(c *sql.DB, arg ...interface{}) *Class$Tab {
  tab := &Class$Tab{List: make(Map$Tab), Keys:make(map[int]string)}
  //tab := new(Class$Tab)
  //tab.List = make(Map$Tab)
  ${Tab}_db = c
  if len(arg) >= 1 {
    kindstr := fmt.Sprintf(\"%s\", reflect.TypeOf(arg[0]))
    if kindstr == \"*log.Logger\" {
      ${Tab}_Log = arg[0].(*log.Logger)
    }
  }
  if len(arg) >= 2 {
    ${Tab}_serialno = arg[1].(string)
    // fmt.Println(reflect.TypeOf(arg[1]), reflect.ValueOf(arg[1]).Kind())
  }
  return tab
}
"

echo "func (tab *Class$Tab) Show(arg ...interface{}) *Class$Tab {"
echo '  for i, value := range tab.List {
    if len(arg) > 0 {
      ks := fmt.Sprintf("%s", reflect.TypeOf(arg[0]))
      if ks == "*log.Logger" {
        tlog := arg[0].(*log.Logger)
        tlog.Println("key=", i,  ", value=", value)
      }
    } else {
      '${Tab}'_Log.Println("key=", i,  ", value=", value)
    }
  }
  return tab
}
'

echo "func (tab *Class$Tab) Clear() *Class$Tab {
  for i, _ := range tab.List {
    delete(tab.List, i)
  }
  for i, _ := range tab.Keys {
    delete(tab.Keys, i)
  }
  return tab
}
"

echo "func (tab *Class$Tab) Count()(int) {
  return len(tab.List)
}

"

echo "func (tab *Class$Tab) Get(key string)(Table$Tab, string) {"
echo '  if v, ok := tab.List[key]; ok {
    return v, ""
  } else {
    '${Tab}'_Log.Println("key =", key, ", List not found.")
    return v, "value not found."
  }
}
'
echo "func (tab *Class$Tab) Set(row Table$Tab) *Class$Tab {
  tab.List[row.$Key] = row
  return tab
}

func (tab *Class$Tab) Serialno(bizid string) *Class$Tab {
  ${Tab}_serialno = bizid
  ${Tab}_Log.Println(\"serialno is set to\", ${Tab}_serialno)
  return tab
}
"
echo "func (tab *Class$Tab) Add(row Table$Tab) *Class$Tab {
  i := len(tab.List)
  tab.List[row.$Key] = row
  tab.Keys[i] = row.$Key
  return tab
}

//extend space for future
func (tab *Class$Tab) Del(key string) *Class$Tab {
  delete(tab.List, key)
  return tab
}
"

if [ $TABTYPE -eq 4 ]; then
echo "// select all columns into List from table
func (tab *Class$Tab) SELECT(key string, yymm string)(string) {
  db := ${Tab}_db
  sql := \"SELECT $column_list FROM $tab\" + \"_\" + yymm +\" WHERE $Key=?\""
echo "  ${Tab}_Log.Println(\"SQL =\", sql)"
echo '  rows, err := db.Query(sql, key)
  if err != nil {
    '${Tab}'_Log.Println("SQL=" + sql + "\n" + err.Error())
    return fmt.Sprintf("%s\n", err)
  }
  defer rows.Close()
'

echo "  var r table$Tab
  i := 0
  for rows.Next() {
    err := rows.Scan(${PARAMLIST})"
echo '    if err != nil {
      '${Tab}'_Log.Println("SQL=" + sql + "\n" + err.Error())
      return fmt.Sprintf("%s\n", err)
    }'

echo "    var R Table$Tab"
if [ $TABTYPE -eq 4 ]; then
echo '    R.YYMM = yymm'
fi;
# 字段转换
for i in ${struct_list}; do
  Col=$(echo $i | tr [A-Z] [a-z] | sed -nr 's/^(.)/\u\1/p')
  echo "    if r.${i}.Valid {
      R.${Col} = r.${i}.String
    }"
done;

echo '
    tab.Add(R)
    i++
  }
  
  if i == 0 {
    '${Tab}'_Log.Println("'$Key' =", key, ", No-data-found.")
  }
  return ""
}
'
 
echo "// select all columns into List from table
func (tab *Class$Tab) Select(key string, arg ...string)(string) {
  var yymm string
  var res string
  if len(arg) == 0 {
    t := time.Now()
    yymm = t.Format(\"0601\")
    res = tab.SELECT(key, yymm)
  } else {
    for i := 0; i < len(arg); i++ {
      yymm = arg[i]
      res += tab.SELECT(key, yymm)
    }
  }
  return res
}
"
  
else
echo "// select all columns into List from table
func (tab *Class$Tab) Select(key string, arg ...string)(string) {
  db := ${Tab}_db"
echo "  sql := \"SELECT $column_list FROM $tab WHERE $Key=?\""
echo '  // specified columns
  if len(arg) == 1 {
    if strings.Index(strings.ToLower(arg[0]), "'$key$'") <= 0 && arg[0] != "*" {
      sql = "SELECT " + "'$key$'," + strings.ToLower(arg[0]) + " FROM '$tab' WHERE '$Key'=?"
    } else {
      sql = "SELECT " + strings.ToLower(arg[0]) + " FROM '$tab' WHERE '$Key'=?"
    }
  } else if len(arg) >= 2 {
    if strings.Index(strings.ToLower(arg[0]), "'$key$'") <= 0 && arg[0] != "*" {
      sql = "SELECT " + "'$key$'," + strings.ToLower(arg[0]) + " FROM '$tab' WHERE " + arg[1]
    } else {
      sql = "SELECT " + strings.ToLower(arg[0]) + " FROM '$tab' WHERE " + arg[1]
    }
  }'
echo "  ${Tab}_Log.Println(\"SQL =\", sql)"

echo "
  plen := 1
  if len(arg) > 3 {
    plen = len(arg) - 2
  }
  
  argnew := make([]interface{}, plen)
  
  if len(arg) < 2 {
    argnew[0] = key
  } else {
    for i := 0; i < len(arg)-2; i++ {
      argnew[i] = arg[i+2]
    }
  }
"
echo '  rows, err := db.Query(sql, argnew...)
  if err != nil {
    '${Tab}'_Log.Println("SQL=" + sql + "\n" + err.Error())
    return fmt.Sprintf("db.Query: %s\n", err)
  }
  defer rows.Close()
'

echo "  var r table$Tab
  columns, err := rows.Columns()
  scanarg := make([]interface{}, len(columns))"

echo "  for i := range columns {"

ii=0;
for i in ${struct_list}; do
  if [ $ii -eq 0 ]; then
    echo "    if strings.ToLower(columns[i]) == \"$i\" {
      scanarg[i] = &r.$i"
  else
  echo "    } else if strings.ToLower(columns[i]) == \"$i\" {
      scanarg[i] = &r.$i"
  fi;
  ii=1
done;
echo "    }
  }
"
echo "  i := 0
  for rows.Next() {
    err := rows.Scan(scanarg...)"
echo '    if err != nil {
      '${Tab}'_Log.Println("SQL=" + sql + "\n" + err.Error())
      return fmt.Sprintf("%s\n", err)
    }'
echo "    var R Table$Tab"
# 字段转换
for i in ${struct_list}; do
  Col=$(echo $i | tr [A-Z] [a-z] | sed -nr 's/^(.)/\u\1/p')
  echo "    if r.${i}.Valid {
      R.${Col} = r.${i}.String
    }"
done;
echo '
    tab.Add(R)
    i++
  }
  
  if i == 0 {
    '${Tab}'_Log.Println("'$Key' =", key, ", No-data-found.")
  }
  return ""
}
'
fi;

if [ $TABTYPE -eq 4 ]; then
echo "func (tab *Class$Tab) SELECTS(yymm string, c string, arg ...interface {})(string) {
  db := ${Tab}_db"
echo '  sql := "SELECT '"$column_list"' FROM '"$tab"'_" + yymm + " WHERE " + c'
else
echo "func (tab *Class$Tab) Selects(c string, arg ...interface {})(string) {
  db := ${Tab}_db"
echo '  sql := "SELECT '"$column_list"' FROM '"$tab"' WHERE " + c'
fi;
echo "  ${Tab}_Log.Println(\"SQL =\", sql)"
echo '  rows, err := db.Query(sql, arg...)
  if err != nil {
    '${Tab}'_Log.Println("SQL=" + sql + "\n" + err.Error())
    return fmt.Sprintf("%s\n", err)
  }
  defer rows.Close()'
echo "
  var r table$Tab"

echo '  for rows.Next() {
    err := rows.Scan('"${PARAMLIST}"')
    if err != nil {
      '${Tab}'_Log.Println("'"$tab"'.rows.Scan" + "\n" + err.Error())
      return fmt.Sprintf("%s\n", err)
    }
    '
echo "    var R Table$Tab"
if [ $TABTYPE -eq 4 ]; then
echo '    R.YYMM = yymm'
fi;
# 字段转换
for i in ${struct_list}; do
  Col=$(echo $i | tr [A-Z] [a-z] | sed -nr 's/^(.)/\u\1/p')
  echo "    if r.${i}.Valid {
      R.${Col} = r.${i}.String
    }"
done;

echo '
    tab.Add(R)
  }
  return ""
}
'

if [ $TABTYPE -eq 4 ]; then
echo "func (tab *Class$Tab) Selects(yymm string, c string, arg ...interface {})(string) {"
echo '  var str string
  if len(yymm) == 4 {
    str = tab.SELECTS(yymm, c, arg...)
  } else if len(yymm) == 5 {
    t := time.Now()
    smon,_ := strconv.Atoi(yymm[0:4])
    emon,_ := strconv.Atoi(t.Format("0601"))
    for i := smon; i <= emon; i++ {
      if (i % 100) == 13 {
        i = i + 88
      }
      str += tab.SELECTS(strconv.Itoa(i), c, arg...)
    }
  } else if len(yymm) == 9 {
    smon,_ := strconv.Atoi(yymm[0:4])
    emon,_ := strconv.Atoi(yymm[5:9])
    for i := smon; i <= emon; i++ {
      if (i % 100) == 13 {
        i = i + 88
      }
      str += tab.SELECTS(strconv.Itoa(i), c, arg...)
    }
  } else {
    '${Tab}'_Log.Println("Unrecognized yymm string = " + yymm)
    return "Unrecognized yymm string = " + yymm
  }
  return str
}
'
fi;

echo "func (tab *Class$Tab) INSERT(row Table$Tab)(string) {
  db := ${Tab}_db"
if [ $TABTYPE -eq 4 ]; then
echo '  var yymm string
  if len(row.YYMM) != 4 {
    t := time.Now()
    yymm = t.Format("0601")
  } else {
    yymm = row.YYMM
  }'
  echo '  sql := "INSERT INTO '$tab'_" + yymm + " ('"$column_list"') VALUES('"$inputList"')"'
else
  echo '  sql := "INSERT INTO '$tab' ('"$column_list"') VALUES('"$inputlist"')"'
fi;
echo "  ${Tab}_Log.Println(\"SQL =\", sql)"
echo '  stmt, err := db.Prepare(sql)
  if err != nil {
    '${Tab}'_Log.Println("SQL=" + sql + "\n" + err.Error())
    return fmt.Sprintf("%s\n", err)
  }
  
  defer stmt.Close()
'
echo "  var r table$Tab"
# 字段转换
for i in ${struct_list}; do
  Col=$(echo $i | tr [A-Z] [a-z] | sed -nr 's/^(.)/\u\1/p')
  echo "  if row.${Col} != \"\" {
    r.${i}.String = row.${Col}
    r.${i}.Valid = true
  }"
done;
echo '
  res, err := stmt.Exec('"$varcolist"')
  if err != nil {
    sql_val := "SQL=" + sql[0:strings.Index(sql, "VALUES")] + "VALUES ("
    v := reflect.ValueOf(row)'
echo "
    sql_val = sql_val + fmt.Sprintf(\"'%v'\", v.Field(0))
    for i := 1; i < v.NumField()-2; i++ {
      sql_val = sql_val + fmt.Sprintf(\",'%v'\", v.Field(i))
    }"
echo '    sql_val = sql_val + ")\n"
    Prd_asset_Log.Println(sql_val + err.Error())
    return (sql_val + err.Error())
  }'
echo '
  rows, err := res.RowsAffected()
  if err != nil {
    '${Tab}'_Log.Println("SQL=" + sql + "\n" + err.Error())
    return fmt.Sprintf("%s\n", err)
  }
  
  if rows > 0 {
    '${Tab}'_Log.Println("'$Tab'.Insert", rows, "rows affected,", "'$Key' =", row.'$Key')
  } else {
    '${Tab}'_Log.Println("'$Tab'.Insert", "Nothing changed,", "'$Key' =", row.'$Key')
  }'
# logtable 增加iog表的写入, insert只会写iog表，不写R表
if [ $TABTYPE -eq 1 -o $TABTYPE -eq 3 ]; then
echo '
  // table for data tracing
  sql = "INSERT INTO '$tab'_iog (Recid,Keyid,Optime,Optype) VALUES(NULL,?,NOW(),?)"
  '${Tab}'_Log.Println("SQL =", sql)
  iogs, err := db.Prepare(sql)
  if err != nil {
    '${Tab}'_Log.Println("SQL=" + sql + "\n" + err.Error())
    return fmt.Sprintf("%s\n", err)
  }
  
  defer iogs.Close()
  res, err = iogs.Exec(row.'$Key', "I")
  if err != nil {
    '${Tab}'_Log.Println("SQL=" + sql + "\n" + err.Error())
    return fmt.Sprintf("%s\n", err)
  }'
fi;
echo '  return ""
}
'
echo "func (tab *Class$Tab) Insert(key string)(string) {
  var str string"
echo '  if key == "" {
    for i, value := range tab.List {
      str = tab.INSERT(value)
      if str != "" {
        return str
      }
      _ = i
    }
  } else {
    if row, ok := tab.List[key]; ok {
      str = tab.INSERT(row)
    } else {
      '${Tab}'_Log.Println("'$Key' =", key, ", value not found.")
      return "key = " + key + ", value not found."
    }
  }
  return str
}
'
if [ $TABTYPE -eq 4 ]; then
echo "// insert into table
func (tab *Class$Tab) DELETE(key string, YYMM string)(string) {
  db := ${Tab}_db"
else
echo "// insert into table
func (tab *Class$Tab) DELETE(key string)(string) {
  db := ${Tab}_db"
fi;

# 改前数据 rolltable 增加iog表的写入
if [ $TABTYPE -eq 2 -o $TABTYPE -eq 3 ]; then
echo '
  // old version data snapshot
  rol := "INSERT INTO r_'$tab' SELECT null,?,NOW(),?,?,t.* FROM '$tab' t WHERE '$key'=?"
  '${Tab}'_Log.Println("SQL =", rol)
  rols, err := db.Prepare(rol)
  if err != nil {
    '${Tab}'_Log.Println("SQL=" + rol + "\n" + err.Error())
    return fmt.Sprintf("%s\n", err)
  }
  
  defer rols.Close()
  _, err = rols.Exec('${Tab}'_serialno, "D", "N", key)
  if err != nil {
    '${Tab}'_Log.Println("SQL=" + rol + "\n" + err.Error())
    return fmt.Sprintf("%s\n", err)
  }
'
fi;

if [ $TABTYPE -eq 4 ]; then
echo '  var yymm string
  if len(YYMM) != 4 {
    t := time.Now()
    yymm = t.Format("0601")
  } else {
    yymm = YYMM
  }'
  echo '  sql := "DELETE FROM '$tab'_" + yymm + " WHERE '$key'=?"'
else
  echo '  sql := "DELETE FROM '$tab' WHERE '$key'=?"'
fi;
echo "  ${Tab}_Log.Println(\"SQL =\", sql)"
echo '  stmt, err := db.Prepare(sql)
  if err != nil {
    '${Tab}'_Log.Println("SQL=" + sql + "\n" + err.Error())
    return fmt.Sprintf("%s\n", err)
  }
  defer stmt.Close()
  res, err := stmt.Exec(key)
  if err != nil {
    '${Tab}'_Log.Println("SQL=" + sql + "\n" + err.Error())
    return fmt.Sprintf("%s\n", err)
  }
  
  rows, err := res.RowsAffected()
  if err != nil {
    '${Tab}'_Log.Println("SQL=" + sql + "\n" + err.Error())
    return fmt.Sprintf("%s\n", err)
  }
  
  if rows > 0 {
    '${Tab}'_Log.Println("'$Tab'.Delete", rows, "rows affected,", "'$Key' =", key)
  } else {
    '${Tab}'_Log.Println("'$Tab'.Delete", "Nothing changed,", "'$Key' =", key)
  }'
# logtable 增加iog表的写入
if [ $TABTYPE -eq 1 -o $TABTYPE -eq 3 ]; then
echo '
  // table for data tracing
  sql = "INSERT INTO '$tab'_iog (Recid,Keyid,Optime,Optype) VALUES(NULL,?,NOW(),?)"
  '${Tab}'_Log.Println("SQL =", sql)
  iogs, err := db.Prepare(sql)
  if err != nil {
    '${Tab}'_Log.Println("SQL=" + sql + "\n" + err.Error())
    return fmt.Sprintf("%s\n", err)
  }
  
  defer iogs.Close()
  res, err = iogs.Exec(key, "D")
  if err != nil {
    '${Tab}'_Log.Println("SQL=" + sql + "\n" + err.Error())
    return fmt.Sprintf("%s\n", err)
  }'
fi;
echo '  return ""
}
'

echo "// insert into table
func (tab *Class$Tab) Delete(key string)(string) {
  var str string"
echo '  if key == "" {
    for i, value := range tab.List {'
if [ $TABTYPE -eq 4 ]; then
  echo '      str = tab.DELETE(value.'$Key', value.YYMM)'
else
  echo '      str = tab.DELETE(value.'$Key')'
fi;
echo '      if str != "" {
        return str
      }
      _ = i
    }
  } else {'

if [ $TABTYPE -eq 4 ]; then
  echo '    r,_ := tab.List[key];'
  echo '    str = tab.DELETE(key, r.YYMM)'
else
  echo '    str = tab.DELETE(key)'
fi;
echo '  }
  return str
}
'

echo "// update key, null stands for all columns
func (tab *Class$Tab) UPDATE(row Table$Tab)(string) {
  db := ${Tab}_db"
# 改前数据 rolltable 增加iog表的写入
if [ $TABTYPE -eq 2 -o $TABTYPE -eq 3 ]; then
echo '
  // old version data snapshot
  rol := "INSERT INTO r_'$tab' SELECT null,?,NOW(),?,?,t.* FROM '$tab' t WHERE '$key'=?"
  '${Tab}'_Log.Println("SQL =", rol)
  rols, err := db.Prepare(rol)
  if err != nil {
    '${Tab}'_Log.Println("SQL=" + rol + "\n" + err.Error())
    return fmt.Sprintf("%s\n", err)
  }
  
  defer rols.Close()
  _, err = rols.Exec('${Tab}'_serialno, "U", "N", row.'$Key')
  if err != nil {
    '${Tab}'_Log.Println("SQL=" + rol + "\n" + err.Error())
    return fmt.Sprintf("%s\n", err)
  }
'
fi;

if [ $TABTYPE -eq 4 ]; then
echo '  var yymm string
  if len(row.YYMM) != 4 {
    t := time.Now()
    yymm = t.Format("0601")
  } else {
    yymm = row.YYMM
  }'
  echo '  sql := "UPDATE '$tab'_" + yymm + " SET '"$update_list"' WHERE '"$Key"'=?"'
else
  echo '  sql := "UPDATE '$tab' SET '"$update_list"' WHERE '"$Key"'=?"'
fi;
echo "  ${Tab}_Log.Println(\"SQL =\", sql)"
echo '  stmt, err := db.Prepare(sql)
  if err != nil {
    '${Tab}'_Log.Println("SQL=" + sql + "\n" + err.Error())
    return fmt.Sprintf("%s\n", err)
  }
  
  defer stmt.Close()
'
echo "  var r table$Tab"
# 字段转换
for i in ${struct_list}; do
  Col=$(echo $i | tr [A-Z] [a-z] | sed -nr 's/^(.)/\u\1/p')
  echo "  if row.${Col} != \"\" {
    r.${i}.String = row.${Col}
    r.${i}.Valid = true
  }"
done;
echo '
  res, err := stmt.Exec('"$uprowlist"', r.'$key')
  if err != nil {
    '${Tab}'_Log.Println("SQL=" + sql + "\n" + err.Error())
    return fmt.Sprintf("%s\n", err)
  }
  
  rows, err := res.RowsAffected()
  if err != nil {
    '${Tab}'_Log.Println("SQL=" + sql + "\n" + err.Error())
    return fmt.Sprintf("%s\n", err)
  }
  
  if rows > 0 {
    '${Tab}'_Log.Println("'$Tab'.Update:", rows, "rows affected,", "'$Key' =", row.'$Key')
  } else {
    '${Tab}'_Log.Println("'$Tab'.Update:", "Nothing changed,", "'$Key' =", row.'$Key')
  }'
# logtable 增加iog表的写入
if [ $TABTYPE -eq 1 -o $TABTYPE -eq 3 ]; then
echo '
  // table for data tracing
  sql = "INSERT INTO '$tab'_iog (Recid,Keyid,Optime,Optype) VALUES(NULL,?,NOW(),?)"
  '${Tab}'_Log.Println("SQL =", sql)
  iogs, err := db.Prepare(sql)
  if err != nil {
    '${Tab}'_Log.Println("SQL=" + sql + "\n" + err.Error())
    return fmt.Sprintf("%s\n", err)
  }
  
  defer iogs.Close()
  res, err = iogs.Exec(row.'$Key', "U")
  if err != nil {
    '${Tab}'_Log.Println("SQL=" + sql + "\n" + err.Error())
    return fmt.Sprintf("%s\n", err)
  }'
fi;
echo '  return ""
}
'

echo "
// update key, null stands for all columns
func (tab *Class$Tab) Update(key string)(string) {
  var str string"
echo '  if key == "" {
    for i, value := range tab.List {
      str = tab.UPDATE(value)
      if str != "" {
        return str
      }
      _ = i
    }
  } else {
    if row, ok := tab.List[key]; ok {
      str = tab.UPDATE(row)
    } else {
      '${Tab}'_Log.Println("'$Key' =", key, ", value not found.")
      return "key = " + key + ", value not found."
    }
  }
  return str
}
'

echo "func (tab *Class$Tab) Updates(key string, c string, arg ...interface {})(string) {
  db := ${Tab}_db"

# 改前数据 rolltable 增加iog表的写入
if [ $TABTYPE -eq 2 -o $TABTYPE -eq 3 ]; then
echo '
  // old version data snapshot
  rol := "INSERT INTO r_'$tab' SELECT null,?,NOW(),?,?,t.* FROM '$tab' t WHERE '$key'=?"
  '${Tab}'_Log.Println("SQL =", rol)
  rols, err := db.Prepare(rol)
  if err != nil {
    '${Tab}'_Log.Println("SQL=" + rol + "\n" + err.Error())
    return fmt.Sprintf("%s\n", err)
  }
  
  defer rols.Close()
  _, err = rols.Exec('${Tab}'_serialno, "U", "N", key)
  if err != nil {
    '${Tab}'_Log.Println("SQL=" + rol + "\n" + err.Error())
    return fmt.Sprintf("%s\n", err)
  }
'
fi;
echo "  argnew := make([]interface{}, 1+len(arg))
  i := 0;
  for i = 0; i < len(arg); i++ {
    argnew[i] = arg[i]
  }
  argnew[i] = key
"

if [ $TABTYPE -eq 4 ]; then
echo '  row := tab.List[key]'
echo '  var yymm string
  if len(row.YYMM) != 4 {
    t := time.Now()
    yymm = t.Format("0601")
  } else {
    yymm = row.YYMM
  }'
  echo '  sql := "UPDATE '$tab'_" + yymm + " SET " + c + " WHERE '$key'=?"'
else
  echo '  sql := "UPDATE '$tab' SET " + c + " WHERE '$key'=?"'
fi;
echo "  ${Tab}_Log.Println(\"SQL =\", sql)"
echo '  stmt, err := db.Prepare(sql)
  if err != nil {
    '${Tab}'_Log.Println("SQL=" + sql + "\n" + err.Error())
    return fmt.Sprintf("%s\n", err)
  }
  
  defer stmt.Close()
  res, err := stmt.Exec(argnew...)
  if err != nil {
    '${Tab}'_Log.Println("SQL=" + sql + "\n" + err.Error())
    return fmt.Sprintf("%s\n", err)
  }
  
  rows, err := res.RowsAffected()
  if err != nil {
    '${Tab}'_Log.Println("SQL=" + sql + "\n" + err.Error())
    return fmt.Sprintf("%s\n", err)
  }
  
  if rows > 0 {
    '${Tab}'_Log.Println("'$Tab'.Updates:", rows, "rows affected,", "'$Key' =", key)
  } else {
    '${Tab}'_Log.Println("'$Tab'.Updates:", "Nothing changed,", "'$Key' =", key)
  }'
# logtable 增加iog表的写入
if [ $TABTYPE -eq 1 -o $TABTYPE -eq 3 ]; then
echo '
  // taiogble for data tracing
  sql = "INSERT INTO '$tab'_iog (Recid,Keyid,Optime,Optype) VALUES(NULL,?,NOW(),?)"
  '${Tab}'_Log.Println("SQL =", sql)
  iogs, err := db.Prepare(sql)
  if err != nil {
    '${Tab}'_Log.Println("SQL=" + sql + "\n" + err.Error())
    return fmt.Sprintf("%s\n", err)
  }
  
  defer iogs.Close()
  res, err = iogs.Exec(key, "U")
  if err != nil {
    '${Tab}'_Log.Println("SQL=" + sql + "\n" + err.Error())
    return fmt.Sprintf("%s\n", err)
  }'
fi;
echo '  return ""
}
'
