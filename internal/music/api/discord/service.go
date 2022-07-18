package discord

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/HalvaPovidlo/discordBotGo/internal/music/search/youtube"
	"github.com/HalvaPovidlo/discordBotGo/internal/pkg"
	"github.com/HalvaPovidlo/discordBotGo/pkg/contexts"
	"github.com/HalvaPovidlo/discordBotGo/pkg/discord/command"
	"github.com/HalvaPovidlo/discordBotGo/pkg/util"
)

const (
	play       = "play "
	skip       = "skip"
	skipFS     = "fs"
	loop       = "loop"
	nowPlaying = "now"
	random     = "random"
	radio      = "radio"
	disconnect = "disconnect"
	hello      = "hello"
)

type Player interface {
	Play(ctx context.Context, query, userID, guildID, channelID string) (*pkg.Song, int, error)
	Skip(ctx context.Context)
	SetLoop(ctx context.Context, b bool)
	LoopStatus() bool
	NowPlaying() *pkg.Song
	SongStatus() pkg.SessionStats
	Disconnect(ctx context.Context) //
	Random(ctx context.Context, n int) ([]*pkg.Song, error)
	SetRadio(ctx context.Context, b bool, guildID, channelID string) error
	RadioStatus() bool
	// SubscribeOnErrors(h player.ErrorHandler)
	// Connect(guildID, channelID string)
	// Enqueue(s *pkg.SongRequest)
	// Stop()
}

type APIConfig struct {
	OpenChannels   []string `json:"open,omitempty"`
	StatusChannels []string `json:"status,omitempty"`
}

type Service struct {
	player Player
	prefix string

	channelsMx     sync.RWMutex
	allChannels    map[string]string   // id name
	openChannels   map[string]struct{} // name{}
	statusChannels map[string]struct{} // name{}
}

func NewCog(player Player, prefix string, config APIConfig) *Service {
	s := Service{
		player:         player,
		prefix:         prefix,
		allChannels:    make(map[string]string),
		openChannels:   make(map[string]struct{}),
		statusChannels: make(map[string]struct{}),
	}

	s.channelsMx.Lock()
	var t struct{}
	for _, v := range config.OpenChannels {
		s.openChannels[v] = t
	}
	for _, v := range config.StatusChannels {
		s.statusChannels[v] = t
	}
	s.channelsMx.Unlock()

	return &s
}

func registerSlashBasicCommand(s *discordgo.Session, debug bool) (unregisterCommand func()) {
	sc := command.NewSlashCommand(
		&discordgo.ApplicationCommand{Name: "basic2-command", Description: "Basic command"},
		func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Hey there! Congratulations, you just executed your first slash command",
				},
			})
		}, debug,
	)
	return sc.RegisterCommand(s)
}

func (s *Service) RegisterCommands(ctx context.Context, session *discordgo.Session, debug bool, logger *zap.Logger) {
	registerSlashBasicCommand(session, debug)
	command.NewMessageCommand(s.prefix+play, s.playMessageHandler, debug).RegisterCommand(session, logger)
	command.NewMessageCommand(s.prefix+skip, s.skipMessageHandler, debug).RegisterCommand(session, logger)
	command.NewMessageCommand(s.prefix+skipFS, s.skipMessageHandler, debug).RegisterCommand(session, logger)
	command.NewMessageCommand(s.prefix+loop, s.loopMessageHandler, debug).RegisterCommand(session, logger)
	command.NewMessageCommand(s.prefix+nowPlaying, s.nowpMessageHandler, debug).RegisterCommand(session, logger)
	command.NewMessageCommand(s.prefix+random, s.randomMessageHandler, debug).RegisterCommand(session, logger)
	command.NewMessageCommand(s.prefix+radio, s.radioMessageHandler, debug).RegisterCommand(session, logger)
	command.NewMessageCommand(s.prefix+disconnect, s.disconnectMessageHandler, debug).RegisterCommand(session, logger)
	command.NewMessageCommand(s.prefix+hello, s.helloMessageHandler, debug).RegisterCommand(session, logger)
	s.updateListeningStatus(ctx, session)
}

func (s *Service) helloMessageHandler(ctx context.Context, session *discordgo.Session, m *discordgo.MessageCreate) {
	_, _ = session.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Hello, %s %s!", m.Author.Token, m.Author.Username))
}

