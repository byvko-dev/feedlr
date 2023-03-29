package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	helpers "github.com/byvko-dev/feedlr/shared/helpers"

	_ "github.com/byvko-dev/feedlr/bot/commands/handlers" // Register command handlers
	_ "github.com/byvko-dev/feedlr/bot/commands/options"  // Register command options

	"github.com/byvko-dev/feedlr/bot/database"
	"github.com/byvko-dev/feedlr/bot/router"
)

// Bot parameters
var (
	GuildID         = helpers.GetEnv("GUILD_ID", "")                       // Test guild ID. If not passed - bot registers commands globally
	CleanupCommands = helpers.GetEnv("CLEANUP_COMMANDS", "true") == "true" // Remove all commands before registering new ones
)

var s *discordgo.Session

func init() {
	var err error
	s, err = discordgo.New("Bot " + helpers.MustGetEnv("DISCORD_TOKEN"))
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

func main() {
	database.Client.Connect() // Connect to the database
	defer database.Client.Close()

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	defer s.Close()

	// Register command handlers
	router.RegisterCommandOptions(s, GuildID, CleanupCommands)
	router.RegisterHandlers(s)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	log.Println("Gracefully shutting down.")
}
