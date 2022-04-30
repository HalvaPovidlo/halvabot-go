package command

import (
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/HalvaPovidlo/discordBotGo/pkg/zap"
)

type MessageHandler func(s *discordgo.Session, m *discordgo.MessageCreate)

type Message struct {
	handler MessageHandler
	Name    string
	debug   bool
}

// NewMessageCommand Message.Name should be passed with prefix
func NewMessageCommand(name string, handler MessageHandler, debug bool) *Message {
	return &Message{
		handler: handler,
		Name:    name,
		debug:   debug,
	}
}

// RegisterCommand checks is every message starts with Message.Name and is it self-message than runs Message.handler
func (m *Message) RegisterCommand(s *discordgo.Session, logger zap.Logger) {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.MessageCreate) {
		if i.Author.ID == s.State.User.ID {
			return
		}
		channel, _ := s.Channel(i.ChannelID)
		if (channel.Name == "debug") != m.debug {
			return
		}
		if strings.HasPrefix(i.Content, m.Name) {
			logger.Infow("message command handled",
				"command", m.Name,
				"query", i.Content)
			m.handler(s, i)
		}
	})
}
