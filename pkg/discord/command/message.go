package command

import (
	"context"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/HalvaPovidlo/discordBotGo/pkg/contexts"
	"github.com/HalvaPovidlo/discordBotGo/pkg/discord"
	"github.com/bwmarrin/discordgo"
)

const maxMessageLength = 50

type MessageHandler func(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate)

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
func (m *Message) RegisterCommand(s *discordgo.Session, logger *zap.Logger) {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.MessageCreate) {
		if i.Author.ID == s.State.User.ID {
			return
		}
		if (i.ChannelID == discord.ChannelDebugID) != m.debug {
			return
		}
		if len(i.Content) <= maxMessageLength {
			i.Content = strings.ToLower(i.Content)
		}
		if strings.HasPrefix(i.Content, m.Name) {
			ctx := contexts.WithValues(context.Background(), logger, "")
			log := contexts.GetLogger(ctx)
			log.Info("message command handled",
				zap.String("command", m.Name),
				zap.String("query", i.Content))
			start := time.Now()
			m.handler(ctx, s, i)
			log.Info("command executed",
				zap.String("command", m.Name),
				zap.Duration("elapsed", time.Since(start)))
		}
	})
}
