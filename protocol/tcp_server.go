package protocol

import (
	"net"
	"runtime"
	"strings"

	"github.com/pengswift/libonepiece/app"
)

type TCPHandler interface {
	Handle(net.Conn)
}

func TCPServer(listener net.Listener, handler TCPHandler, l app.Logger) {
	l.Info("TCP: listener on %s", listener.Addr())

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
				l.Info("NOTICE: temporary Accept() failure - %s", err)
				runtime.Gosched()
				continue
			}

			if !strings.Contains(err.Error(), "use of closed network connection") {
				l.Info("ERROR: listener.Accept() - %s", err)
			}
			break
		}
		go handler.Handle(clientConn)
	}
	l.Info("TCP: closing %s", listener.Addr())
}
