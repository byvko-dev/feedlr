package utils

import (
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
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
