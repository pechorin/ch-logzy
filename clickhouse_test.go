package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mailru/go-clickhouse"
	"log"
	"testing"
	"time"
)

func TestClickhouse(t *testing.T) {
	conn, err := NewClickhouse()

	if err != nil {
		t.Error(err)
	}

	if err := createTestLogsDatabase(conn); err != nil {
		t.Error(err)
	}

	if err := readTestLatest(conn); err != nil {
		t.Error(err)
	}
}

func createTestLogsDatabase(conn *sql.DB) error {
	if _, err := conn.Exec(`
		CREATE TABLE IF NOT EXISTS ch_logzy_logs_test (
			time DateTime DEFAULT now(),
			date Date DEFAULT toDate(time),
			category String,
			Level String,
			log String
		) ENGINE = Memory;
	`); err != nil {
		return fmt.Errorf("can't create test table: %v", err)
	}

	return nil
}

func readTestLatest(conn *sql.DB) error {
	limit := 10
	iterated := 0

	rows, err := conn.Query(fmt.Sprintf("SELECT * FROM ch_logzy_logs_test LIMIT %v", limit))

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
		iterated = iterated + 1
	}

	if iterated != limit {
		return fmt.Errorf("not all records writed or readed: (writed/readed) %v/%v", limit, iterated)
	}

	return nil
}
