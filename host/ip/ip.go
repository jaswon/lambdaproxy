package ip

import (
	"bytes"
	"io"
	"net/http"
)

func Get() (string, error) {
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
