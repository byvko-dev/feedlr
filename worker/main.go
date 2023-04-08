package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

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
	Thumbnail struct {
		URL string `json:"url,omitempty"`
	} `json:"thumbnail,omitempty"`
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
	rateLimitStr := helpers.GetEnv("TASKS_QUEUE_PREFETCH", "10")
	rateLimit, err := strconv.Atoi(rateLimitStr)
	if err != nil {
		log.Printf("Failed to parse TASKS_QUEUE_PREFETCH: %v", err)
		rateLimit = 10
	}
	mq := messaging.GetClient()
	mq.Connect(queueName)

	cancel := make(chan struct{})
	mq.Subscribe(queueName, rateLimit, func(body []byte) {
		// Sleep for 1 second to avoid rate limiting
		start := time.Now()
		defer func() {
			if time.Since(start) < 1*time.Second {
				time.Sleep(time.Second - time.Since(start))
			}
		}()

		// Unmarshal the task
		var task tasks.Task
		err := json.Unmarshal(body, &task)
		if err != nil {
			log.Printf("Failed to unmarshal task: %v\n%v", err, string(body))
			return
		}

		// Convert the payload to JSON
		payload, err := json.MarshalIndent(taskToDiscordPayload(task), "", "  ")
		if err != nil {
			log.Printf("Failed to marshal payload for feed %v: %v", task.FeedID, err)
			return
		}

		err = postToDiscordWebhook(task.WebhookURL, payload)
		if err != nil {
			log.Printf("Failed to POST webhook for feed %v: %v", task.FeedID, err)
			return
		}

		log.Printf("Created a post for feed %v\n%v", task.FeedID, string(payload))

	}, cancel)
}

func taskToDiscordPayload(task tasks.Task) DiscordWebhookPayload {
	var embed DiscordEmbed
	embed.Title = task.Post.Title
	embed.Description = task.Post.Description
	embed.URL = task.Post.Link
	embed.Timestamp = task.Post.PubDate
	embed.Image.URL = task.Post.Image

	var data DiscordWebhookPayload
	data.Embeds = append(data.Embeds, embed)
	data.Username = task.WebhookName
	return data
}

func postToDiscordWebhook(url string, payload []byte) error {
	// Send the payload to Discord
	res, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == 204 {
		return nil
	}

	// Decode the response
	var response map[string]any
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return err
	}

	// Check the response
	if response["message"] != "" {
		return errors.New(response["message"].(string))
	}

	return errors.New("bad status code, no error")
}
