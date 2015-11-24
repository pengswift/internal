package main

import (
	"time"
)

import (
	. "github.com/pengswift/libonepiece/logger"
)

func main() {
	log := make(Logger)
	defer log.Close()
	log.AddFilter("stdout", DEBUG, NewConsoleLogWriter())
	log.Info("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
}
