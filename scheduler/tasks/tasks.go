package tasks

import (
	"fmt"
	"log"
	"sync"
	"time"

	prisma "github.com/byvko-dev/feedlr/prisma/client"
	"github.com/byvko-dev/feedlr/scheduler/database"
	"github.com/byvko-dev/feedlr/scheduler/processing"
	"github.com/byvko-dev/feedlr/shared/helpers"
	"github.com/byvko-dev/feedlr/shared/tasks"
)

var apiURL = helpers.MustGetEnv("DISCORD_API_URL")

type sliceWithLock[T any] struct {
	sync.Mutex
	items []T
}

func CreateAllFeedsTasks(queue string) {
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
	for _, feed := range feeds {
		// Skip feeds without webhooks
		if len(feed.Webhooks()) == 0 {
			log.Printf("Feed %v has no webhooks, skipping", feed.ID)
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
			err := CreateFeedTasks(queue, feed, lastFetch, 0)
			if err != nil {
				log.Printf("Cannot create tasks for feed %v: %v", feed.ID, err)
				return
			}
		}(feed)
	}

	wg.Wait() // Wait for all feed and webhook goroutines to finish
}

func CreateFeedTasks(queue string, feed prisma.FeedModel, postsCutoff time.Time, limit int) error {
	startTime := time.Now()
	db := database.GetDatabase()

	posts, err := processing.GetFeedPosts(feed.URL, postsCutoff, limit)
	if err != nil {
		return fmt.Errorf("cannot get feed posts: %w", err)
	}

	feedTasks := feedPostsToTasks(feed, feed.Webhooks(), posts)
	if len(feedTasks) == 0 {
		log.Printf("No tasks for feed %v", feed.ID)
		return nil
	}

	// Update feeds last fetched
	defer func(items []tasks.Task) {
		var updatedFeeds []string
		for _, task := range items {
			updatedFeeds = append(updatedFeeds, task.FeedID)
		}
		err = db.UpdateFeedsLastFetched(startTime, updatedFeeds...)
		if err != nil {
			log.Printf("Cannot update feeds last fetched: %v", err)
		}
	}(feedTasks)

	// Send tasks to RabbitMQ
	return newTasks(queue, feedTasks...)
}
