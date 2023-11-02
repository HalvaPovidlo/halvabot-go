package music

import (
	"context"
	"fmt"
	"github.com/diamondburned/arikawa/v3/api"
	ads "github.com/diamondburned/arikawa/v3/discord"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/HalvaPovidlo/halvabot-go/internal/pkg/item"
	"github.com/HalvaPovidlo/halvabot-go/pkg/contexts"
)

const (
	messageSearching       = ":trumpet: **Searching** :mag_right:"
	messageSkip            = ":fast_forward: **Skipped** :thumbsup:"
	messageFound           = "**Song found** :notes:"
	messageNotFound        = ":x: **Song not found**"
	messageAgeRestriction  = ":underage: **Song is blocked**"
	messageLoopEnabled     = ":white_check_mark: **Loop enabled**"
	messageLoopDisabled    = ":x: **Loop disabled**"
	messageRadioEnabled    = ":white_check_mark: **Radio enabled**"
	messageRadioDisabled   = ":x: **Radio disabled**"
	messageNotVoiceChannel = ":x: **You have to be in a voice channel to use this command**"
)

const (
	statusLevel = iota
	infoLevel
)

func (h *DiscordHandler) sendComplexMessage(ctx context.Context, channelID ads.ChannelID, msg api.SendMessageData, level int) {
	if h.toDelete(channelID, level) {
		return
	}
	go func() {
		_, err := h.state.SendMessageComplex(channelID, msg)
		if err != nil {
			contexts.GetLogger(ctx).Error("sending message",
				zap.String("channel", channelID.String()),
				zap.String("msg", msg.Content),
				zap.Error(err))
		}
	}()
}

func (h *DiscordHandler) sendSearchingMessage(ctx context.Context, channelID ads.ChannelID) {
	h.sendComplexMessage(ctx, channelID, strmsg(messageSearching), statusLevel)
}

func (h *DiscordHandler) sendSkipMessage(ctx context.Context, channelID ads.ChannelID) {
	h.sendComplexMessage(ctx, channelID, strmsg(messageSkip), statusLevel)
}

func (h *DiscordHandler) sendFoundMessage(ctx context.Context, channelID ads.ChannelID, artist, title string, playbacks int) {
	msg := fmt.Sprintf("%s `%s - %s` %s", messageFound, artist, title, intToEmoji(playbacks))
	h.sendComplexMessage(ctx, channelID, strmsg(msg), statusLevel)
}

func (h *DiscordHandler) sendNotFoundMessage(ctx context.Context, channelID ads.ChannelID) {
	h.sendComplexMessage(ctx, channelID, strmsg(messageNotFound), statusLevel)
}

func (h *DiscordHandler) sendAgeRestrictionMessage(ctx context.Context, channelID ads.ChannelID) {
	h.sendComplexMessage(ctx, channelID, strmsg(messageAgeRestriction), statusLevel)
}

func (h *DiscordHandler) sendLoopMessage(ctx context.Context, channelID ads.ChannelID, enabled bool) {
	if enabled {
		h.sendComplexMessage(ctx, channelID, strmsg(messageLoopEnabled), statusLevel)
	} else {
		h.sendComplexMessage(ctx, channelID, strmsg(messageLoopDisabled), statusLevel)
	}
}

func (h *DiscordHandler) sendRadioMessage(ctx context.Context, channelID ads.ChannelID, enabled bool) {
	if enabled {
		h.sendComplexMessage(ctx, channelID, strmsg(messageRadioEnabled), statusLevel)
	} else {
		h.sendComplexMessage(ctx, channelID, strmsg(messageRadioDisabled), statusLevel)
	}
}

func (h *DiscordHandler) sendNotInVoiceWarning(ctx context.Context, channelID ads.ChannelID) {
	h.sendComplexMessage(ctx, channelID, strmsg(messageNotVoiceChannel), statusLevel)
}

func (h *DiscordHandler) sendNowPlayingMessage(ctx context.Context, channelID ads.ChannelID, song *item.Song, pos float64) {
	msg := api.SendMessageData{
		Embeds: []ads.Embed{
			{
				URL:         song.URL,
				Type:        ads.ImageEmbed,
				Title:       song.Title,
				Description: "",
				Color:       0,
				Image: &ads.EmbedImage{
					URL: song.ArtworkURL,
				},
				Video:    nil,
				Provider: nil,
				Author: &ads.EmbedAuthor{
					Name: song.ArtistName,
					URL:  song.ArtistURL,
				},
				Fields: []ads.EmbedField{
					{
						Name:   "Duration",
						Value:  (time.Duration(song.Duration) * time.Second).String(),
						Inline: true,
					},
					{
						Name:   "Estimated time",
						Value:  (time.Duration(song.Duration-pos) * time.Second).String(),
						Inline: true,
					},
				},
			},
		},
	}
	h.sendComplexMessage(ctx, channelID, msg, infoLevel)
}

func (h *DiscordHandler) sendRandomMessage(ctx context.Context, channelID ads.ChannelID, songs []*item.Song) {
	msg := ""
	for _, song := range songs {
		// TODO: change "$" to constant
		if song.ArtistName != "" {
			msg += fmt.Sprintf("`%s%s - %s`\n", "$"+play, song.ArtistName, song.Title)
		} else {
			msg += fmt.Sprintf("`%s%s`\n", "$"+play, song.Title)
		}
	}
	h.sendComplexMessage(ctx, channelID, strmsg(msg), infoLevel)
}

func (h *DiscordHandler) toDelete(channelID ads.ChannelID, level int) bool {
	h.channelsMx.RLock()
	name := h.allChannels[channelID]
	_, status := h.statusChannels[name]
	_, open := h.openChannels[name]
	h.channelsMx.RUnlock()
	if level <= infoLevel && !(open || status) {
		return true
	}
	if level <= statusLevel && !status {
		return true
	}
	return false
}

func (h *DiscordHandler) loadChannelsID(guildID ads.GuildID) {
	h.channelsMx.Lock()
	defer h.channelsMx.Unlock()
	if len(h.allChannels) != 0 {
		return
	}

	channels, err := h.state.Channels(guildID)
	if err != nil {
		return
	}
	for _, v := range channels {
		h.allChannels[v.ID] = v.Name
	}
}

func intToEmoji(n int) string {
	if n == 0 {
		return ""
	}
	number := strconv.Itoa(n)
	res := ""
	for i := range number {
		res += digitAsEmoji(string(number[i]))
	}
	return res
}

func strmsg(msg string) api.SendMessageData {
	return api.SendMessageData{Content: msg}
}

func digitAsEmoji(digit string) string {
	switch digit {
	case "1":
		return "1️⃣"
	case "2":
		return "2️⃣"
	case "3":
		return "3️⃣"
	case "4":
		return "4️⃣"
	case "5":
		return "5️⃣"
	case "6":
		return "6️⃣"
	case "7":
		return "7️⃣"
	case "8":
		return "8️⃣"
	case "9":
		return "9️⃣"
	case "0":
		return "0️⃣"
	}
	return ""
}
