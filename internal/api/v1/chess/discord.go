package chess

import (
	"context"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/api/cmdroute"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/pkg/errors"

	"github.com/HalvaPovidlo/halvabot-go/internal/chess/lichess"
	"github.com/HalvaPovidlo/halvabot-go/pkg/discord"
)

const chess = "chess"

type chessClient interface {
	StartOpenGame(ctx context.Context) (*lichess.OpenGameResponse, error)
}

type Service struct {
	client chessClient
	prefix string
}

func NewDiscordChessHandler(prefix string, client chessClient) *Service {
	s := Service{
		prefix: prefix,
		client: client,
	}

	return &s
}

func (s *Service) RegisterCommands(ds *discord.Service) {
	ds.RegisterCommand(api.CreateCommandData{
		Name:        chess,
		Description: "Start chess game",
	}, s.chessCommandHandler)
	ds.RegisterMessageCommand(chess, s.chessMessageHandler)
}

func (s *Service) chessMessageHandler(ctx context.Context, c *gateway.MessageCreateEvent) (*api.SendMessageData, error) {
	game, err := s.client.StartOpenGame(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "start chess game")
	}
	return &api.SendMessageData{Content: game.Challenge.URL}, nil
}

func (s *Service) chessCommandHandler(ctx context.Context, data cmdroute.CommandData) (*api.InteractionResponseData, error) {
	game, err := s.client.StartOpenGame(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "start chess game")
	}
	return &api.InteractionResponseData{Content: option.NewNullableString(game.Challenge.URL)}, nil
}
