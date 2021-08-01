package main

import (
	"log"
	"net/http"
	"time"

	"cmd/shared"

	"golang.org/x/crypto/ssh"

	"github.com/elazarl/goproxy"
	"github.com/hashicorp/yamux"

	"github.com/aws/aws-lambda-go/lambda"
)

func connectSSH(host, user, key string) (*ssh.Client, error) {
	signer, err := ssh.ParsePrivateKey([]byte(key))
	if err != nil {
		return nil, err
	}
	return ssh.Dial("tcp", host, &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	})
}

func getTunnel(client *ssh.Client, tunnel string) (*yamux.Session, error) {
	service, err := client.Dial("tcp", tunnel)
	if err != nil {
		return nil, err
	}

	return yamux.Server(service, nil)
}

func HandleRequest(req shared.Request) error {
	log.Printf("new proxy request, connecting to %s", req.Host)
	client, err := connectSSH(req.Host, req.User, req.Key)
	if err != nil {
		return err
	}
	defer client.Close()

	log.Printf("establishing tunnel on %s", req.Tunnel)
	tunnel, err := getTunnel(client, req.Tunnel)
	if err != nil {
		return err
	}
	defer tunnel.Close()

	log.Println("starting proxy server")
	startTime := time.Now()

	defer log.Printf("closing proxy server after %s", time.Since(startTime).String())
	return http.Serve(tunnel, goproxy.NewProxyHttpServer())
}

func main() {
	lambda.Start(HandleRequest)
}
