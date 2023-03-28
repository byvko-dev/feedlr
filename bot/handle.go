package main

import (
	"fmt"
	"log"
	"runtime/debug"

	"github.com/bwmarrin/discordgo"
)

func Handle(handler func(ctx Context) error, middleware ...func(ctx Context) bool) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// Recover from panics
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic while handling command: %v", r)
				log.Printf("Stack trace: %v", string(debug.Stack()))

				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Internal server error",
					},
				})
			}
		}()

		ctx := Context{
			session: s,
			data:    i,
		}
		// Run middleware
		for _, m := range append(globalMiddleware, middleware...) {
			if !m(ctx) {
				return
			}
		}
		// Handle command
		err := handler(ctx)

		// Handle errors
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("```%v```", err.Error()),
				},
			})
		}
	}
}
