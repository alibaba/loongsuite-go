package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"time"
)

func main() {
	var mysqlDSN string
	mysqlDSN = os.Getenv("MYSQL_DSN")
	if mysqlDSN == "" {
		mysqlDSN = "test:test@tcp(127.0.0.1:3306)/test"
	}
	db, err := sql.Open("mysql", mysqlDSN)
	if err != nil {
		return
	}
	// See "Important settings" section.
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	if _, err := db.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS usersx (id char(255), name VARCHAR(255), age INTEGER)`); err != nil {
		fmt.Printf("exec create error: %v", err)
	}

	if _, err := db.ExecContext(context.Background(), `INSERT INTO usersx (id, name, age) VALUES ( ?, ?, ?)`, "0", "foo", 10); err != nil {
		fmt.Printf("exec insert error: %v", err)
	}

	// test sql inject
	maliciousAnd := "'foo' AND 1 = 1"
	injectedSql := fmt.Sprintf("SELECT * FROM usersx WHERE id = '0' AND name = %s", maliciousAnd)
	if _, err := db.Query(injectedSql); err != nil {
		fmt.Printf("exec query error: %v", err)
	}
}
