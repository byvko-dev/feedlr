package handlers

import (
	"errors"
	"log"

	"github.com/byvko-dev/feedlr/bot/database"
	"github.com/byvko-dev/feedlr/bot/router"
	prisma "github.com/byvko-dev/feedlr/prisma/client"
)

func init() {
	router.Handler("remove", func(ctx router.Context) error {
		link, ok := router.GetOptionDataValue[string](ctx, "url")
		if !ok {
			return ctx.Reply("Missing required argument `url`")
		}

		// Get the guild from the database
		channel, err := ctx.Channel()
		if err != nil {
			log.Printf("Failed to get channel from Discord: %v", err)
			return ctx.Reply("Failed to get channel from Discord")
		}

		// Get webhooks from the database
		webhooks, err := database.Client.FindWebhooksByChannelID(channel.ID)
		if err != nil {
			if errors.Is(err, prisma.ErrNotFound) {
				return ctx.Reply("This channel is not subscribed to any feeds")
			}
			log.Printf("Failed to get feed from the database: %v", err)
			return ctx.Reply("Failed to get feed from the database")
		}

		if len(webhooks) == 0 {
			return ctx.Reply("This channel is not subscribed to any feeds")
		}

		for _, webhook := range webhooks {
			if webhook.Feed().URL == link {
				err := database.Client.DeleteWebhook(webhook.ID)
				if err != nil {
					log.Printf("Failed to delete webhook from the database: %v", err)
					return ctx.Reply("Failed to delete webhook from the database")
				}

				err = ctx.DeleteWebhook(webhook.ExternalID)
				if err != nil {
					log.Printf("Failed to delete webhook: %v", err)
					return ctx.Reply("Failed to delete webhook from Discord")
				}
				return ctx.Reply("Successfully removed feed")
			}
		}

		return ctx.Reply("This channel is not subscribed to this feed")
	})
}
