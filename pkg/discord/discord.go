package discord

import (
	"context"
	"github.com/HalvaPovidlo/halvabot-go/pkg/contexts"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/api/cmdroute"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/diamondburned/arikawa/v3/voice"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"strings"
	"time"
)

const (
	// debugChannelID TODO: load from session and guildID
	debugChannelID = "747743319039148062"

	MonkaS               = "<:monkaS:817041877718138891>"
	messageInternalError = ":x: **Internal error** " + MonkaS
)

type MessageHandlerFunc func(ctx context.Context, c *gateway.MessageCreateEvent) (*api.SendMessageData, error)
type CommandHandlerFunc func(ctx context.Context, data cmdroute.CommandData) (*api.InteractionResponseData, error)

type Service struct {
	id                   discord.UserID
	commands             []api.CreateCommandData
	messageCommandPrefix string
	debug                bool
	debugChannel         discord.ChannelID
	logger               *zap.Logger
	r                    *cmdroute.Router
	*state.State
}

func New(token, messageCommandPrefix string, debug bool, logger *zap.Logger) *Service {
	debugChannelID, _ := discord.ParseSnowflake(debugChannelID)
	s := &Service{
		debug:                debug,
		logger:               logger,
		messageCommandPrefix: messageCommandPrefix,
		debugChannel:         discord.ChannelID(debugChannelID),
		r:                    cmdroute.NewRouter(),
		State:                state.New("Bot " + token),
	}

	voice.AddIntents(s)
	s.AddIntents(gateway.IntentGuilds)
	s.AddIntents(gateway.IntentGuildMessages)

	s.AddHandler(func(*gateway.ReadyEvent) {
		me, _ := s.Me()
		s.id = me.ID
	})

	return s
}

func (s *Service) Connect(ctx context.Context) error {
	s.State = s.State.WithContext(ctx)
	if err := cmdroute.OverwriteCommands(s, s.commands); err != nil {
		return errors.Wrap(err, "update commands")
	}
	return errors.Wrap(s.State.Connect(ctx), "connect session")
}

func (s *Service) RegisterCommand(cmd api.CreateCommandData, cmdFunc CommandHandlerFunc) {
	s.commands = append(s.commands, cmd)
	s.r.AddFunc(cmd.Name, func(ctx context.Context, data cmdroute.CommandData) *api.InteractionResponseData {
		ctx = contexts.WithCommandValues(context.Background(), cmd.Name, s.logger, "")
		log := contexts.GetLogger(ctx)

		start := time.Now()
		log.Info("command handled")
		response, err := cmdFunc(ctx, data)
		if err != nil {
			log.Error("command execution failed", zap.Error(err), zap.Duration("elapsed", time.Since(start)))
			return &api.InteractionResponseData{
				Content:         option.NewNullableString(messageInternalError),
				Flags:           discord.EphemeralMessage,
				AllowedMentions: &api.AllowedMentions{},
			}
		}
		log.Info("command executed", zap.Duration("elapsed", time.Since(start)))
		return response
	})
}

func (s *Service) RegisterMessageCommand(name string, handle MessageHandlerFunc) {
	s.AddHandler(func(c *gateway.MessageCreateEvent) {
		defer func() {
			if e := recover(); e != nil {
				s.logger.Error("panic during message handling", zap.Any("error", e), zap.Stack("stack"))
			}
		}()

		if s.skip(c) {
			return
		}

		command := s.messageCommandPrefix + name
		if command == strings.ToLower(c.Content[0:len(command)]) {
			ctx := contexts.WithCommandValues(context.Background(), name, s.logger, "")
			log := contexts.GetLogger(ctx)

			start := time.Now()
			log.Info("command handled", zap.String("query", c.Content))
			c.Content = c.Content[len(command)-1:]
			response, err := handle(ctx, c)
			if err != nil {
				log.Error("command execution failed", zap.Error(err), zap.Duration("elapsed", time.Since(start)))
				if _, err := s.SendMessage(c.ChannelID, messageInternalError); err != nil {
					log.Error("send message failed", zap.Error(err))
				}
			} else {
				if response != nil {
					if _, err := s.SendMessageComplex(c.ChannelID, *response); err != nil {
						log.Error("send message failed", zap.Error(err))
					}
				}
				log.Info("command executed", zap.Duration("elapsed", time.Since(start)))
			}
		}
	})
}

func (s *Service) skip(c *gateway.MessageCreateEvent) bool {
	switch {
	case c.Member.User.ID == s.id:
		return true
	case (c.ChannelID == s.debugChannel) != s.debug:
		return true
	}
	return false
}
