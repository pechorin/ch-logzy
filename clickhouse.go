package main

import (
	// "context"
	"database/sql"
	"fmt"
	_ "github.com/mailru/go-clickhouse"
)

func NewClickhouse() *sql.DB {
	// fmt.Println(c)
	conn, err := sql.Open("clickhouse", "http://localhost:8123/default")

	if err != nil {
		panic(fmt.Sprintf("can't connect to clickhouse %v", err))
	}

	if err := conn.Ping(); err != nil {
		panic(fmt.Sprintf("can't ping clickhouse: %v", err))
	}

	return conn
}
