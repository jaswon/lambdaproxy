package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"path"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/yamux"

	"cmd/host/key"
	"cmd/shared"
)

const forwardingAddr = "localhost:6789"

var client *lambda.Lambda

func init() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	client = lambda.New(sess, &aws.Config{})
	log.Println("client initialized")
}

func getIP() (string, error) {
	resp, err := http.Get("https://checkip.amazonaws.com")
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bytes.TrimSpace(body)), err
}

func startForwarder(tunnel *yamux.Session) net.Listener {
	forwarder, err := net.Listen("tcp", forwardingAddr)
	if err != nil {
		log.Fatalf("failed to start forwarder: %+v", err)
	}
	log.Printf("forwarder started on %s", forwarder.Addr().String())

	go func() {
		for {
			c, err := forwarder.Accept()
			if err != nil {
				if netErr, ok := err.(net.Error); ok {
					if netErr.Timeout() && netErr.Temporary() {
						log.Println("timed out, continuing")
						continue
					}
				}

				log.Printf("unable to accept forward request: %+v", err)
				return
			}

			// handle request
			go func() {
				stream, err := tunnel.OpenStream()
				if err != nil {
					log.Printf("unable to open tunnel stream: %+v", err)
					return
				}

				bidirectionalCopy(c, stream)
			}()
		}
	}()

	return forwarder
}

func handleTunnel(tunnelListener net.Listener) {
	conn, err := tunnelListener.Accept()
	if err != nil {
		log.Fatalf("failed to accept tunnel connection: %+v", err)
	}
	log.Println("tunnel connection established")
	defer conn.Close()

	session, err := yamux.Client(conn, nil)
	defer session.Close()

	forwarder := startForwarder(session)
	defer forwarder.Close()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	log.Println("received interrupt, stopping proxy")
}

func startExitNode(host, username, keystring, tunnel string) {
	payload, err := json.Marshal(shared.Request{
		Address: host,
		User:    username,
		Key:     keystring,
		Tunnel:  tunnel,
	})
	if err != nil {
		log.Fatalf("unable to marshal request: %v", err)
	}

	_, err = client.Invoke(&lambda.InvokeInput{
		FunctionName: aws.String("proxy"),
		Payload:      payload,
	})
	if err != nil {
		log.Fatalf("lambda invoke failed: %v", err)
	}
}

func main() {
	hostIP, err := getIP()
	if err != nil {
		log.Fatalf("cant get ip: %+v", err)
	}
	log.Printf("public ip: %s", hostIP)

	curUser, err := user.Current()
	if err != nil {
		log.Fatalf("cant get current user: %+v", err)
	}
	log.Printf("current user: %s", curUser.Username)

	pk, err := key.New(path.Join(curUser.HomeDir, ".ssh/authorized_keys"))
	if err != nil {
		log.Fatalf("cant initialize ssh key: %+v", err)
	}
	defer pk.Invalidate()

	tunnelListener, err := net.Listen("tcp", "")
	if err != nil {
		log.Fatalf("cant start tunnel listener: %+v", err)
	}
	tunnelAddr := tunnelListener.Addr().String()
	log.Printf("listening for tunnel connections on %s", tunnelAddr)

	go handleTunnel(tunnelListener)

	startExitNode(
		hostIP,
		curUser.Username,
		pk.Private(),
		tunnelAddr,
	)
}
