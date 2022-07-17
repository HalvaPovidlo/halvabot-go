package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"

	"github.com/HalvaPovidlo/discordBotGo/internal/chess/lichess"
	"github.com/HalvaPovidlo/discordBotGo/pkg/contexts"
	"github.com/HalvaPovidlo/discordBotGo/pkg/discord"
	"github.com/HalvaPovidlo/discordBotGo/pkg/discord/command"
)

const chess = "chess"

type chessClient interface {
	StartOpenGame(ctx context.Context) (*lichess.OpenGameResponse, error)
}

type Service struct {
	client chessClient
	prefix string
}

func NewCog(prefix string, client chessClient) *Service {
	s := Service{
		prefix: prefix,
		client: client,
	}

	return &s
}

func (s *Service) RegisterCommands(session *discordgo.Session, debug bool, logger *zap.Logger) {
	command.NewMessageCommand(s.prefix+chess, s.chessMessageHandler, debug).RegisterCommand(session, logger)
}

func (s *Service) chessMessageHandler(ctx context.Context, session *discordgo.Session, m *discordgo.MessageCreate) {
	game, err := s.client.StartOpenGame(ctx)
	if err != nil {
		s.sendInternalErrorMessage(ctx, session, m)
		return
	}
	go session.ChannelMessageSend(m.ChannelID, game.Challenge.URL)
}

func (s *Service) sendInternalErrorMessage(ctx context.Context, ds *discordgo.Session, m *discordgo.MessageCreate) {
	go func() {
		_, err := ds.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{Content: discord.MessageInternalError})
		if err != nil {
			contexts.GetLogger(ctx).Error("sending message",
				zap.String("channel", m.ChannelID),
				zap.Error(err))
		}
	}()
}
