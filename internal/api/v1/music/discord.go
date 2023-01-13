package music

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/HalvaPovidlo/halvabot-go/internal/music/search/youtube"
	"github.com/HalvaPovidlo/halvabot-go/pkg/contexts"
	"github.com/HalvaPovidlo/halvabot-go/pkg/discord/command"
	"github.com/HalvaPovidlo/halvabot-go/pkg/util"
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

type APIConfig struct {
	OpenChannels   []string `json:"open,omitempty"`
	StatusChannels []string `json:"status,omitempty"`
}

type DiscordHandler struct {
	player *Service
	prefix string

	channelsMx     sync.RWMutex
	allChannels    map[string]string   // id name
	openChannels   map[string]struct{} // name{}
	statusChannels map[string]struct{} // name{}
}

func NewDiscordMusicHandler(player *Service, prefix string, config APIConfig) *DiscordHandler {
	s := DiscordHandler{
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

func (h *DiscordHandler) RegisterCommands(ctx context.Context, session *discordgo.Session, debug bool, logger *zap.Logger) {
	registerSlashBasicCommand(session, debug)
	command.NewMessageCommand(h.prefix+play, h.playMessageHandler, debug).RegisterCommand(session, logger)
	command.NewMessageCommand(h.prefix+skip, h.skipMessageHandler, debug).RegisterCommand(session, logger)
	command.NewMessageCommand(h.prefix+skipFS, h.skipMessageHandler, debug).RegisterCommand(session, logger)
	command.NewMessageCommand(h.prefix+loop, h.loopMessageHandler, debug).RegisterCommand(session, logger)
	command.NewMessageCommand(h.prefix+nowPlaying, h.nowMessageHandler, debug).RegisterCommand(session, logger)
	command.NewMessageCommand(h.prefix+random, h.randomMessageHandler, debug).RegisterCommand(session, logger)
	command.NewMessageCommand(h.prefix+radio, h.radioMessageHandler, debug).RegisterCommand(session, logger)
	command.NewMessageCommand(h.prefix+disconnect, h.disconnectMessageHandler, debug).RegisterCommand(session, logger)
	command.NewMessageCommand(h.prefix+hello, h.helloMessageHandler, debug).RegisterCommand(session, logger)
	h.updateListeningStatus(ctx, session)
}

func (h *DiscordHandler) helloMessageHandler(ctx context.Context, session *discordgo.Session, m *discordgo.MessageCreate) {
	_, _ = session.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Hello, %s %s!", m.Author.Token, m.Author.Username))
}

func (h *DiscordHandler) playMessageHandler(ctx context.Context, ds *discordgo.Session, m *discordgo.MessageCreate) {
	h.deleteMessage(ctx, ds, m, statusLevel)
	query := strings.TrimPrefix(m.Content, h.prefix+play)
	query = util.StandardizeSpaces(query)

	id, err := findAuthorVoiceChannelID(ds, m)
	logger := contexts.GetLogger(ctx)
	if err != nil {
		h.sendNotInVoiceWarning(ctx, ds, m)
		logger.Error("failed to find author's voice channel", zap.Error(err))
		return
	}
	h.sendSearchingMessage(ctx, ds, m)
	song, err := h.player.Play(ctx, query, m.Author.ID, m.GuildID, id)
	if err != nil {
		if errors.Is(err, youtube.ErrSongNotFound) {
			h.sendNotFoundMessage(ctx, ds, m)
			return
		}
		if strings.Contains(err.Error(), "can't bypass age restriction") {
			h.sendAgeRestrictionMessage(ctx, ds, m)
			return
		}
		logger.Error("player play song", zap.String("query", query), zap.Error(err))
		h.sendInternalErrorMessage(ctx, ds, m, statusLevel)
		return
	}
	h.sendFoundMessage(ctx, ds, m, song.ArtistName, song.Title, song.Playbacks)
}

func (h *DiscordHandler) skipMessageHandler(ctx context.Context, session *discordgo.Session, m *discordgo.MessageCreate) {
	h.deleteMessage(ctx, session, m, statusLevel)
	h.player.Skip(ctx)
	h.sendSkipMessage(ctx, session, m)
}

func (h *DiscordHandler) loopMessageHandler(ctx context.Context, session *discordgo.Session, m *discordgo.MessageCreate) {
	h.deleteMessage(ctx, session, m, statusLevel)
	b := h.player.LoopStatus()
	h.sendLoopMessage(ctx, session, m, !b)
	h.player.SetLoop(ctx, !b)
}

func (h *DiscordHandler) nowMessageHandler(ctx context.Context, session *discordgo.Session, m *discordgo.MessageCreate) {
	h.deleteMessage(ctx, session, m, infoLevel)
	h.sendNowPlayingMessage(ctx, session, m, h.player.NowPlaying(), h.player.SongStatus().Pos)
}

func (h *DiscordHandler) randomMessageHandler(ctx context.Context, session *discordgo.Session, m *discordgo.MessageCreate) {
	h.deleteMessage(ctx, session, m, statusLevel)
	songs, err := h.player.Random(ctx, 10)
	if err != nil {
		contexts.GetLogger(ctx).Error("get random songs", zap.Error(err))
		h.sendInternalErrorMessage(ctx, session, m, infoLevel)
		return
	}
	h.sendRandomMessage(ctx, session, m, songs)
}

func (h *DiscordHandler) radioMessageHandler(ctx context.Context, ds *discordgo.Session, m *discordgo.MessageCreate) {
	h.deleteMessage(ctx, ds, m, statusLevel)
	if h.player.RadioStatus() {
		h.sendRadioMessage(ctx, ds, m, false)
		_ = h.player.SetRadio(ctx, false, "", "")
		return
	}
	id, err := findAuthorVoiceChannelID(ds, m)
	logger := contexts.GetLogger(ctx)
	if err != nil {
		h.sendNotInVoiceWarning(ctx, ds, m)
		logger.Error("failed to find author's voice channel", zap.Error(err))
		return
	}
	err = h.player.SetRadio(ctx, true, m.GuildID, id)
	if err != nil {
		h.sendInternalErrorMessage(ctx, ds, m, statusLevel)
		logger.Error("enable radio", zap.Error(err))
	} else {
		h.sendRadioMessage(ctx, ds, m, true)
	}
}

func (h *DiscordHandler) disconnectMessageHandler(ctx context.Context, session *discordgo.Session, m *discordgo.MessageCreate) {
	h.deleteMessage(ctx, session, m, statusLevel)
	h.player.Disconnect(ctx)
}

func (h *DiscordHandler) updateListeningStatus(ctx context.Context, session *discordgo.Session) {
	// TODO: dirty temp code
	// better way to use channels like error chan
	timer := time.NewTicker(1 * time.Second)
	go func() {
		defer timer.Stop()
		for {
			select {
			case <-timer.C:
				song := h.player.NowPlaying()
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

func (h *DiscordHandler) deleteMessage(ctx context.Context, session *discordgo.Session, m *discordgo.MessageCreate, level int) {
	go func() {
		h.loadChannelsID(session, m.GuildID)
		if h.toDelete(m.ChannelID, level) {
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
