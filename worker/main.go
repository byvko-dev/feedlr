package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/byvko-dev/feedlr/shared/helpers"
	"github.com/byvko-dev/feedlr/shared/messaging"
	"github.com/byvko-dev/feedlr/shared/tasks"
)

type DiscordEmbed struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url,omitempty"`
	Timestamp   string `json:"timestamp,omitempty"`
	Color       int    `json:"color,omitempty"`
	Image       struct {
		URL string `json:"url,omitempty"`
	} `json:"image,omitempty"`
}
type DiscordWebhookPayload struct {
	Content  string         `json:"content,omitempty"`
	Username string         `json:"username,omitempty"`
	Embeds   []DiscordEmbed `json:"embeds,omitempty"`
}

type DiscordWebhookResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func main() {
	queueName := helpers.MustGetEnv("TASKS_QUEUE")
	mq := messaging.GetClient()
	mq.Connect(queueName)

	cancel := make(chan struct{})
	mq.Subscribe(queueName, func(body []byte) {
		// Unmarshal the task
		var task tasks.Task
		err := json.Unmarshal(body, &task)
		if err != nil {
			log.Printf("Failed to unmarshal task: %v", err)
			return
		}

		var embed DiscordEmbed
		embed.Title = task.Post.Title
		embed.Description = task.Post.Description
		embed.URL = task.Post.Link
		embed.Timestamp = task.Post.PubDate
		if task.Post.Image != "" {
			embed.Image.URL = task.Post.Image
		}

		var data DiscordWebhookPayload
		data.Embeds = append(data.Embeds, embed)
		data.Username = task.WebhookName

		// Convert the payload to JSON
		payload, err := json.Marshal(data)
		if err != nil {
			log.Printf("Failed to marshal payload for feed %v: %v", task.FeedID, err)
			return
		}

		// Send the payload to Discord
		res, err := http.Post(task.WebhookURL, "application/json", bytes.NewBuffer(payload))
		if err != nil {
			log.Printf("Failed to POST webhook for feed %v: %v", task.FeedID, err)
			return
		}
		defer res.Body.Close()

		if res.StatusCode == 204 {
			return
		}

		// Decode the response
		var response map[string]any
		err = json.NewDecoder(res.Body).Decode(&response)
		if err != nil {
			log.Printf("Failed to decode response for feed %v: %v", task.FeedID, err)
			return
		}

		// Check the response
		if response["message"] != "" {
			log.Printf("Failed to POST webhook for feed %v: %v\n%+v", task.FeedID, response["message"], response)
			return
		}

		log.Printf("Failed to POST webhook for feed %v: bad status code, no error.", task.FeedID)
	}, cancel)
}
