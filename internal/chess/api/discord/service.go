package discord

import (
	"github.com/bwmarrin/discordgo"

	"github.com/HalvaPovidlo/discordBotGo/internal/chess/lichess"
	"github.com/HalvaPovidlo/discordBotGo/pkg/contexts"
	"github.com/HalvaPovidlo/discordBotGo/pkg/discord"
	"github.com/HalvaPovidlo/discordBotGo/pkg/discord/command"
	"github.com/HalvaPovidlo/discordBotGo/pkg/zap"
)

const chess = "chess"

type chessClient interface {
	StartOpenGame() (*lichess.OpenGameResponse, error)
}

type Service struct {
	ctx    contexts.Context
	client chessClient
	prefix string
	logger zap.Logger
}

func NewCog(ctx contexts.Context, prefix string, client chessClient, logger zap.Logger) *Service {
	s := Service{
		ctx:    ctx,
		prefix: prefix,
		client: client,
		logger: logger,
	}

	return &s
}

func (s *Service) RegisterCommands(session *discordgo.Session, debug bool, logger zap.Logger) {
	command.NewMessageCommand(s.prefix+chess, s.chessMessageHandler, debug).RegisterCommand(session, logger)
}

func (s *Service) chessMessageHandler(session *discordgo.Session, m *discordgo.MessageCreate) {
	game, err := s.client.StartOpenGame()
	if err != nil {
		s.sendInternalErrorMessage(session, m)
		return
	}
	go session.ChannelMessageSend(m.ChannelID, game.Challenge.URL)
}

func (s *Service) sendInternalErrorMessage(ds *discordgo.Session, m *discordgo.MessageCreate) {
	go func() {
		_, err := ds.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{Content: discord.MessageInternalError})
		if err != nil {
			s.logger.Errorw("sending message",
				"channel", m.ChannelID,
				"msg", discord.MessageInternalError,
				"err", err)
		}
	}()
}
