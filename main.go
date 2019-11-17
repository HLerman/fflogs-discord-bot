package main

import (
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/hlerman/fflogs-discord-bot/users"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("settings")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	dg, err := discordgo.New("Bot " + viper.GetString("discordToken"))
	if err != nil {
		log.Fatal("error creating Discord session,", err)
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		log.Fatal("error opening connection,", err)
	}

	// Wait here until CTRL-C or other term signal is received.
	log.Info("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	// If the message is "pong" reply with "Ping!"
	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}

	//if m.Content[:13] == "!addCharacter" {
	r := regexp.MustCompile("^!addCharacter ([0-9]+)$")
	match := r.FindStringSubmatch(m.Content)
	if len(match) > 0 {
		//DEBUG : s.ChannelMessageSend(.ChannelID, match[1])
		id, _ := strconv.Atoi(match[1])
		users.AddCharacter(m, s, id)
	}
	//}
}
