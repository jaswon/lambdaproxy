package tunneler

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os/user"
	"path"

	"cmd/host/ip"
	"cmd/host/key"
	"cmd/shared"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
)

var client *lambda.Lambda

func init() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	client = lambda.New(sess, &aws.Config{})
	log.Println("client initialized")
}

type Tunneler interface {
	Connect(string)
	Invalidate()
}

type tunneler struct {
	host string
	user *user.User
	key  key.Key
}

func (t *tunneler) Connect(tunnelAddr string) {
	payload, err := json.Marshal(shared.Request{
		Host:   t.host,
		Tunnel: tunnelAddr,
		Key:    t.key.Private(),
		User:   t.user.Username,
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

func (t *tunneler) Invalidate() {
	t.key.Invalidate()
}

func New() (Tunneler, error) {
	hostIP, err := ip.Get()
	if err != nil {
		return nil, fmt.Errorf("cant get ip: %w", err)
	}
	log.Printf("public ip: %s", hostIP)

	curUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("cant get current user: %w", err)
	}
	log.Printf("current user: %s", curUser.Username)

	pk, err := key.New(path.Join(curUser.HomeDir, ".ssh/authorized_keys"))
	if err != nil {
		return nil, fmt.Errorf("cant initialize ssh key: %w", err)
	}

	return &tunneler{
		host: net.JoinHostPort(hostIP, "22"),
		user: curUser,
		key:  pk,
	}, nil
}
