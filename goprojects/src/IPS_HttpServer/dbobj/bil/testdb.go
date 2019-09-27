// swpLog
package bil

import (
	"database/sql"
)

const dbUser = "BOSS_BIL"
const tableName = "testdb"
const fields = "a,b"

type Testdb struct {
	A string
	B string
}

func (dbtable *Testdb) Insert(tx *sql.Tx, tb *Testdb) (err error) {
	_, err = tx.Exec("INSERT INTO "+tableName+" ("+fields+") "+
		" VALUES (:1,:2)",
		tb.A,
		tb.B)
	return
}
