package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"

	. "github.com/pengswift/libonepiece/logger"
)

const (
	filename = "flw.log"
)

func main() {
	log := make(Logger)

	log.AddFilter("file", DEBUG, NewFileLogWriter(filename, false))
	log.Close()

	flw := NewFileLogWriter(filename, false)
	flw.SetFormat("[%D %T] [%L] (%S) %M")
	flw.SetRotate(false)
	flw.SetRotateSize(0)
	flw.SetRotateLines(0)
	flw.SetRotateDaily(false)

	log.AddFilter("file", INFO, flw)
	log.Debug("Everything is created now (notice that I will not be printing to the file)")
	log.Info("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	log.Fatal("Time to close out!")

	log.Close()

	fd, _ := os.Open(filename)
	in := bufio.NewReader(fd)
	fmt.Print("Messages logged to file were: (line numbers not included)\n")
	for lineno := 1; ; lineno++ {
		line, err := in.ReadString('\n')
		if err == io.EOF {
			break
		}
		fmt.Printf("%3d:\t%s", lineno, line)
	}
	fd.Close()

	os.Remove(filename)
}
