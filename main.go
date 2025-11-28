package main

import (
	"nostar/cmd"
	"nostar/internal/logger"
)

func main() {
	glogger := logger.Load()
	defer glogger.Sync() // 必要
	cmd.Execute()
}
