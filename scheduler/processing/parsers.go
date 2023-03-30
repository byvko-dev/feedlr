package processing

import (
	"regexp"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/byvko-dev/feedlr/scheduler/utils"
	"github.com/byvko-dev/feedlr/shared/tasks"
	"github.com/mmcdole/gofeed"
)

var converters = map[string]*md.Converter{}
var regexURL = regexp.MustCompile(`([\w+]+\:\/\/)?([\w\d-]+\.)*[\w-]+[\.\:]\w+([\/\?\=\&\#\.]?[\w-]+)*\/?`)

var imageFetchers = []func(gofeed.Item) string{}

func init() {
	// Set up image fetchers, they are executed in order

	// Check item image, this is likely the intended image
	imageFetchers = append(imageFetchers, func(item gofeed.Item) string {
		if item.Image != nil && item.Image.URL != "" {
			return item.Image.URL
		}
		return ""
	})

	// Check post description, this is likely a thumbnail
	imageFetchers = append(imageFetchers, func(item gofeed.Item) string {
		return findImage(item.Description + " " + item.Content)
	})

	// Check post link, this is likely an external resource
	imageFetchers = append(imageFetchers, func(item gofeed.Item) string {
		url := findUrl(item.Description + " " + item.Content)
		if url == "" {
			return ""
		}
		data, _ := utils.Fetch(url)
		if data == nil {
			return ""
		}
		return findMetadataImageURL(string(data))
	})

	// Check page metadata
	imageFetchers = append(imageFetchers, func(item gofeed.Item) string {
		if item.Link == "" {
			return ""
		}
		data, _ := utils.Fetch(item.Link)
		if data == nil {
			return ""
		}
		return findMetadataImageURL(string(data))
	})
}

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

func findUrl(content string) string {
	return regexURL.FindString(content)
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

		if description != item.Title || item.Content == "" {
			post.Description = description
		} else {
			post.Description = item.Content
		}

		// Set the post's image
		for _, fetcher := range imageFetchers {
			image := fetcher(item)
			if image != "" {
				post.Image = image
				break
			}
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
