package processing

import (
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/byvko-dev/feedlr/scheduler/utils"
	"github.com/byvko-dev/feedlr/shared/tasks"
	"github.com/mmcdole/gofeed"
)

var converters = map[string]*md.Converter{}

func init() {
	// Description converter
	converter := md.NewConverter("", true, nil)
	converter.AddRules(
		md.Rule{
			// Remove images
			Filter:      []string{"img"},
			Replacement: func(_ string, _ *goquery.Selection, _ *md.Options) *string { return md.String("") },
		},
		md.Rule{
			// Remove links
			Filter:      []string{"a"},
			Replacement: func(content string, _ *goquery.Selection, _ *md.Options) *string { return md.String(content) },
		},
	)
	converters["description"] = converter
}

func parseContent(content string, converter string) (string, error) {
	if converter == "" || converters[converter] == nil {
		return content, nil
	}
	return converters[converter].ConvertString(content)
}

func findImage(content string) string {
	p := strings.NewReader(content)
	doc, err := goquery.NewDocumentFromReader(p)
	if err != nil {
		return ""
	}

	img, _ := doc.Find("img[src]").First().Attr("src")
	return img
}

func findMetadataImageURL(content string) string {
	p := strings.NewReader(content)
	doc, err := goquery.NewDocumentFromReader(p)
	if err != nil {
		return ""
	}

	img, _ := doc.Find("meta[property=\"og:image\"]").First().Attr("content")
	return img
}

func feedItemsToPosts(items []gofeed.Item) ([]tasks.Post, error) {
	// Filter out posts that are older than the cutoff
	posts := make([]tasks.Post, 0)
	for _, item := range items {
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
		if img := findImage(item.Description + " " + item.Content); img != "" { // Check post description, this is likely a thumbnail
			post.Image = img
		} else if item.Image != nil && item.Image.URL != "" { // Check item image, this is likely an avatar
			post.Image = item.Image.URL
		} else if data, _ := utils.Fetch(item.Link); data != nil { // Check page metadata
			post.Image = findMetadataImageURL(string(data))
		}

		posts = append(posts, post)
	}

	// Reverse the posts so that the most recent one is first
	for i := len(posts)/2 - 1; i >= 0; i-- {
		opp := len(posts) - 1 - i
		posts[i], posts[opp] = posts[opp], posts[i]
	}

	return posts, nil
}
