package memcached

import (
	"database/sql"
	"strings"

	"errors"
	"fmt"
	"reflect"
)

func SqlSelect(dbs interface{}, sqls string, arg ...interface{}) ([]map[string]string, error) {
	var rows *sql.Rows
	var err error

	if reflect.TypeOf(dbs).Elem().Name() == "DB" {
		rows, err = dbs.(*sql.DB).Query(sqls, arg...)
	} else if reflect.TypeOf(dbs).Elem().Name() == "Tx" {
		rows, err = dbs.(*sql.Tx).Query(sqls, arg...)
	} else {
		fmt.Println("db.type.name =", reflect.TypeOf(dbs).Elem().Name())
		return nil, errors.New("Un-recognized db object")
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	var cols []sql.NullString
	for i := 0; i < len(columns); i++ {
		cols = append(cols, sql.NullString{"", false})
	}

	colvar := make([]interface{}, len(columns))
	for i := 0; i < len(columns); i++ {
		colvar[i] = &cols[i]
	}

	var result []map[string]string

	for rows.Next() {
		line := make(map[string]string)
		err := rows.Scan(colvar...)
		if err != nil {
			return nil, err
		}
		// output all columns value
		for i := 0; i < len(columns); i++ {
			if cols[i].Valid {
				line[strings.ToLower(columns[i])] = cols[i].String
			} else {
				line[strings.ToLower(columns[i])] = ""
			}
		}
		result = append(result, line)
	}
	return result, nil
}
