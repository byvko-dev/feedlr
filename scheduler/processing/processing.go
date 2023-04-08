package processing

import (
	"log"
	"time"

	"github.com/byvko-dev/feedlr/scheduler/utils"
	"github.com/byvko-dev/feedlr/shared/tasks"
	"github.com/mmcdole/gofeed"
)

func GetFeedPosts(feedURL string, cutoff time.Time, limit int) ([]tasks.Post, error) {
	data, err := utils.Fetch(feedURL)
	if err != nil {
		return nil, err
	}

	// Parse the feed
	fp := gofeed.NewParser()
	feed, err := fp.ParseString(string(data))
	if err != nil {
		return nil, err
	}

	// Filter out posts that are older than the cutoff
	var items []gofeed.Item
	for i, item := range feed.Items {
		if limit != 0 && i >= limit {
			break
		}

		if item.PublishedParsed.After(cutoff) {
			items = append(items, *item)
		}
	}

	posts, err := feedItemsToPosts(items)
	if err != nil {
		return nil, err
	}

	log.Printf("Found %d posts for %s", len(posts), feedURL)

	return posts, nil
}
