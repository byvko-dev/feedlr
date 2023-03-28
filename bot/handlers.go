package main

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/bwmarrin/discordgo"
	prisma "github.com/byvko-dev/feedlr/prisma/client"
)

var (
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		// /add handler
		"add": Handle(func(ctx Context) error {
			// Get the URL from the command options
			link, ok := getOptionDataValue[string](ctx, "url")
			if !ok {
				return ctx.Reply("Missing required argument `url`")
			}
			parsed, err := url.ParseRequestURI(link)
			if err != nil {
				return ctx.Reply(fmt.Sprintf("Invalid URL:\n```%v```", err.Error()))
			}

			// Validate the feed
			if !validateFeed(parsed.String()) {
				return ctx.Reply("This is not a valid RSS feed")
			}

			// Get the guild from the database
			dcGuild, err := ctx.Guild()
			if err != nil {
				log.Printf("Failed to get guild from Discord: %v", err)
				return ctx.Reply("Failed to get guild from Discord")
			}
			guild, err := db.GetOrCreateGuild(dcGuild.ID, dcGuild.Name)
			if err != nil {
				log.Printf("Failed to get guild from the database: %v", err)
				return ctx.Reply("Failed to get guild from the database")
			}

			// Get feed from the database
			feed, err := db.GetOrCreateFeed(guild.ID, parsed.String())
			if err != nil {
				log.Printf("Failed to get feed from the database: %v", err)
				return ctx.Reply("Failed to get feed from the database")
			}

			for _, webhook := range feed.Webhooks() {
				if webhook.ChannelID == ctx.data.ChannelID {
					return ctx.Reply("This feed is already added to this channel")
				}
			}

			// Get the channel from the command options
			channel, err := ctx.Channel()
			if err != nil {
				log.Printf("Failed to get channel from Discord: %v", err)
				return ctx.Reply("Failed to get channel from Discord")
			}

			// Create a webhook for the channel
			webhook, err := ctx.session.WebhookCreate(channel.ID, fmt.Sprintf("Feedlr - %v", parsed.String()), "")
			if err != nil {
				log.Printf("Failed to create webhook: %v", err)
				return ctx.Reply("Failed to create webhook")
			}

			// Add the webhook to the feed
			name, _ := getOptionDataValue[string](ctx, "name")
			if name == "" {
				name = fmt.Sprintf("RSS from %s", parsed.Host)
			}
			hook, err := db.CreateWebhook(feed.ID, channel.ID, name, webhook.ID, webhook.Token)
			if err != nil {
				log.Printf("Failed to add webhook to the database: %v", err)
				return ctx.Reply("Failed to add webhook to the database")
			}

			return ctx.Reply(fmt.Sprintf("Successfully added %s (%s) to %s", hook.Name, feed.URL, channel.Mention()))
		}),

		// /test handler
		"test": Handle(func(ctx Context) error {
			return ctx.Reply("This command is not implemented yet")
		}),

		// /list handler
		"list": Handle(func(ctx Context) error {
			// Get the guild from the database
			channel, err := ctx.Channel()
			if err != nil {
				log.Printf("Failed to get channel from Discord: %v", err)
				return ctx.Reply("Failed to get channel from Discord")
			}

			// Get webhooks from the database
			webhooks, err := db.FindWebhooksByChannelID(channel.ID)
			if err != nil {
				if errors.Is(err, prisma.ErrNotFound) {
					return ctx.Reply("This channel is not subscribed to any feeds")
				}
				log.Printf("Failed to get feed from the database: %v", err)
				return ctx.Reply("Failed to get feed from the database")
			}

			var feedToWebhookName = make(map[string]string)
			var feedIds []string
			for _, webhook := range webhooks {
				feedToWebhookName[webhook.FeedID] = webhook.Name
				feedIds = append(feedIds, webhook.FeedID)
			}

			// Get feeds from the database
			feeds, err := db.GetManyFeeds(feedIds)
			if err != nil {
				log.Printf("Failed to get feed from the database: %v", err)
				return ctx.Reply("Failed to get feed from the database")
			}

			var feedNames []string
			for _, feed := range feeds {
				name := feedToWebhookName[feed.ID]
				if name == "" {
					feedNames = append(feedNames, feed.URL)
					continue
				}
				feedNames = append(feedNames, fmt.Sprintf("%s (%s)", name, feed.URL))
			}

			return ctx.Reply(fmt.Sprintf("This channel is subscribed to %d feeds\n```%s```", len(feedNames), strings.Join(feedNames, "\n")))
		}),

		// /remove handler
		"remove": Handle(func(ctx Context) error {
			return ctx.Reply("This command is not implemented")
		}),

		// /clear handler
		"clear": Handle(func(ctx Context) error {
			// Get the guild from the database
			channel, err := ctx.Channel()
			if err != nil {
				log.Printf("Failed to get channel from Discord: %v", err)
				return ctx.Reply("Failed to get channel from Discord")
			}

			webhooks, err := db.FindWebhooksByChannelID(channel.ID)
			if err != nil {
				if errors.Is(err, prisma.ErrNotFound) {
					return ctx.Reply("This channel is not subscribed to any feeds")
				}
				log.Printf("Failed to get feed from the database: %v", err)
				return ctx.Reply("Failed to get feed from the database")
			}

			var failedToDelete []string
			for _, webhook := range webhooks {
				err = ctx.session.WebhookDelete(webhook.ID)
				if err != nil {
					log.Printf("Failed to delete webhook: %v", err)
					failedToDelete = append(failedToDelete, webhook.Name)
				}
			}

			// Get webhooks from the database
			err = db.DeleteWebhooksByChannelID(channel.ID)
			if err != nil {
				if errors.Is(err, prisma.ErrNotFound) {
					return ctx.Reply("This channel is not subscribed to any feeds")
				}
				log.Printf("Failed to get feed from the database: %v", err)
				return ctx.Reply("Failed to get feed from the database")
			}

			if len(failedToDelete) > 0 {
				return ctx.Reply(fmt.Sprintf("Successfully removed %d feeds, but failed to delete %d webhooks\nPlease delete the following Webhooks manually:\n```%s```", len(webhooks), len(failedToDelete), strings.Join(failedToDelete, "\n")))
			}

			return ctx.Reply(fmt.Sprintf("Successfully removed %d feeds", len(webhooks)))
		}),

		// /check handler
		"check": Handle(func(ctx Context) error {
			// Check if the webhooks are still valid
			return ctx.Reply("This command is not implemented")
		}),

		// /help handler
		"help": Handle(func(ctx Context) error {
			return ctx.Reply(strings.Join(helpShards, "\n\n"))
		}),
	}
)

func registerHandlers() {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}
