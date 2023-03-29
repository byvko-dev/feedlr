package handlers

import (
	"fmt"
	"log"
	"net/url"

	"github.com/byvko-dev/feedlr/bot/database"
	router "github.com/byvko-dev/feedlr/bot/router"
	"github.com/byvko-dev/feedlr/bot/utils"
)

func init() {
	router.Handler("add", func(ctx router.Context) error {
		// Get the URL from the command options
		link, ok := router.GetOptionDataValue[string](ctx, "url")
		if !ok {
			return ctx.Reply("Missing required argument `url`")
		}
		parsed, err := url.ParseRequestURI(link)
		if err != nil {
			return ctx.Reply(fmt.Sprintf("Invalid URL:\n```%v```", err.Error()))
		}

		// Validate the feed
		if !utils.ValidateFeed(parsed) {
			return ctx.Reply("This is not a valid RSS feed")
		}

		// Get the guild from the database
		dcGuild, err := ctx.Guild()
		if err != nil {
			log.Printf("Failed to get guild from Discord: %v", err)
			return ctx.Reply("Failed to get guild from Discord")
		}
		guild, err := database.Client.GetOrCreateGuild(dcGuild.ID, dcGuild.Name)
		if err != nil {
			log.Printf("Failed to get guild from the database: %v", err)
			return ctx.Reply("Failed to get guild from the database")
		}

		channel, err := ctx.Channel()
		if err != nil {
			log.Printf("Failed to get channel from Discord: %v", err)
			return ctx.Reply("Failed to get channel from Discord")
		}

		// Get feed from the database
		feed, err := database.Client.GetOrCreateFeed(guild.ID, parsed.String())
		if err != nil {
			log.Printf("Failed to get feed from the database: %v", err)
			return ctx.Reply("Failed to get feed from the database")
		}

		for _, webhook := range feed.Webhooks() {
			if webhook.ChannelID == channel.ID {
				return ctx.Reply("This feed is already added to this channel")
			}
		}

		// Webhook name
		webhookName, _ := router.GetOptionDataValue[string](ctx, "name")
		if webhookName == "" {
			webhookName = fmt.Sprintf("RSS from %s", parsed.Host)
		}

		// Create a webhook for the channel
		webhook, err := ctx.CreateWebhook(fmt.Sprintf("Feedlr - %v", webhookName), "")
		if err != nil {
			log.Printf("Failed to create webhook: %v", err)
			return ctx.Reply("Failed to create webhook")
		}

		// Add the webhook to the feed

		hook, err := database.Client.CreateWebhook(feed.ID, channel.ID, webhookName, webhook.ID, webhook.Token)
		if err != nil {
			log.Printf("Failed to add webhook to the database: %v", err)
			return ctx.Reply("Failed to add webhook to the database")
		}

		return ctx.Reply(fmt.Sprintf("Successfully added %s (%s) to %s", hook.Name, feed.URL, channel.Mention()))
	})
}
