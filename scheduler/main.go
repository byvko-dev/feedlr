package main

import (
	"log"
	"time"

	"github.com/byvko-dev/feedlr/scheduler/database"
	"github.com/byvko-dev/feedlr/scheduler/messaging"
	"github.com/byvko-dev/feedlr/scheduler/tasks"
	"github.com/byvko-dev/feedlr/shared/helpers"

	"github.com/go-co-op/gocron"
)

func main() {
	db := database.GetDatabase()
	db.Connect()
	defer db.Close()

	mq := messaging.GetClient()
	mq.Connect()
	defer mq.Close()

	s := gocron.NewScheduler(time.UTC)

	_, err := s.Cron(helpers.MustGetEnv("RSS_PULL_CRON")).Do(createRSSTasksHandler())
	if err != nil {
		log.Fatalf("Cannot create cron job: %v", err)
	}

	log.Println("Starting scheduler...")
	s.StartBlocking()

}

func createRSSTasksHandler() func() {
	now := time.Now()
	lastRun := &now
	return func() {
		log.Printf("Starting tasks. Last run: %v", lastRun.Format(time.RFC3339))
		tasks.CreateRSSTasks(*lastRun)
		now := time.Now()
		lastRun = &now
	}
}
