package utils

import (
	"time"

	"github.com/byvko-dev/feedlr/shared/tasks"
	"github.com/mmcdole/gofeed"
)

func GetFeedPosts(feedURL string, cutoff time.Time) ([]tasks.Post, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedURL)
	if err != nil {
		return nil, err
	}

	posts := make([]tasks.Post, 0)
	for _, item := range feed.Items {
		if item.PublishedParsed.After(cutoff) {
			posts = append(posts, tasks.Post{
				Title:       item.Title,
				Description: item.Description,
				Link:        item.Link,
				PubDate:     item.PublishedParsed.Format(time.RFC3339),
			})
		}
	}

	// Reverse the posts so that the most recent one is first
	for i := len(posts)/2 - 1; i >= 0; i-- {
		opp := len(posts) - 1 - i
		posts[i], posts[opp] = posts[opp], posts[i]
	}

	return posts, nil
}
