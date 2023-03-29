package processing

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/byvko-dev/feedlr/scheduler/utils"
	"github.com/mmcdole/gofeed"
)

func TestParseFeedItems(t *testing.T) {
	feedURL := "https://nitter.net/CNN/rss"

	data, err := utils.Fetch(feedURL)
	if err != nil {
		t.Fatal(err)
	}

	fp := gofeed.NewParser()
	feed, err := fp.ParseString(string(data))
	if err != nil {
		t.Fatal(err)
	}

	// Filter out posts that are older than the cutoff
	var items []gofeed.Item
	items = append(items, *feed.Items[0])

	posts, err := feedItemsToPosts(items)
	if err != nil {
		t.Fatal(err)
	}

	bytes, err := json.MarshalIndent(posts, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	log.Println(string(bytes))
}
