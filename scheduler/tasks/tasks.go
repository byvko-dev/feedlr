package tasks

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/byvko-dev/feedlr/scheduler/database"
	"github.com/byvko-dev/feedlr/scheduler/processing"
	"github.com/byvko-dev/feedlr/shared/helpers"
	"github.com/byvko-dev/feedlr/shared/messaging"
	"github.com/byvko-dev/feedlr/shared/tasks"
)

var apiURL = helpers.MustGetEnv("DISCORD_API_URL")

func CreateRSSTasks(queue string, postsSince time.Time) {
	db := database.GetDatabase()
	feeds, err := db.GetAllFeeds()
	if err != nil {
		log.Printf("Cannot get feeds: %v", err)
		return
	}

	// Create tasks
	var pendingTasks []tasks.Task
	for _, feed := range feeds {
		// Skip feeds without webhooks
		if len(feed.Webhooks()) == 0 {
			continue
		}

		posts, err := processing.GetFeedPosts(feed.URL, postsSince)
		if err != nil {
			log.Printf("Cannot get feed posts: %v", err)
			continue
		}

		for _, webhook := range feed.Webhooks() {
			for _, post := range posts {
				pendingTasks = append(pendingTasks, tasks.Task{
					FeedID:      feed.ID,
					WebhookURL:  fmt.Sprintf("%s/webhooks/%s/%s", apiURL, webhook.ExternalID, webhook.Token),
					WebhookName: webhook.Name,
					Post:        post,
				})
			}
		}
	}

	log.Printf("Creating %d tasks", len(pendingTasks))

	// Send tasks to RabbitMQ
	for _, task := range pendingTasks {
		err = newTask(queue, task)
		if err != nil {
			log.Printf("Cannot send tasks: %v", err)
			continue
		}
	}
}

func newTask(queue string, task tasks.Task) error {
	payload, err := json.Marshal(task)
	if err != nil {
		return err
	}

	mq := messaging.GetClient()
	return mq.Publish(queue, payload)
}
