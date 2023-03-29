package processing

import (
	"time"

	"github.com/byvko-dev/feedlr/scheduler/utils"
	"github.com/byvko-dev/feedlr/shared/tasks"
	"github.com/mmcdole/gofeed"
)

func GetFeedPosts(feedURL string, cutoff time.Time) ([]tasks.Post, error) {
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
			if img := findImage(item.Description); img != "" { // Check post description, this is likely a thumbnail
				post.Image = img
			} else if item.Image != nil && item.Image.URL != "" { // Check item image, this is likely an avatar
				post.Image = item.Image.URL
			} else if data, _ := utils.Fetch(item.Link); data != nil { // Check page metadata
				post.Image = findMetadataImageURL(string(data))
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
