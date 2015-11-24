package main

import (
	"time"
)

import (
	"github.com/pengswift/libonepiece/logger"
)

func main() {
	log := make(logger.Logger)

	log.AddFilter("network", logger.DEBUG, logger.NewSocketLogWriter("udp", "127.0.0.1:8089"))

	log.Info("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))

	time.Sleep(500000)
	log.Close()
}
