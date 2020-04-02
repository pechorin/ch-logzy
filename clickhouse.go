package main

import (
	// "context"
	"database/sql"
	"fmt"
	_ "github.com/mailru/go-clickhouse"
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

// func GetAvailableDatabases(conn *sql.DB) ([]string, error) {}
// func RunQuery(conn *sql.DB, query Query) (interface{}, error) {}
