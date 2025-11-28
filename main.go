package main

import (
	"nostar/cmd"
	"nostar/internal/logger"
)

func main() {
	defer logger.LogPanic()
	cmd.Execute()
}
