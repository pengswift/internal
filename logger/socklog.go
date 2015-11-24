package logger

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
)

type SocketLogWriter chan *LogRecord

func (w SocketLogWriter) LogWrite(rec *LogRecord) {
	w <- rec
}

func (w SocketLogWriter) Close() {
	close(w)
}

func NewSocketLogWriter(proto, hostport string) SocketLogWriter {
	sock, err := net.Dial(proto, hostport)
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewSocketLogWriter(%q): %s\n", hostport, err)
		return nil
	}

	w := SocketLogWriter(make(chan *LogRecord, LogBufferLength))

	go func() {
		defer func() {
			if sock != nil && proto == "tcp" {
				sock.Close()
			}
		}()

		for rec := range w {
			js, err := json.Marshal(rec)
			if err != nil {
				fmt.Fprint(os.Stderr, "SocketLogWriter(%q): %s", hostport, err)
				return
			}

			_, err = sock.Write(js)
			if err != nil {
				fmt.Fprintf(os.Stderr, "SocketLogWriter(%q): %s", hostport, err)
				return
			}
		}
	}()

	return w
}
