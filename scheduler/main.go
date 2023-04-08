package main

import (
	"log"
	"time"

	"github.com/byvko-dev/feedlr/scheduler/database"
	"github.com/byvko-dev/feedlr/scheduler/tasks"
	"github.com/byvko-dev/feedlr/shared/helpers"
	"github.com/byvko-dev/feedlr/shared/messaging"

	"github.com/go-co-op/gocron"
)

func main() {
	db := database.GetDatabase()
	db.Connect()
	defer db.Close()

	queueName := helpers.MustGetEnv("TASKS_QUEUE")
	mq := messaging.GetClient()
	mq.Connect(queueName)
	defer mq.Close()

	s := gocron.NewScheduler(time.UTC)

	_, err := s.Cron(helpers.MustGetEnv("RSS_PULL_CRON")).Do(createRSSTasksHandler(queueName))
	if err != nil {
		log.Fatalf("Cannot create cron job: %v", err)
	}

	log.Println("Starting scheduler...")
	s.StartBlocking()

}

func createRSSTasksHandler(queue string) func() {
	return func() {
		log.Println("Creating RSS tasks...")
		tasks.CreateAllFeedsTasks(queue, nil, 0)
	}
}
