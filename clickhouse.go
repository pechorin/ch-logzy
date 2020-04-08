package main

import (
	_ "database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/mailru/go-clickhouse"
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

func AvailableTables(conn *sqlx.DB) (tables []string, err error) {
	if err := conn.Select(&tables, "SHOW TABLES"); err != nil {
		return tables, err
	}

	return tables, err
}

// func RunQuery(conn *sql.DB, query Query) (interface{}, error) {}
