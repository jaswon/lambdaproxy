package tunnel

import (
	"golang.org/x/crypto/ssh"
)

type Tunnel interface{}

type tunnel struct {
	client *ssh.Client
}

func New(host, user, port string, key []byte) (Tunnel, error) {
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}
	conn, err := ssh.Dial("tcp", host, &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	})
	if err != nil {
		return nil, err
	}

	return tunnel{conn}, nil
}
