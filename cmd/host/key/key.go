package key

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"golang.org/x/crypto/ssh"
)

type Key interface {
	Private() []byte
	Invalidate() error
}

type key struct {
	authfile string
	private  []byte
	public   []byte
}

func (k *key) Private() []byte {
	return k.private
}

func (k *key) Invalidate() error {
	f, err := os.ReadFile(k.authfile)
	if err != nil {
		return err
	}

	removed := bytes.ReplaceAll(f, k.public, []byte{})

	return os.WriteFile(k.authfile, removed, 0600)
}

func New(authfile string) (Key, error) {
	private, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, errors.New("cannot generate key")
	}
	privateBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(private),
	})

	public, err := ssh.NewPublicKey(private.Public())
	if err != nil {
		return nil, errors.New("cannot get public ssh key")
	}
	publicBytes := ssh.MarshalAuthorizedKey(public)

	contents, err := os.ReadFile(authfile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("cannot read authorized keys: %w", err)
		}
	}

	if bytes.Index(contents, publicBytes) != -1 {
		return nil, errors.New("key already exists")
	}

	added := append(contents, publicBytes...)
	err = os.WriteFile(authfile, added, 0644)
	if err != nil {
		return nil, fmt.Errorf("cannot write to authorized keys: %w", err)
	}

	return &key{
		authfile: authfile,
		private:  privateBytes,
		public:   publicBytes,
	}, nil
}
