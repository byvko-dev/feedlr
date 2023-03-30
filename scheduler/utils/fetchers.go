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
		Timeout:   45 * time.Second,
		Transport: getProxyTransport(),
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := client.Do(req)
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
