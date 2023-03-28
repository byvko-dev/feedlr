package main

import (
	"github.com/bwmarrin/discordgo"
)

var commands = []*discordgo.ApplicationCommand{}
var helpShards = []string{}

func init() {
	// Help
	helpShards = append(helpShards, ""+
		"Add a new RSS feed to the current channel\n"+
		"Usage: /add <url> [name]\n"+
		"Example: /add url:https://example.com/rss.xml name:Example RSS feed")
	commands = append(commands,
		&discordgo.ApplicationCommand{
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
		},
	)

	// Test
	helpShards = append(helpShards, ""+
		"Test one of the RSS feeds in the current channel\n"+
		"Usage: /test")
	commands = append(commands,
		&discordgo.ApplicationCommand{
			Name:        "test",
			Description: "Test an RSS feed",
		},
	)

	// List
	helpShards = append(helpShards, ""+
		"List all RSS feeds in the current channel\n"+
		"Usage: /list")
	commands = append(commands,
		&discordgo.ApplicationCommand{
			Name:        "list",
			Description: "List all RSS feeds in the current channel",
		},
	)

	// Remove
	helpShards = append(helpShards, ""+
		"Remove one of the RSS feeds from the current channel\n"+
		"Usage: /remove")
	commands = append(commands,
		&discordgo.ApplicationCommand{
			Name:        "remove",
			Description: "Remove an RSS feed from the current channel",
		},
	)

	// Clear
	helpShards = append(helpShards, ""+
		"Remove all RSS feeds from the current channel\n"+
		"Usage: /clear")
	commands = append(commands,
		&discordgo.ApplicationCommand{
			Name:        "clear",
			Description: "Remove all RSS feeds from the current channel",
		},
	)

	// Help
	commands = append(commands,
		&discordgo.ApplicationCommand{
			Name:        "help",
			Description: "Show a help message",
		},
	)
}
