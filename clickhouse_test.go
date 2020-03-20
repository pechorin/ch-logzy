package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mailru/go-clickhouse"
	"log"
	"testing"
	"time"
)

var tableName = "ch_logzy_logs_test"

func TestClickhouseBasicConnection(t *testing.T) {
	conn, err := NewClickhouse()

	if err != nil {
		t.Error(err)
	}

	if err := createDB(conn); err != nil {
		t.Error(err)
	}

	if err := populate(conn); err != nil {
		t.Error(err)
	}

	if err := read(conn); err != nil {
		t.Error(err)
	}
}

func createDB(conn *sql.DB) error {
	if _, err := conn.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %v", tableName)); err != nil {
		return fmt.Errorf("can't drop test table: %v", err)
	}

	if _, err := conn.Exec(fmt.Sprintf(`
		CREATE TABLE %v (
			time DateTime DEFAULT now(),
			date Date DEFAULT toDate(time),
			category String,
			level String,
			log String
		) ENGINE = Memory;
	`, tableName)); err != nil {
		return fmt.Errorf("can't create test table: %v", err)
	}

	return nil
}

func populate(conn *sql.DB) error {
	tx, err := conn.Begin()

	if err != nil {
		return err
	}

	stm, err := tx.Prepare(fmt.Sprintf(`
		INSERT INTO %v (
			category,
			level,
			log
		) VALUES (
			?, ?, ?
	)`, tableName))

	if err != nil {
		return err
	}

	for i := 0; i < 10; i++ {
		_, err := stm.Exec("default", "info", fmt.Sprintf("test message %v", i))

		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func read(conn *sql.DB) error {
	limit := 10
	iterated := 0

	query := fmt.Sprintf("SELECT * FROM ch_logzy_logs_test LIMIT %v", limit)
	rows, err := conn.Query(query)

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

		if err := rows.Err(); err != nil {
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
