package utils

import (
	"errors"
	"io"
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
	if resp == nil {
		return nil, errors.New("no response")
	}
	defer resp.Body.Close()

	// Read the response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}
