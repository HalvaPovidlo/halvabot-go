package discord

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"

	"github.com/HalvaPovidlo/discordBotGo/internal/audio"
	"github.com/HalvaPovidlo/discordBotGo/internal/music/player"
	"github.com/HalvaPovidlo/discordBotGo/internal/pkg"
	"github.com/HalvaPovidlo/discordBotGo/pkg/contexts"
	"github.com/HalvaPovidlo/discordBotGo/pkg/discord/command"
	"github.com/HalvaPovidlo/discordBotGo/pkg/util"
	"github.com/HalvaPovidlo/discordBotGo/pkg/zap"
)

const (
	play       = "play "
	skip       = "skip"
	loop       = "loop"
	nowPlaying = "now"
	random     = "random"
	radio      = "radio"
	disconnect = "disconnect"
	hello      = "hello"
)

type Player interface {
	Play(ctx contexts.Context, query, guildID, channelID string) (*pkg.Song, int, error)
	Skip()
	SetLoop(b bool)
	LoopStatus() bool
	NowPlaying() *pkg.Song
	Stats() audio.SessionStats
	Disconnect() //
	SubscribeOnErrors(h player.ErrorHandler)
	Random(ctx contexts.Context, n int) ([]*pkg.Song, error)
	SetRadio(ctx contexts.Context, b bool, guildID, channelID string) error
	RadioStatus() bool
	// Connect(guildID, channelID string)
	// Enqueue(s *pkg.SongRequest)
	// Stop()
}

type APIConfig struct {
	OpenChannels   []string `json:"open,omitempty"`
	StatusChannels []string `json:"status,omitempty"`
}

type Service struct {
	ctx    contexts.Context
	player Player
	prefix string
	logger zap.Logger

	channelsMx     sync.RWMutex
	allChannels    map[string]string   // id name
	openChannels   map[string]struct{} // name{}
	statusChannels map[string]struct{} // name{}
}

