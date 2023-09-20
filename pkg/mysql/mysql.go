package mysql

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// New -.
func New(url string) (*sql.DB, error) {
	db, err := sql.Open("mysql", url)
	if err != nil {
		panic(err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return db, nil
}
