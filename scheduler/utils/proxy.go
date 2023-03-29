package utils

import (
	"log"
	"net/http"
	"net/url"

	"github.com/byvko-dev/feedlr/shared/helpers"
)

func getProxy() func(*http.Request) (*url.URL, error) {
	proxyURL := helpers.GetEnv("PROXY_URL", "")
	if proxyURL == "" {
		return nil
	}
	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		log.Printf("Failed to parse proxy URL: %v\n", err)
		return nil
	}
	return http.ProxyURL(parsedURL)
}
