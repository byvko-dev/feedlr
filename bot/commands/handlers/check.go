package handlers

import router "github.com/byvko-dev/feedlr/bot/router"

func init() {
	router.Handler("check", func(ctx router.Context) error {
		// Check if the webhooks are still valid
		return ctx.Reply("This command is not implemented")
	})
}
