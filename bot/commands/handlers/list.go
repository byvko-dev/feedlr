package handlers

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/byvko-dev/feedlr/bot/database"
	"github.com/byvko-dev/feedlr/bot/router"
	prisma "github.com/byvko-dev/feedlr/prisma/client"
)

func init() {
	router.Handler("list", func(ctx router.Context) error {
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

		var feedNames []string
		for _, webhook := range webhooks {
			feedNames = append(feedNames, fmt.Sprintf("%s (%s)", webhook.Name, webhook.Feed().URL))
		}

		return ctx.Reply(fmt.Sprintf("This channel is subscribed to %d feeds\n```%s```", len(feedNames), strings.Join(feedNames, "\n")))
	})
}
