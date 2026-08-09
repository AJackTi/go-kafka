package mysql

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// New -.
func New(url string) (*sql.DB, error) {
	db, err := sql.Open("mysql", normalizeDSN(url))
	if err != nil {
		return nil, fmt.Errorf("mysql open: %w", err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return db, nil
}

func normalizeDSN(value string) string {
	return strings.TrimPrefix(strings.TrimSpace(value), "mysql://")
}
