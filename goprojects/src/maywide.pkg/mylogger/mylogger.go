// mylogger
package mylogger

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/axgle/mahonia"
)

type MyLogger struct {
	logger   *log.Logger
	head     string
	file     *os.File
	filename string
	postfix  string
}

var logger *MyLogger

// 获取跟踪
func GetLogger() *MyLogger {
	return logger
}

//打开日志文件
func NewLogger(fname string) (myloger *MyLogger, err error) {
	filename := fname
	logfile := os.Stdout

	if filename == "/dev/null" || filename == "" {
		nilout, _ := os.OpenFile(os.DevNull, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		logfile = nilout
	}

	postfix := " "
	if filename != "stdout" && filename != "stderr" && filename != "/dev/null" && filename != "" {
		_, _, day := time.Now().Date()
		postfix := strconv.Itoa(day)
		if day < 10 {
			postfix = "0" + postfix
		}

		logfile, err = os.OpenFile(filename+"."+postfix, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("%s\r\n", err.Error())
			os.Exit(-1)
		}
		os.Symlink(filename+"."+postfix, filename)
	}
	filepoint := log.New(logfile, "", log.Ldate|log.Ltime)
	myloger = &MyLogger{filepoint, " ", logfile, filename, postfix}
	logger = myloger
	return
}

//克隆一个日志，改变打印头
func (ml *MyLogger) CloneLogger(head string) (myloger *MyLogger) {
	ml.checkDayChg()
	myloger = &MyLogger{ml.logger, " ", ml.file, ml.filename, ml.postfix}
	myloger.SetHead(head)
	logger = myloger
	return
}

func (ml *MyLogger) GetLogger() *log.Logger {
	return ml.logger
}

//检查日期变更
func (ml *MyLogger) checkDayChg() {
	if ml.filename == "stdout" || ml.filename == "stderr" || ml.filename == "/dev/null" {
		return
	}
	_, _, day := time.Now().Date()
	postfix := strconv.Itoa(day)
	if day < 10 {
		postfix = "0" + postfix
	}
	if postfix == ml.postfix {
		return
	}
	ml.file.Close()
	logfile, err := os.OpenFile(ml.filename+"."+postfix, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("%s\r\n", err.Error())
		os.Exit(-1)
	}
	os.Remove(ml.filename)
	os.Symlink(ml.filename+"."+postfix, ml.filename)

	logger := log.New(logfile, "", log.Ldate|log.Ltime)
	ml.file = logfile
	ml.logger = logger

}

//设置打印头
func (ml *MyLogger) SetHead(head string) {
	if head == "" {
		ml.head = ""
		return
	} else if head == " " {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		ml.head = "[" + strconv.Itoa(r.Intn(10000)) + "]"
	} else {
		ml.head = "[" + head + "]"
	}
}

//格式化打印
func (ml *MyLogger) Printf(format string, v ...interface{}) {
	ml.checkDayChg()
	if ml.head == "" || ml.head == " " {
		ml.logger.Printf(format+"\n", v...)
	} else {
		ml.logger.Printf(ml.head+" "+format+"\n", v...)
	}
}

//logging only
func (ml *MyLogger) Println(v ...interface{}) {
	ml.checkDayChg()
	str := strings.TrimRight(fmt.Sprintln(v...), "\n")
	if ml.head != "" && ml.head != " " {
		ml.logger.Println(ml.head, str)
	} else {
		ml.logger.Println(str)
	}
}

//convert and logging
func (ml *MyLogger) Log(v ...interface{}) {
	ml.checkDayChg()
	str := fmt.Sprint(v...)
	if runtime.GOOS != "windows" {
		str = mahonia.NewEncoder("gbk").ConvertString(str)
	}

	if ml.head != "" && ml.head != " " {
		ml.logger.Println(ml.head, str)
	} else {
		ml.logger.Println(str)
	}
}

//convert and logging and exit
func (ml *MyLogger) Panic(v ...interface{}) {
	ml.checkDayChg()
	str := fmt.Sprint(v...)
	if runtime.GOOS != "windows" {
		str = mahonia.NewEncoder("gbk").ConvertString(str)
	}

	if ml.head != "" && ml.head != " " {
		ml.logger.Println(ml.head, str)
	} else {
		ml.logger.Println(str)
	}
	os.Exit(1)
}

//convert and logging
func (ml *MyLogger) Raw(v ...interface{}) {
	ml.checkDayChg()
	str := fmt.Sprint(v...)
	if runtime.GOOS != "windows" {
		str = mahonia.NewEncoder("gbk").ConvertString(str)
	}

	l := ml.logger.Flags()
	ml.logger.SetFlags(0)
	ml.logger.Printf(str)
	ml.logger.SetFlags(l)
	return
}

//func main() {
//	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
//	myloger := New(logger, "head")
//	myloger.Println("sdfdsf")
//}

func SetHead(head string) {
	if logger == nil {
		return
	}
	ml := logger
	if head == "" {
		ml.head = ""
		return
	} else if head == " " {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		ml.head = "[" + strconv.Itoa(r.Intn(10000)) + "]"
	} else {
		ml.head = head
	}
}

//格式化打印
func Printf(format string, v ...interface{}) {
	logger.Printf(format, v...)
}

//logging only
func Println(v ...interface{}) {
	logger.Println(v...)
}

//convert and logging
func Log(v ...interface{}) {
	logger.Log(v...)
}

//convert and logging and exit
func Panic(v ...interface{}) {
	logger.Panic(v...)
}

//convert and logging
func Raw(v ...interface{}) {
	logger.Raw(v...)
	return
}
