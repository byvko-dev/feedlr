package processing

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/byvko-dev/feedlr/scheduler/utils"
	"github.com/mmcdole/gofeed"
	"github.com/stretchr/testify/assert"
)

func TestParseFeedItems(t *testing.T) {
	feedURL := "https://rsshub.app/twitter/user/cnn/readable=1&addLinkForPics=1&includeRts=0&excludeReplies=1"

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

	assert.Equal(t, 1, len(posts))

	bytes, err := json.MarshalIndent(posts, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	log.Println(string(bytes))
}
