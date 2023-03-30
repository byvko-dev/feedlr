package tasks

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	prisma "github.com/byvko-dev/feedlr/prisma/client"
	"github.com/byvko-dev/feedlr/scheduler/database"
	"github.com/byvko-dev/feedlr/scheduler/processing"
	"github.com/byvko-dev/feedlr/shared/helpers"
	"github.com/byvko-dev/feedlr/shared/messaging"
	"github.com/byvko-dev/feedlr/shared/tasks"
)

var apiURL = helpers.MustGetEnv("DISCORD_API_URL")

type sliceWithLock[T any] struct {
	sync.Mutex
	items []T
}

func CreateRSSTasks(queue string) {
	// This is used as lastFetch for all feeds with posts
	// it is very likely unnecessary, but it is here to not miss any super quick RSS updates
	jobStartTime := time.Now()

	db := database.GetDatabase()
	feeds, err := db.GetAllFeeds()
	if err != nil {
		log.Printf("Cannot get feeds: %v", err)
		return
	}

	// Create tasks
	var wg sync.WaitGroup
	var pendingTasks sliceWithLock[tasks.Task]
	for _, feed := range feeds {
		// Skip feeds without webhooks
		if len(feed.Webhooks()) == 0 {
			continue
		}

		wg.Add(1) // Start goroutine for each feed
		go func(feed prisma.FeedModel) {
			defer wg.Done() // Mark feed goroutine as done

			lastFetch, ok := feed.LastFetch()
			if !ok {
				// If feed has never been fetched, set last fetch to now and skip
				err = db.UpdateFeedsLastFetched(jobStartTime, feed.ID)
				if err != nil {
					log.Printf("Cannot update feed last fetched: %v", err)
				}
				return
			}

			posts, err := processing.GetFeedPosts(feed.URL, lastFetch)
			if err != nil {
				log.Printf("Cannot get feed posts: %v", err)
				return
			}

			for _, webhook := range feed.Webhooks() {
				wg.Add(1) // Start goroutine for each webhook
				go func(feed prisma.FeedModel, webhook prisma.WebhookModel, posts []tasks.Post) {
					defer wg.Done() // Mark webhook goroutine as done

					// Create tasks for each post
					var webhookTasks []tasks.Task
					for _, post := range posts {
						task := tasks.Task{
							FeedID:      feed.ID,
							WebhookURL:  fmt.Sprintf("%s/webhooks/%s/%s", apiURL, webhook.ExternalID, webhook.Token),
							WebhookName: webhook.Name,
							Post:        post,
						}
						webhookTasks = append(webhookTasks, task)
					}

					// Add tasks to pending tasks using lock
					pendingTasks.Lock()
					pendingTasks.items = append(pendingTasks.items, webhookTasks...)
					pendingTasks.Unlock()
				}(feed, webhook, posts)
			}
		}(feed)
	}

	wg.Wait() // Wait for all feed and webhook goroutines to finish

	if len(pendingTasks.items) == 0 {
		log.Printf("No tasks to create from %v feeds\n", len(feeds))
		return
	}

	// Update feeds last fetched
	defer func(items []tasks.Task) {
		var updatedFeeds []string
		for _, task := range items {
			updatedFeeds = append(updatedFeeds, task.FeedID)
		}
		err = db.UpdateFeedsLastFetched(jobStartTime, updatedFeeds...)
		if err != nil {
			log.Printf("Cannot update feeds last fetched: %v", err)
		}
	}(pendingTasks.items)

	log.Printf("Creating %d tasks", len(pendingTasks.items))

	// Send tasks to RabbitMQ
	for _, task := range pendingTasks.items {
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
