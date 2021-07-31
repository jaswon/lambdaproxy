package proxy

import (
	"io"
	"log"
	"net"
)

func Proxy(server, client *net.TCPConn) {
	serverClosed := make(chan struct{}, 1)
	clientClosed := make(chan struct{}, 1)

	go broker(server, client, clientClosed)
	go broker(client, server, serverClosed)

	var waitFor chan struct{}
	select {
	case <-clientClosed:
		server.SetLinger(0)
		server.CloseRead()
		waitFor = serverClosed
	case <-serverClosed:
		client.CloseRead()
		waitFor = clientClosed
	}

	<-waitFor
}

func broker(dst, src net.Conn, srcClosed chan struct{}) {
	if _, err := io.Copy(dst, src); err != nil {
		log.Println("copy error: ", err)
	}
	if err := src.Close(); err != nil {
		log.Println("close error: ", err)
	}
	srcClosed <- struct{}{}
}
