package main

import (
	"encoding/json"
	"log"
	"os"
	"testing"

	"github.com/byvko-dev/feedlr/shared/tasks"
)

const input = `{
	"image": "https://openaicom.imgix.net/e2d6b8cb-8def-4aa7-854f-58f9f5261a5c/march-20-chatgpt-outage.png?auto=compress%2Cformat\u0026fit=min\u0026fm=jpg\u0026q=80\u0026rect=%2C%2C%2C",
	"link": "https://openai.com/blog/march-20-chatgpt-outage",
	"title": "March 20 ChatGPT outage: Here's what happened",
	"description": "An update on our findings, the actions we've taken, and technical details of the bug.",
	"pub_date": "2023-03-24T07:00:00Z"
}`

func TestPostToDiscordWebhook(t *testing.T) {
	var post tasks.Post
	err := json.Unmarshal([]byte(input), &post)
	if err != nil {
		t.Fatal(err)
	}

	task := tasks.Task{
		FeedID:      "123",
		WebhookURL:  os.Getenv("TEST_WEBHOOK_URL"),
		WebhookName: "Test",
		Post:        post,
	}

	bytes, err := json.MarshalIndent(taskToDiscordPayload(task), "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	log.Println(string(bytes))

	err = postToDiscordWebhook(task.WebhookURL, bytes)
	if err != nil {
		t.Fatal(err)
	}

	log.Println("Success!")
}
