package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"

	"github.com/HalvaPovidlo/discordBotGo/cmd/config"
	"github.com/HalvaPovidlo/discordBotGo/internal/music"
	"github.com/HalvaPovidlo/discordBotGo/pkg/zap"
)

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}
	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}
}

func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		panic(err)
	}

	logger := zap.NewLogger()

	session, err := discordgo.New("Bot " + cfg.Discord.Token)
	if err != nil {
		logger.Errorw("error creating Discord session",
			"err", err)
		return
	}
	logger.Infow("Bot initialized")

	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		logger.Infof("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	session.Identify.Intents = discordgo.IntentsAllWithoutPrivileged
	err = session.Open()
	if err != nil {
		logger.Errorw("error opening connection", "err", err)
		return
	}
	logger.Infow("Bot session opened", "SessionID", session.State.SessionID)

	defer func(session *discordgo.Session) {
		_ = session.Close()
		logger.Infow("Bot session closed")
	}(session)

	session.AddHandler(messageCreate)

	musicPlayer := music.NewPlayer(logger)
	musicPlayer.RegisterCommands(session)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	logger.Infow("Gracefully shutdowning")
	defer logger.Sync()
}
