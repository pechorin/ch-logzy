package main

import (
	"testing"
)

func TestNewPushServer(t *testing.T) {
	server := NewPushServer()
	err := server.Run()

	if err != nil {
		t.Error(err)
	}
}
