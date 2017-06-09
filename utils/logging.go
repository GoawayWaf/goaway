package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"errors"
	"time"
	"strings"
)

var logger *log.Logger
var dbOut *dbWriter

type dbWriter struct {
	conn *sql.DB
}

func NewDbWriter(conn *sql.DB) *dbWriter {
	return &dbWriter{conn: conn}
}

func (d *dbWriter) Write(p []byte) (n int, err error) {
	query := "INSERT INTO log (host, date, msg) VALUES (?, ?, ?)"
	host, _ := os.Hostname()
	_, err = d.conn.Exec(query, host, time.Now().Format(MYSQLDATETIME), strings.Trim(string(p)[20:], "\n"))

	n = len(p)
	if err != nil {
		n = 0
	}

	return
}

func init() {
	fmt.Println("initializing utils.logger")
	out, err := os.OpenFile("/var/log/goaway/error.log", os.O_CREATE | os.O_APPEND | os.O_RDWR, 0755)
	if err != nil {
		fmt.Println("logging errors to", os.Stderr.Name())
		out = os.Stderr
	}

	/*dbConn, err := sql.Open(
		"mysql",
		fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?charset=utf8",
			"root",
			"password",
			"104.154.164.254",
			"3306",
			"goaway"))
	if err != nil {
		panic(err)
	}

	dbOut = NewDbWriter(dbConn)*/

	multi := io.MultiWriter(out, os.Stdout)
	//logger obj default set
	logger = log.New(multi, "", log.LstdFlags)
}

func NewLogger(logFileName string) *log.Logger {

	if len(logFileName) <= 0 {
		CheckAndPanic(errors.New("you must provide a log file name"))
	}

	out, err := os.OpenFile(logFileName, os.O_CREATE | os.O_APPEND | os.O_RDWR, 0755)
	if err != nil {
		fmt.Println(err.Error())
	}
	multi := io.MultiWriter(out, os.Stdout)
	return log.New(multi, "", log.LstdFlags)
}

func CheckAndLogError(err error, extra ...interface{}) {
	if err != nil {
		_, fileName, fileLine, ok := runtime.Caller(1)
		line := getLogLine(err.Error(), fileName, fileLine, ok)
		logger.Println(line, extra)
	}
}

func LogErrorMessage(message string) {
	_, fileName, fileLine, ok := runtime.Caller(1)
	line := getLogLine(message, fileName, fileLine, ok)
	logger.Println(line)
}

func CheckFatal(err error) {
	if err != nil {
		_, fileName, fileLine, ok := runtime.Caller(1)
		line := getLogLine(err.Error(), fileName, fileLine, ok)
		logger.Fatal(line)
	}
}

func LogAndFatal(message string) {
	_, fileName, fileLine, ok := runtime.Caller(1)
	line := getLogLine(message, fileName, fileLine, ok)
	logger.Fatal(line)
}

func CheckAndPanic(err error) {
	if err != nil {
		_, fileName, fileLine, ok := runtime.Caller(1)
		line := getLogLine(err.Error(), fileName, fileLine, ok)
		panic(line)
	}
}
func PanicMessage(message string) {
	_, fileName, fileLine, ok := runtime.Caller(1)
	line := getLogLine(message, fileName, fileLine, ok)
	panic(line)
}

func getLogLine(message string, fileName string, fileLine int, ok bool) string {

	if ok {
		return fmt.Sprintf("%s:%d %s", fileName, fileLine, message)
	}
	return message
}
