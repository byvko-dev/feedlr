package utils

import (
	"crypto/tls"
	"log"
	"net/http"
	"net/url"

	"github.com/byvko-dev/feedlr/shared/helpers"
)

func getProxyTransport() *http.Transport {
	proxyURL := helpers.GetEnv("PROXY_URL", "")
	if proxyURL == "" {
		return nil
	}

	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		log.Printf("Failed to parse proxy URL: %v\n", err)
		return nil
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Proxy:           http.ProxyURL(parsedURL),
	}
	return transport
}
