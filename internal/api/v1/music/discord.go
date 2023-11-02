package music

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/diamondburned/arikawa/v3/api"
	ads "github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/state/store"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/HalvaPovidlo/halvabot-go/internal/music/search/youtube"
	"github.com/HalvaPovidlo/halvabot-go/pkg/contexts"
	"github.com/HalvaPovidlo/halvabot-go/pkg/discord"
	"github.com/HalvaPovidlo/halvabot-go/pkg/util"
)

const (
	play       = "play"
	skip       = "skip"
	skipFS     = "fs"
	loop       = "loop"
	nowPlaying = "now"
	random     = "random"
	radio      = "radio"
	disconnect = "disconnect"
)

type APIConfig struct {
	OpenChannels   []string `json:"open,omitempty"`
	StatusChannels []string `json:"status,omitempty"`
}

type DiscordHandler struct {
	player *Service
	state  *state.State

	channelsMx     sync.RWMutex
	allChannels    map[ads.ChannelID]string // id name
	openChannels   map[string]struct{}      // name{}
	statusChannels map[string]struct{}      // name{}
}

func NewDiscordMusicHandler(player *Service, config APIConfig) *DiscordHandler {
	s := DiscordHandler{
		player:         player,
		allChannels:    make(map[ads.ChannelID]string),
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

	s.updateListeningStatus(context.Background())
	return &s
}

func (h *DiscordHandler) RegisterCommands(ds *discord.Service) {
	ds.RegisterMessageCommand(play, h.playMessageHandler)
	ds.RegisterMessageCommand(skip, h.skipMessageHandler)
	ds.RegisterMessageCommand(skipFS, h.skipMessageHandler)
	ds.RegisterMessageCommand(loop, h.loopMessageHandler)
	ds.RegisterMessageCommand(nowPlaying, h.nowMessageHandler)
	ds.RegisterMessageCommand(random, h.randomMessageHandler)
	ds.RegisterMessageCommand(radio, h.radioMessageHandler)
	ds.RegisterMessageCommand(disconnect, h.disconnectMessageHandler)
}

func (h *DiscordHandler) helloMessageHandler(ctx context.Context, session *discordgo.Session, m *discordgo.MessageCreate) {
	_, _ = session.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Hello, %s %s!", m.Author.Token, m.Author.Username))
}

func (h *DiscordHandler) playMessageHandler(ctx context.Context, m *gateway.MessageCreateEvent) (*api.SendMessageData, error) {
	r, err := h.playHandler(ctx, m.GuildID, m.ChannelID, m.ID, m.Content, m.Member)
	if err != nil {
		return nil, err
	}
	return responseToMessage(r), nil
}

func responseToMessage(r *api.InteractionResponseData) *api.SendMessageData {
	if r == nil {
		return nil
	}
	var embeds []ads.Embed
	if r.Embeds != nil {
		embeds = *r.Embeds
	}
	var components ads.ContainerComponents
	if r.Components != nil {
		components = *r.Components
	}
	return &api.SendMessageData{
		Content:         r.Content.Val,
		TTS:             r.TTS,
		Embeds:          embeds,
		Files:           r.Files,
		Components:      components,
		AllowedMentions: r.AllowedMentions,
	}
}

func (h *DiscordHandler) playHandler(
	ctx context.Context,
	guildID ads.GuildID, channelID ads.ChannelID, messageID ads.MessageID, content string,
	member *ads.Member) (*api.InteractionResponseData, error) {
	h.deleteMessage(ctx, guildID, channelID, messageID, statusLevel)
	query := util.StandardizeSpaces(content)

	author := member.User
	voiceChannelID, err := h.findAuthorVoiceChannelID(guildID, author.ID)
	logger := contexts.GetLogger(ctx)
	if err != nil {
		if !errors.Is(store.ErrNotFound, err) {
			return nil, err
		}
		logger.Warn("failed to find author's voice channel", zap.Error(err))
		h.sendNotInVoiceWarning(ctx, channelID)
		return nil, nil
	}

	h.sendSearchingMessage(ctx, channelID)
	song, err := h.player.Play(ctx, query, author.ID, guildID, voiceChannelID)
	switch {
	case errors.Is(err, youtube.ErrSongNotFound):
		h.sendNotFoundMessage(ctx, channelID)
		return nil, nil
	case strings.Contains(err.Error(), "can't bypass age restriction"):
		h.sendAgeRestrictionMessage(ctx, channelID)
		return nil, nil
	case err != nil:
		return nil, errors.Wrap(err, "player play song")
	}

	h.sendFoundMessage(ctx, channelID, song.ArtistName, song.Title, song.Playbacks)
	return nil, nil
}

func (h *DiscordHandler) skipMessageHandler(ctx context.Context, m *gateway.MessageCreateEvent) (*api.SendMessageData, error) {
	h.deleteMessage(ctx, m.GuildID, m.ChannelID, m.ID, statusLevel)
	h.player.Skip(ctx)
	h.sendSkipMessage(ctx, m.ChannelID)
}

func (h *DiscordHandler) loopMessageHandler(ctx context.Context, m *gateway.MessageCreateEvent) (*api.SendMessageData, error) {
	h.deleteMessage(ctx, m.GuildID, m.ChannelID, m.ID, statusLevel)
	b := h.player.LoopStatus()
	h.sendLoopMessage(ctx, m.ChannelID, !b)
	h.player.SetLoop(ctx, !b)
}

func (h *DiscordHandler) nowMessageHandler(ctx context.Context, m *gateway.MessageCreateEvent) (*api.SendMessageData, error) {
	h.deleteMessage(ctx, m.GuildID, m.ChannelID, m.ID, infoLevel)
	h.sendNowPlayingMessage(ctx, m.ChannelID, h.player.NowPlaying(), h.player.SongStatus().Pos)
}

func (h *DiscordHandler) randomMessageHandler(ctx context.Context, m *gateway.MessageCreateEvent) (*api.SendMessageData, error) {
	h.deleteMessage(ctx, m.GuildID, m.ChannelID, m.ID, statusLevel)
	songs, err := h.player.Random(ctx, 10)
	if err != nil {
		return nil, errors.Wrap(err, "random song")
	}
	h.sendRandomMessage(ctx, m.ChannelID, songs)
}

func (h *DiscordHandler) radioMessageHandler(ctx context.Context, m *gateway.MessageCreateEvent) (*api.SendMessageData, error) {
	h.deleteMessage(ctx, m.GuildID, m.ChannelID, m.ID, statusLevel)
	if h.player.RadioStatus() {
		h.sendRadioMessage(ctx, m.ChannelID, false)
		_ = h.player.SetRadio(ctx, false, "", "")
		return nil, nil
	}

	id, err := h.findAuthorVoiceChannelID(m.GuildID, m.Member.User.ID)
	logger := contexts.GetLogger(ctx)
	if err != nil {
		if !errors.Is(store.ErrNotFound, err) {
			return nil, err
		}
		logger.Warn("failed to find author's voice channel", zap.Error(err))
		h.sendNotInVoiceWarning(ctx, m.ChannelID)
		return nil, nil
	}

	err = h.player.SetRadio(ctx, true, m.GuildID, id)
	if err != nil {
		return nil, errors.Wrap(err, "enable radio")
	} else {
		h.sendRadioMessage(ctx, m.ChannelID, true)
	}
	return nil, nil
}

func (h *DiscordHandler) disconnectMessageHandler(ctx context.Context, m *gateway.MessageCreateEvent) (*api.SendMessageData, error) {
	h.deleteMessage(ctx, m.GuildID, m.ChannelID, m.ID, statusLevel)
	h.player.Disconnect(ctx)
	return nil, nil
}

func (h *DiscordHandler) updateListeningStatus(ctx context.Context) {
	// TODO: dirty temp code
	// better way to use channels like error chan
	//timer := time.NewTicker(1 * time.Second)
	//go func() {
	//	defer timer.Stop()
	//	for {
	//		select {
	//		case <-timer.C:
	//			song := h.player.NowPlaying()
	//			title := ""
	//			if song != nil {
	//				title = song.Title
	//			}
	//			me, _ := h.state.Me()
	//			ctx.Gat
	//			_ = h.state.Upda().SetState
	//		case <-ctx.Done():
	//			return
	//		}
	//	}
	//}()
}

func (h *DiscordHandler) deleteMessage(ctx context.Context, guildID ads.GuildID, channelID ads.ChannelID, messageID ads.MessageID, level int) {
	go func() {
		h.loadChannelsID(guildID)
		if h.toDelete(channelID, level) {
			err := h.state.DeleteMessage(channelID, messageID, "wrong message channel")
			if err != nil {
				contexts.GetLogger(ctx).Error("deleting message", zap.Error(err))
			}
		}
	}()
}

func (h *DiscordHandler) findAuthorVoiceChannelID(guildID ads.GuildID, userID ads.UserID) (ads.ChannelID, error) {
	voiceState, err := h.state.VoiceState(guildID, userID)
	if err != nil {
		return ads.NullChannelID, errors.Wrap(err, "find user voice state")
	}
	return voiceState.ChannelID, nil
}
