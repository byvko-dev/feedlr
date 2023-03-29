package utils

import (
	"net/http"
	"time"
)

func Fetch(url string) ([]byte, error) {
	// Use a custom HTTP client with a timeout and proxy
	client := &http.Client{
		Timeout:   15 * time.Second,
		Transport: getProxyTransport(),
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	content := make([]byte, resp.ContentLength)
	_, err = resp.Body.Read(content)
	if err != nil {
		return nil, err
	}

	return content, nil
}
