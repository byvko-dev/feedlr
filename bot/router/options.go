package router

import (
	"github.com/bwmarrin/discordgo"
)

var commandOptions []*discordgo.ApplicationCommand

func RegisterCommandOptions(s *discordgo.Session, guildID string, cleanup bool) ([]*discordgo.ApplicationCommand, error) {
	if cleanup {
		commands, err := s.ApplicationCommands(s.State.User.ID, guildID)
		if err != nil {
			return nil, err
		}

		if guildID != "" {
			// Delete global commands as well
			globalCommands, err := s.ApplicationCommands(s.State.User.ID, "")
			if err != nil {
				return nil, err
			}
			commands = append(commands, globalCommands...)
		}

		for _, v := range commands {
			err := s.ApplicationCommandDelete(s.State.User.ID, guildID, v.ID)
			if err != nil {
				return nil, err
			}
		}
	}

	var registered []*discordgo.ApplicationCommand
	for _, command := range commandOptions {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, command)
		if err != nil {
			return nil, err
		}
		registered = append(registered, cmd)
	}
	return registered, nil
}

func Describe(opts *discordgo.ApplicationCommand) {
	commandOptions = append(commandOptions, opts)
}
