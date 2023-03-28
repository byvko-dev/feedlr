package main

import (
	"fmt"
	"net/http"
)

var validatorURL = mustGetEnv("RSS_VALIDATOR_URL")

func validateFeed(url string) bool {
	resp, err := http.Get(fmt.Sprintf(validatorURL, url))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return false
	}
	return true
}
