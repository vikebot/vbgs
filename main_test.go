package main

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	config = loadConfig("config/debug.json")
	gsInit()
	battleInit()
	os.Exit(m.Run())
}
