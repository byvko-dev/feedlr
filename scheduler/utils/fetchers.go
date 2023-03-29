package utils

import (
	"net/http"
	"time"

	"github.com/byvko-dev/feedlr/shared/tasks"
	"github.com/mmcdole/gofeed"
)

func GetFeedPosts(feedURL string, cutoff time.Time) ([]tasks.Post, error) {
	// Use a custom HTTP client with a timeout and proxy
	client := &http.Client{
		Timeout:   45 * time.Second,
		Transport: getProxyTransport(),
	}

	resp, err := client.Get(feedURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse the feed
	fp := gofeed.NewParser()
	feed, err := fp.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	// Filter out posts that are older than the cutoff
	posts := make([]tasks.Post, 0)
	for _, item := range feed.Items {
		if item.PublishedParsed.After(cutoff) {
			post := tasks.Post{
				Title:   item.Title,
				Link:    item.Link,
				PubDate: item.PublishedParsed.Format(time.RFC3339),
			}

			// Parse the post's description
			description, err := parseContent(item.Description, "description")
			if err != nil {
				return nil, err
			}
			post.Description = description

			// Set the post's image
			if item.Image != nil {
				post.Image = item.Image.URL
			}
			posts = append(posts, post)
		}
	}

	// Reverse the posts so that the most recent one is first
	for i := len(posts)/2 - 1; i >= 0; i-- {
		opp := len(posts) - 1 - i
		posts[i], posts[opp] = posts[opp], posts[i]
	}

	return posts, nil
}
