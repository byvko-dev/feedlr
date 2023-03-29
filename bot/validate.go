package main

import (
	"fmt"
	"net/http"

	helpers "github.com/byvko-dev/feedlr/shared/helpers"
)

var validatorURL = helpers.MustGetEnv("RSS_VALIDATOR_URL")

func validateFeed(url string) bool {
	resp, err := http.Get(fmt.Sprintf(validatorURL, url))
	if err != nil {
		return false
	}
	if resp == nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}
