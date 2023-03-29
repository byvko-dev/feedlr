package router

import (
	"fmt"
	"log"
	"runtime/debug"

	"github.com/bwmarrin/discordgo"
)

var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){}

func RegisterHandlers(s *discordgo.Session) {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if discordgo.InteractionApplicationCommand == i.Type {
			if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		}
	})
}

func Handler(name string, handler func(ctx Context) error, middleware ...func(ctx Context) bool) {
	commandHandlers[name] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