func NewCog(ctx contexts.Context, player Player, prefix string, logger zap.Logger, config APIConfig) *Service {
	s := Service{
		ctx:            ctx,
		player:         player,
		prefix:         prefix,
		logger:         logger,
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

	s.player.SubscribeOnErrors(s.HandleError)
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

func (s *Service) RegisterCommands(session *discordgo.Session, debug bool, logger zap.Logger) {
	registerSlashBasicCommand(session, debug)
	command.NewMessageCommand(s.prefix+play, s.playMessageHandler, debug).RegisterCommand(session, logger)
	command.NewMessageCommand(s.prefix+skip, s.skipMessageHandler, debug).RegisterCommand(session, logger)
	command.NewMessageCommand(s.prefix+loop, s.loopMessageHandler, debug).RegisterCommand(session, logger)
	command.NewMessageCommand(s.prefix+nowPlaying, s.nowpMessageHandler, debug).RegisterCommand(session, logger)
	command.NewMessageCommand(s.prefix+random, s.randomMessageHandler, debug).RegisterCommand(session, logger)
	command.NewMessageCommand(s.prefix+radio, s.radioMessageHandler, debug).RegisterCommand(session, logger)
	command.NewMessageCommand(s.prefix+disconnect, s.disconnectMessageHandler, debug).RegisterCommand(session, logger)
	command.NewMessageCommand(s.prefix+hello, s.helloMessageHandler, debug).RegisterCommand(session, logger)
	s.updateListeningStatus(contexts.Background(), session)
}

func (s *Service) helloMessageHandler(session *discordgo.Session, m *discordgo.MessageCreate) {
	_, _ = session.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Hello, %s %s!", m.Author.Token, m.Author.Username))
}

func (s *Service) playMessageHandler(ds *discordgo.Session, m *discordgo.MessageCreate) {
	s.deleteMessage(ds, m, statusLevel)
	query := strings.TrimPrefix(m.Content, s.prefix+play)
	query = util.StandardizeSpaces(query)

	id, err := findAuthorVoiceChannelID(ds, m)
	if err != nil {
		s.sendNotInVoiceWarning(ds, m)
		s.logger.Error(err, "failed to find author's voice channel")
		return
	}
	s.sendSearchingMessage(ds, m)
	song, playbacks, err := s.player.Play(s.ctx, query, m.GuildID, id)
	if err != nil {
		if pe, ok := err.(*player.Error); ok {
			switch pe {
			case player.ErrSongNotFound:
				s.sendNotFoundMessage(ds, m)
				return
			case player.ErrStorageQueryFailed:
				s.sendInternalErrorMessage(ds, m, statusLevel)
				s.logger.Error(errors.Wrap(err, "database interaction failed"))
			default:
				s.logger.Error(errors.Wrap(err, "play with service"))
				return
			}
		}
	}
	s.sendFoundMessage(ds, m, song.ArtistName, song.Title, playbacks)
}

func (s *Service) skipMessageHandler(session *discordgo.Session, m *discordgo.MessageCreate) {
	s.deleteMessage(session, m, statusLevel)
	s.player.Skip()
}

func (s *Service) loopMessageHandler(session *discordgo.Session, m *discordgo.MessageCreate) {
	s.deleteMessage(session, m, statusLevel)
	b := s.player.LoopStatus()
	s.sendLoopMessage(session, m, !b)
	s.player.SetLoop(!b)
}

func (s *Service) nowpMessageHandler(session *discordgo.Session, m *discordgo.MessageCreate) {
	s.deleteMessage(session, m, infoLevel)
	s.sendNowPlayingMessage(session, m, s.player.NowPlaying(), s.player.Stats().Pos)
}

func (s *Service) randomMessageHandler(session *discordgo.Session, m *discordgo.MessageCreate) {
	s.deleteMessage(session, m, statusLevel)
	songs, err := s.player.Random(s.ctx, 10)
	if err != nil {
		s.logger.Error(errors.Wrap(err, "get random songs"))
		return
	}
	s.sendRandomMessage(session, m, songs)
}

func (s *Service) radioMessageHandler(ds *discordgo.Session, m *discordgo.MessageCreate) {
	s.deleteMessage(ds, m, statusLevel)
	if s.player.RadioStatus() {
		s.sendRadioMessage(ds, m, false)
		_ = s.player.SetRadio(s.ctx, false, "", "")
		return
	}
	id, err := findAuthorVoiceChannelID(ds, m)
	if err != nil {
		s.sendNotInVoiceWarning(ds, m)
		s.logger.Error(err, "failed to find author's voice channel")
		return
	}
	err = s.player.SetRadio(s.ctx, true, m.GuildID, id)
	if err != nil {
		s.sendInternalErrorMessage(ds, m, statusLevel)
		s.logger.Error(errors.Wrap(err, "enable radio"))
	} else {
		s.sendRadioMessage(ds, m, true)
	}
}

func (s *Service) disconnectMessageHandler(session *discordgo.Session, m *discordgo.MessageCreate) {
	s.deleteMessage(session, m, statusLevel)
	s.player.Disconnect()
}

func (s *Service) HandleError(err error) {
	s.logger.Error(errors.Wrap(err, "discord api"))
}

func (s *Service) updateListeningStatus(ctx context.Context, session *discordgo.Session) {
	// TODO: dirty temp code
	// better way to use channels like error chan
	timer := time.NewTicker(5 * time.Second)
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

func (s *Service) deleteMessage(session *discordgo.Session, m *discordgo.MessageCreate, level int) {
	go func() {
		s.loadChannelsID(session, m.GuildID)
		if s.toDelete(m.ChannelID, level) {
			err := session.ChannelMessageDelete(m.ChannelID, m.Message.ID)
			if err != nil {
				s.logger.Error(errors.Wrap(err, "deleting message"))
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
