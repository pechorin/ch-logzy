package main

import (
	// "context"
	"database/sql"
	"fmt"
	_ "github.com/mailru/go-clickhouse"
	"log"
	"time"
)

func NewClickhouse() *sql.DB {
	// fmt.Println(c)
	conn, err := sql.Open("clickhouse", "http://localhost:8123/default")

	if err != nil {
		log.Fatal(fmt.Sprintf("can't connect to clickhouse %v", err))
	}

	if err := conn.Ping(); err != nil {
		log.Fatal(fmt.Sprintf("can't ping clickhouse: %v", err))
	}

	return conn
}

func CreateTestLogsDatabase(conn *sql.DB) {
	_, err := conn.Exec(`
		CREATE TABLE IF NOT EXISTS ch_logzy_logs_test (
			time DateTime
			date Date
			category String
			Level String
			log String
		) engine = Memory
	`)

	if err != nil {
		log.Fatal(fmt.Sprintf("can't create test table: %v", err))
	}
}

func ReadTestLatest(conn *sql.DB) {
	rows, err := conn.Query(`
		SELECT * FROM ch_logzy_logs_test LIMIT %v
	`, 10)

	if err != nil {
		log.Fatal(err)
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
			log.Fatal(fmt.Sprintf("can't scan row: %v", err))
		}

		log.Printf("log: %v %v %v %v %v", logTime, logDate, category, level, logText)
	}
}

func InsertTestLog(conn *sql.DB, insertMap map[string]interface{}) {
	tx, err := conn.Begin()

	if err != nil {
		log.Fatal(err)
	}

	stm, err := tx.Prepare(`
		INSERT INTO ch_logzy_logs_test (category, level, log) VALUES (?, ?, ?)
	`)

	if err != nil {
		log.Fatal(err)
	}

	if _, err := stm.Exec(
		insertMap["category"],
		insertMap["level"],
		insertMap["log"],
	); err != nil {
		log.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

func RunTestInsertionCycle(conn *sql.DB) {
	for i := 0; i < 100; i++ {
		el := make(map[string]interface{})
		el["category"] = "default"
		el["level"] = "info"
		el["log"] = "test log message " + string(i)

		InsertTestLog(conn, el)
	}
}
