package main

var globalMiddleware = []func(ctx Context) bool{
	bannedCheckMiddleware,
}

func bannedCheckMiddleware(ctx Context) bool {
	guild, _ := db.GetGuild(ctx.data.GuildID)
	if guild != nil && guild.IsBanned {
		ctx.Reply("This guild is banned from using this bot")
		return false
	}

	return true
}
