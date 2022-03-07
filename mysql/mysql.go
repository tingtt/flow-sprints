package mysql

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

var dsn string

func SetDSNTCP(user string, password string, host string, port int, db string) string {
	dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, port, db)
	return fmt.Sprintf("%s:********@tcp(%s:%d)/%s", user, host, port, db)
}

func Open() (*sql.DB, error) {
	if dsn == "" {
		return nil, errors.New("dsn does not set")
	}
	return sql.Open("mysql", dsn)
}
