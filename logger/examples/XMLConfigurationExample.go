package main

import (
	"github.com/pengswift/libonepiece/logger"
)

func main() {
	logger.LoadConfiguration("example.xml")

	logger.Debug("This will only go to those of you really cool UDP kids!  If you change enabled=true.")
	logger.Info("Oh no!  %d + %d = %d!", 2, 2, 2+2)
	logger.Fatal("About that time, eh chaps?")
}
