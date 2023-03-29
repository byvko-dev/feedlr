package router

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

var modalHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){}

func RegisterModalHandlers(s *discordgo.Session) {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if discordgo.InteractionModalSubmit == i.Type {
			data := i.ModalSubmitData()
			for id, handler := range modalHandlers {
				if strings.HasPrefix(data.CustomID, id) {
					handler(s, i)
					break
				}
			}
		}
	})
}

func Modal(id string, handler func(ctx Context) error, middleware ...func(ctx Context) bool) {
	modalHandlers[id] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
					Content: err.Error(),
				},
			})
		}
	}
}
