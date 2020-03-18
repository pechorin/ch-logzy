package main

import (
	// "context"
	"database/sql"
	"fmt"
	_ "github.com/mailru/go-clickhouse"
	"log"
	"time"
)

func NewClickhouse() (*sql.DB, error) {
	// fmt.Println(c)
	conn, err := sql.Open("clickhouse", "http://localhost:8123/default")

	if err != nil {
		return nil, fmt.Errorf("can't connect to clickhouse %v", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("can't ping clickhouse: %v", err)
	}

	return conn, nil
}

func CreateTestLogsDatabase(conn *sql.DB) error {
	if _, err := conn.Exec(`
		CREATE TABLE IF NOT EXISTS ch_logzy_logs_test (
			time DateTime
			date Date
			category String
			Level String
			log String
		) engine = Memory
	`); err != nil {
		return fmt.Errorf("can't create test table: %v", err)
	}

	return nil
}

func ReadTestLatest(conn *sql.DB) error {
	rows, err := conn.Query(`
		SELECT * FROM ch_logzy_logs_test LIMIT %v
	`, 10)

	if err != nil {
		return err
	}

	for rows.Next() {
		var (
			logTime  time.Time
			logDate  time.Time
			category string
			level    string
			logText  string
		)

		if err := rows.Scan(&logTime, &logDate, &category, &level, &logText); err != nil {
			return fmt.Errorf("can't scan row: %v", err.Error())
		}

		log.Printf("log: %v %v %v %v %v", logTime, logDate, category, level, logText)
	}

	return nil
}
