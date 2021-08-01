package ip

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
)

const checkUrl = "https://checkip.amazonaws.com"

func ProxyGet(proxy string) (string, error) {
	proxyUrl, err := url.Parse(proxy)
	if err != nil {
		return "", err
	}

	return get(&http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		},
	})
}

func Get() (string, error) {
	return get(http.DefaultClient)
}

func get(client *http.Client) (string, error) {
	resp, err := client.Get(checkUrl)
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
