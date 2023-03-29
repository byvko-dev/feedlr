package router

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

/*
Returns the value of an option with the given name, or false if the option is not present

This function cannot be a method on the Context struct because it has a type parameter
*/
func GetOptionDataValue[T any](ctx Context, name string) (T, bool) {
	opts := ctx.data.ApplicationCommandData().Options
	if opts == nil {
		return *new(T), false
	}

	for _, option := range opts {
		if option.Name == name {
			value, err := parseOption(ctx, option.Type, option)
			if err != nil {
				return *new(T), false
			}
			return value.(T), true
		}
	}

	return *new(T), false
}

func parseOption(ctx Context, t discordgo.ApplicationCommandOptionType, option *discordgo.ApplicationCommandInteractionDataOption) (any, error) {
	switch t {
	case discordgo.ApplicationCommandOptionString:
		return option.StringValue(), nil
	case discordgo.ApplicationCommandOptionInteger:
		return option.IntValue(), nil
	case discordgo.ApplicationCommandOptionBoolean:
		return option.BoolValue(), nil
	case discordgo.ApplicationCommandOptionUser:
		return option.UserValue(ctx.session), nil
	case discordgo.ApplicationCommandOptionChannel:
		return option.ChannelValue(ctx.session), nil
	case discordgo.ApplicationCommandOptionRole:
		return option.RoleValue(ctx.session, ctx.data.GuildID), nil
	}

	return nil, fmt.Errorf("invalid option type: %v", t)
}
