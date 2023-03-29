package handlers

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/byvko-dev/feedlr/bot/database"
	router "github.com/byvko-dev/feedlr/bot/router"
	prisma "github.com/byvko-dev/feedlr/prisma/client"
)

func init() {
	router.Handler("clear", func(ctx router.Context) error {
		// Get the guild from the database
		channel, err := ctx.Channel()
		if err != nil {
			log.Printf("Failed to get channel from Discord: %v", err)
			return ctx.Reply("Failed to get channel from Discord")
		}

		webhooks, err := database.Client.FindWebhooksByChannelID(channel.ID)
		if err != nil {
			if errors.Is(err, prisma.ErrNotFound) {
				return ctx.Reply("This channel is not subscribed to any feeds")
			}
			log.Printf("Failed to get feed from the database: %v", err)
			return ctx.Reply("Failed to get feed from the database")
		}

		var failedToDelete []string
		for _, webhook := range webhooks {
			err = ctx.DeleteWebhook(webhook.ExternalID)
			if err != nil {
				log.Printf("Failed to delete webhook: %v", err)
				failedToDelete = append(failedToDelete, webhook.Name)
			}
		}

		// Get webhooks from the database
		err = database.Client.DeleteWebhooksByChannelID(channel.ID)
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
	})

}
