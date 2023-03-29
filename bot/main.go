package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	helpers "github.com/byvko-dev/feedlr/shared/helpers"
)

// Bot parameters
var (
	GuildID        = helpers.GetEnv("GUILD_ID", "")                  // Test guild ID. If not passed - bot registers commands globally
	RemoveCommands = helpers.MustGetEnv("REMOVE_COMMANDS") == "true" // Remove all commands after shutdowning or not
)

var s *discordgo.Session

func init() {
	var err error
	s, err = discordgo.New("Bot " + helpers.MustGetEnv("DISCORD_TOKEN"))
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}

	// Register command handlers
	registerHandlers()
}

func main() {
	db.Connect() // Connect to the database
	defer db.Close()

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	if RemoveCommands {
		log.Println("Removing commands...")
		for _, v := range registeredCommands {
			// Only remove commands that were created by this bot
			err := s.ApplicationCommandDelete(s.State.User.ID, GuildID, v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}

	log.Println("Gracefully shutting down.")
}
