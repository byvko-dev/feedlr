package tasks

import (
	"encoding/json"
	"fmt"
	"sync"

	prisma "github.com/byvko-dev/feedlr/prisma/client"
	"github.com/byvko-dev/feedlr/shared/messaging"
	"github.com/byvko-dev/feedlr/shared/tasks"
)

func webhookPostsToTasks(feed prisma.FeedModel, wh prisma.WebhookModel, posts []tasks.Post) []tasks.Task {
	var webhookTasks []tasks.Task
	for _, post := range posts {
		task := tasks.Task{
			FeedID:      feed.ID,
			WebhookURL:  fmt.Sprintf("%s/webhooks/%s/%s", apiURL, wh.ExternalID, wh.Token),
			WebhookName: wh.Name,
			Post:        post,
		}
		webhookTasks = append(webhookTasks, task)
	}
	return webhookTasks
}

func feedPostsToTasks(feed prisma.FeedModel, wh []prisma.WebhookModel, posts []tasks.Post) []tasks.Task {
	var wg sync.WaitGroup
	var pendingTasks sliceWithLock[tasks.Task]

	for _, webhook := range wh {
		wg.Add(1) // Start goroutine for each webhook
		go func(feed prisma.FeedModel, webhook prisma.WebhookModel, posts []tasks.Post) {
			defer wg.Done() // Mark webhook goroutine as done

			// Create tasks for each post
			webhookTasks := webhookPostsToTasks(feed, webhook, posts)

			// Add tasks to pending tasks using lock
			pendingTasks.Lock()
			pendingTasks.items = append(pendingTasks.items, webhookTasks...)
			pendingTasks.Unlock()
		}(feed, webhook, posts)
	}

	wg.Wait()
	return pendingTasks.items

}

func newTask(queue string, task tasks.Task) error {
	payload, err := json.Marshal(task)
	if err != nil {
		return err
	}

	mq := messaging.GetClient()
	return mq.Publish(queue, payload, true)
}
