package utils

import (
	"fmt"
	"net/http"
	"net/url"

	helpers "github.com/byvko-dev/feedlr/shared/helpers"
)

var validatorURL = helpers.MustGetEnv("RSS_VALIDATOR_URL")

func ValidateFeed(feedURL *url.URL) bool {
	feedURL.Scheme = "https" // Force HTTPS
	resp, err := http.Get(fmt.Sprintf(validatorURL, feedURL.String()))
	if err != nil {
		return false
	}
	if resp == nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}
