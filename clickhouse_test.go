package main

import (
	"fmt"
	"log"
	"testing"
)

func TestClickhouse(t *testing.T) {
	conn, err := NewClickhouse()

	if err != nil {
		log.Fatalf(err.Error())
	}

	if err := CreateTestLogsDatabase(conn); err != nil {
		log.Fatalf(err.Error())
	}

	fmt.Println("TestClickhouse() passed")
}
