package main

import (
	"fmt"
	"log"
	"time"

	"github.com/byvko-dev/feedlr/shared/helpers"
	"github.com/byvko-dev/feedlr/shared/tasks"
	"github.com/go-co-op/gocron"
)

var apiURL = helpers.MustGetEnv("DISCORD_API_URL")

func main() {
	db.Connect()
	defer db.Close()
	defer Disconnect() // Close the connection to RabbitMQ

	s := gocron.NewScheduler(time.UTC)
	lastRun := time.Now()
	_, err := s.Cron(helpers.MustGetEnv("RSS_PULL_CRON")).Do(
		func() {
			log.Println("Starting tasks...")
			createTasks(lastRun)
			lastRun = time.Now()
		},
	)
	if err != nil {
		log.Fatalf("Cannot create cron job: %v", err)
	}
	s.StartBlocking()
}

func createTasks(postsSince time.Time) {
	feeds, err := db.GetAllFeeds()
	if err != nil {
		log.Printf("Cannot get feeds: %v", err)
		return
	}

	// Create tasks
	var pendingTasks []tasks.Task
	for _, feed := range feeds {
		posts, err := GetFeedPosts(feed.URL, postsSince)
		if err != nil {
			log.Printf("Cannot get feed posts: %v", err)
			continue
		}

		for _, webhook := range feed.Webhooks() {
			for _, post := range posts {
				pendingTasks = append(pendingTasks, tasks.Task{
					FeedID:     feed.ID,
					WebhookURL: fmt.Sprintf("%s/webhooks/%s/%s", apiURL, webhook.ID, webhook.Token),
					Post:       post,
				})
			}
		}
	}

	// Send tasks to RabbitMQ
	for _, task := range pendingTasks {
		err = NewTask(task)
		if err != nil {
			log.Printf("Cannot send tasks: %v", err)
			continue
		}
	}
}