func (s *Service) playMessageHandler(ctx context.Context, ds *discordgo.Session, m *discordgo.MessageCreate) {
	s.deleteMessage(ctx, ds, m, statusLevel)
	query := strings.TrimPrefix(m.Content, s.prefix+play)
	query = util.StandardizeSpaces(query)

	id, err := findAuthorVoiceChannelID(ds, m)
	logger := contexts.GetLogger(ctx)
	if err != nil {
		s.sendNotInVoiceWarning(ctx, ds, m)
		logger.Error("failed to find author's voice channel", zap.Error(err))
		return
	}
	s.sendSearchingMessage(ctx, ds, m)
	song, playbacks, err := s.player.Play(ctx, query, m.Author.ID, m.GuildID, id)
	if err != nil {
		if errors.Is(err, youtube.ErrSongNotFound) {
			s.sendNotFoundMessage(ctx, ds, m)
			return
		}
		if strings.Contains(err.Error(), "can't bypass age restriction") {
			s.sendAgeRestrictionMessage(ctx, ds, m)
			return
		}
		logger.Error("player play song", zap.String("query", query), zap.Error(err))
		s.sendInternalErrorMessage(ctx, ds, m, statusLevel)
		return
	}
	s.sendFoundMessage(ctx, ds, m, song.ArtistName, song.Title, playbacks)
}

func (s *Service) skipMessageHandler(ctx context.Context, session *discordgo.Session, m *discordgo.MessageCreate) {
	s.deleteMessage(ctx, session, m, statusLevel)
	s.player.Skip(ctx)
	s.sendSkipMessage(ctx, session, m)
}

func (s *Service) loopMessageHandler(ctx context.Context, session *discordgo.Session, m *discordgo.MessageCreate) {
	s.deleteMessage(ctx, session, m, statusLevel)
	b := s.player.LoopStatus()
	s.sendLoopMessage(ctx, session, m, !b)
	s.player.SetLoop(ctx, !b)
}

func (s *Service) nowpMessageHandler(ctx context.Context, session *discordgo.Session, m *discordgo.MessageCreate) {
	s.deleteMessage(ctx, session, m, infoLevel)
	s.sendNowPlayingMessage(ctx, session, m, s.player.NowPlaying(), s.player.SongStatus().Pos)
}

func (s *Service) randomMessageHandler(ctx context.Context, session *discordgo.Session, m *discordgo.MessageCreate) {
	s.deleteMessage(ctx, session, m, statusLevel)
	songs, err := s.player.Random(ctx, 10)
	if err != nil {
		contexts.GetLogger(ctx).Error("get random songs", zap.Error(err))
		s.sendInternalErrorMessage(ctx, session, m, infoLevel)
		return
	}
	s.sendRandomMessage(ctx, session, m, songs)
}

func (s *Service) radioMessageHandler(ctx context.Context, ds *discordgo.Session, m *discordgo.MessageCreate) {
	s.deleteMessage(ctx, ds, m, statusLevel)
	if s.player.RadioStatus() {
		s.sendRadioMessage(ctx, ds, m, false)
		_ = s.player.SetRadio(ctx, false, "", "")
		return
	}
	id, err := findAuthorVoiceChannelID(ds, m)
	logger := contexts.GetLogger(ctx)
	if err != nil {
		s.sendNotInVoiceWarning(ctx, ds, m)
		logger.Error("failed to find author's voice channel", zap.Error(err))
		return
	}
	err = s.player.SetRadio(ctx, true, m.GuildID, id)
	if err != nil {
		s.sendInternalErrorMessage(ctx, ds, m, statusLevel)
		logger.Error("enable radio", zap.Error(err))
	} else {
		s.sendRadioMessage(ctx, ds, m, true)
	}
}

func (s *Service) disconnectMessageHandler(ctx context.Context, session *discordgo.Session, m *discordgo.MessageCreate) {
	s.deleteMessage(ctx, session, m, statusLevel)
	s.player.Disconnect(ctx)
}

func (s *Service) updateListeningStatus(ctx context.Context, session *discordgo.Session) {
	// TODO: dirty temp code
	// better way to use channels like error chan
	timer := time.NewTicker(1 * time.Second)
	go func() {
		defer timer.Stop()
		for {
			select {
			case <-timer.C:
				song := s.player.NowPlaying()
				title := ""
				if song != nil {
					title = song.Title
				}
				_ = session.UpdateListeningStatus(title)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (s *Service) deleteMessage(ctx context.Context, session *discordgo.Session, m *discordgo.MessageCreate, level int) {
	go func() {
		s.loadChannelsID(session, m.GuildID)
		if s.toDelete(m.ChannelID, level) {
			err := session.ChannelMessageDelete(m.ChannelID, m.Message.ID)
			if err != nil {
				contexts.GetLogger(ctx).Error("deleting message", zap.Error(err))
			}
		}
	}()
}

func findAuthorVoiceChannelID(s *discordgo.Session, m *discordgo.MessageCreate) (string, error) {
	guild, err := s.State.Guild(m.GuildID)
	if err != nil {
		return "", err
	}
	id := ""
	for _, voiceState := range guild.VoiceStates {
		if voiceState.UserID == m.Author.ID {
			id = voiceState.ChannelID
			break
		}
	}
	if id == "" {
		return "", errors.New("unable to find user voice channel")
	}

	return id, nil
}
