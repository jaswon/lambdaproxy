package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/hashicorp/yamux"

	"cmd/host/tunneler"
)

var (
	forwardingPort = flag.String("p", "6789", "port used to listen for proxy requests")
)

func init() { flag.Parse() }

func serve(forwarder net.Listener, tunnel *yamux.Session) {
	for {
		c, err := forwarder.Accept()
		if err != nil {
			log.Printf("unable to accept forward request: %+v", err)
			return
		}

		// handle request
		go func(c net.Conn) {
			stream, err := tunnel.OpenStream()
			if err != nil {
				log.Printf("unable to open tunnel stream: %+v", err)
				return
			}

			bidirectionalCopy(c, stream)
		}(c)
	}
}

func main() {
	// setup ssh tunneler
	t, err := tunneler.New()
	if err != nil {
		log.Fatalf("unable to setup tunneler: %+v", err)
	}
	defer t.Invalidate()

	// setup forwarding listener
	forwarder, err := net.Listen("tcp", net.JoinHostPort("localhost", *forwardingPort))
	if err != nil {
		log.Fatalf("failed to start forwarder: %+v", err)
	}
	defer forwarder.Close()
	log.Printf("forwarder listening on %s", forwarder.Addr().String())

	// setup tunnel listener
	tunnel, err := net.Listen("tcp", "")
	if err != nil {
		log.Fatalf("failed to start tunnel: %+v", err)
	}
	defer tunnel.Close()
	tunnelAddr := tunnel.Addr().String()
	log.Printf("tunnel listening on %s", tunnelAddr)

	// establish tunnel connection
	go t.Connect(tunnelAddr)

	conn, err := tunnel.Accept()
	if err != nil {
		log.Fatalf("failed to accept tunnel connection: %+v", err)
	}
	log.Println("tunnel connection established")
	defer conn.Close()

	session, err := yamux.Client(conn, nil)
	defer session.Close()

	// start forwarder
	go serve(forwarder, session)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	log.Println("received interrupt, stopping proxy")
}
