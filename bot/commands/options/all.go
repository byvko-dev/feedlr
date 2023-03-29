package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/byvko-dev/feedlr/bot/router"
)

func init() {
	router.Describe(&discordgo.ApplicationCommand{
		Name:        "add",
		Description: "Add a new RSS feed to the current channel",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "url",
				Description: "URL of the RSS feed",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "name",
				Description: "Optional name of the feed",
				Required:    false,
			},
		},
	})

	router.Describe(&discordgo.ApplicationCommand{
		Name:        "list",
		Description: "List all RSS feeds in the current channel",
	})

	router.Describe(&discordgo.ApplicationCommand{
		Name:        "remove",
		Description: "Remove an RSS feed from the current channel",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "url",
				Description: "URL of the RSS feed",
				Required:    true,
			},
		},
	})

	router.Describe(&discordgo.ApplicationCommand{
		Name:        "clear",
		Description: "Remove all RSS feeds from the current channel",
	})
}
