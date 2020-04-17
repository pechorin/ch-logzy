package main

import (
	_ "database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/mailru/go-clickhouse"
	"log"
)

// initialize Clickhouse db layer
func NewClickhouse() (*sqlx.DB, error) {
	// fmt.Println(c)
	conn, err := sqlx.Open("clickhouse", "http://localhost:8123/default")

	if err != nil {
		return nil, fmt.Errorf("can't created clickhouse db layer %v", err)
	}

	return conn, nil
}

// Get available clickhouse tables
func FetchClickhouseTables(conn *sqlx.DB) (tables []string, err error) {
	if err := conn.Select(&tables, "SHOW TABLES"); err != nil {
		return tables, err
	}

	return tables, err
}

// Fetch clickhouse results
func FetchClickhouse(conn *sqlx.DB, q ClientQuery) (result []map[string]interface{}, err error) {
	query := fmt.Sprintf("SELECT * FROM %s", q.Table)

	log.Printf("execquery -> %s", query)

	rows, err := conn.Queryx(query)
	if err != nil {
		return result, err
	}

	for rows.Next() {
		row := make(map[string]interface{})
		err := rows.MapScan(row)
		if err != nil {
			return result, err
		}

		result = append(result, row)
	}

	return result, err
}
