package pool

import (
	"os"
	"sync"
	"time"

	"maywide.pkg/database"
	mylog "maywide.pkg/mylogger"
)

var Addr string = "10.205.28.86:9527"
var Driver string = "oracle"
var MinPoolSize int = 2
var MaxPoolSize int = 5
var Show bool = false
var DSN string
var IdleSecond int = 120

type DBPool struct {
	Pool      chan *database.DB
	Connected int
	Used      int
	Freed     int
	Mutex     sync.Mutex
}

// global calling
var dbi *DBPool

// init by import
func init() {
	dbi = new(DBPool)
	dbdebug := os.Getenv("DB_DEBUG")
	if dbdebug != "Y" && dbdebug != "y" {
		Show = false
	} else {
		Show = true
	}
	// init connection pool
	dbi.Pool = make(chan *database.DB, MaxPoolSize)
	dbi.Connected = 0
	dbi.Used = 0
	dbi.Freed = 0
	return
}

func initPool() {
	dbi.Mutex.Lock()
	defer dbi.Mutex.Unlock()

	if dbi == nil || dbi.Connected != 0 {
		return
	}

	// connect minPoolSize
	for i := 0; i < MinPoolSize; i++ {
		database.Addr = Addr
		database.Show = Show
		database.Driver = Driver
		db := database.Connect(DSN, -1)
		dbi.Pool <- db
		dbi.Connected++
		dbi.Freed++
	}
	dbi.Used = 0
	mylog.Log("init.connected=", dbi.Connected, ", used=", dbi.Used, ", free=", dbi.Freed)
}

func GetDB() *database.DB {
	// init and get db
	initPool()

	// mutex to get database
	dbi.Mutex.Lock()
	// connect when no active connection
	if dbi.Freed <= 0 && dbi.Connected < MaxPoolSize {
		// only connect one time. error will be raise
		db := database.Connect(DSN, 0)
		//		if db.Sqlcode != 0 {
		//			mylog.Log("ERR", db.Sqlcode, db.Sqlerrm)
		//			db.Close()
		//			dbi.Mutex.Unlock()
		//			return nil
		//		}
		mylog.Println("new database connect, session id =", db.Sid)
		dbi.Connected++
		dbi.Used++
		dbi.Mutex.Unlock()
		return db
	}
	dbi.Mutex.Unlock()

	mylog.Println("wating for db.pool ...")
	// wait until active connecttion exists
	db, ok := <-dbi.Pool
	if ok == false {
		mylog.Println("Cannot get db from DBPool.")
		os.Exit(0)
	}

	dbi.Mutex.Lock()
	dbi.Used++
	dbi.Freed--
	dbi.Mutex.Unlock()

	err := db.Ping()
	if err != nil {
		mylog.Println("database.id=", db.Sid, " was terminated, auto reconnect!!!")
		db.Disconnect()
		db.Connect()
	}
	return db
}

func PutDB(db *database.DB) {
	dbi.Mutex.Lock()
	dbi.Pool <- db
	dbi.Used--
	dbi.Freed++
	dbi.Mutex.Unlock()
}

// clear idle connections between 5min
func AutoClear(seconds int) {
	for {
		if dbi == nil || dbi.Connected == 0 {
			time.Sleep(time.Second * time.Duration(seconds))
		}

		dbi.Mutex.Lock()
		if dbi.Freed > 0 && dbi.Connected > MinPoolSize {
			freeFlow := dbi.Connected - MinPoolSize
			if freeFlow > dbi.Freed {
				freeFlow = dbi.Freed
			}
			// release db by freeFlow
			for i := 0; i < freeFlow; i++ {
				db, ok := <-dbi.Pool
				if ok == false {
					mylog.Println("Cannot get db from DBPool.")
					os.Exit(0)
				}
				mylog.Log("database.sid=", db.Sid, " was release by idle.")
				db.Disconnect()
				dbi.Connected--
				dbi.Freed--
			}
		}
		dbi.Mutex.Unlock()
		time.Sleep(time.Second * time.Duration(seconds))
	}
}
