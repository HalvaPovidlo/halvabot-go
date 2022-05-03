package discord

import (
	"fmt"
	"strconv"
	"time"

	dg "github.com/bwmarrin/discordgo"

	"github.com/HalvaPovidlo/discordBotGo/internal/pkg"
)

const (
	monkaS = "<:monkaS:817041877718138891>"
)

const (
	messageSearching       = ":trumpet: **Searching** :mag_right:"
	messageFound           = "**Song found** :notes:"
	messageLoopEnabled     = ":white_check_mark: **Loop enabled**"
	messageLoopDisabled    = ":x: **Loop disabled**"
	messageRadioEnabled    = ":white_check_mark: **Radio enabled**"
	messageRadioDisabled   = ":x: **Radio disabled**"
	messageNotVoiceChannel = ":x: **You have to be in a voice channel to use this command**"
	messageInternalError   = ":x: **Internal error** " + monkaS
)

const (
	statusLevel = iota
	infoLevel
)

func (s *Service) sendComplexMessage(session *dg.Session, channelID string, msg *dg.MessageSend, level int) {
	if s.toDelete(channelID, level) {
		return
	}
	go func() {
		_, err := session.ChannelMessageSendComplex(channelID, msg)
		if err != nil {
			s.logger.Errorw("sending message",
				"channel", channelID,
				"msg", msg,
				"err", err)
		}
	}()
}

func (s *Service) sendSearchingMessage(ds *dg.Session, m *dg.MessageCreate) {
	s.sendComplexMessage(ds, m.ChannelID, strmsg(messageSearching), statusLevel)
}

func (s *Service) sendFoundMessage(ds *dg.Session, m *dg.MessageCreate, artist, title string, playbacks int) {
	msg := fmt.Sprintf("%s `%s - %s` %s", messageFound, artist, title, intToEmoji(playbacks))
	s.sendComplexMessage(ds, m.ChannelID, strmsg(msg), statusLevel)
}

func (s *Service) sendLoopMessage(ds *dg.Session, m *dg.MessageCreate, enabled bool) {
	if enabled {
		s.sendComplexMessage(ds, m.ChannelID, strmsg(messageLoopEnabled), statusLevel)
	} else {
		s.sendComplexMessage(ds, m.ChannelID, strmsg(messageLoopDisabled), statusLevel)
	}
}

func (s *Service) sendRadioMessage(ds *dg.Session, m *dg.MessageCreate, enabled bool) {
	if enabled {
		s.sendComplexMessage(ds, m.ChannelID, strmsg(messageRadioEnabled), statusLevel)
	} else {
		s.sendComplexMessage(ds, m.ChannelID, strmsg(messageRadioDisabled), statusLevel)
	}
}

func (s *Service) sendNotInVoiceWarning(ds *dg.Session, m *dg.MessageCreate) {
	s.sendComplexMessage(ds, m.ChannelID, strmsg(messageNotVoiceChannel), statusLevel)
}

func (s *Service) sendNowPlayingMessage(ds *dg.Session, m *dg.MessageCreate, song *pkg.Song, pos float64) {
	msg := &dg.MessageSend{
		Embeds: []*dg.MessageEmbed{
			{
				URL:         song.URL,
				Type:        dg.EmbedTypeImage,
				Title:       song.Title,
				Description: "",
				Timestamp:   "",
				Color:       0,
				Image: &dg.MessageEmbedImage{
					URL:      song.ArtworkURL,
					ProxyURL: "",
				},
				Video:    nil,
				Provider: nil,
				Author: &dg.MessageEmbedAuthor{
					Name: song.ArtistName,
					URL:  song.ArtistURL,
				},
				Fields: []*dg.MessageEmbedField{
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
	s.sendComplexMessage(ds, m.ChannelID, msg, infoLevel)
}

func (s *Service) sendRandomMessage(ds *dg.Session, m *dg.MessageCreate, songs []*pkg.Song) {
	msg := ""
	for _, song := range songs {
		if song.ArtistName != "" {
			msg += fmt.Sprintf("`%s%s - %s`\n", s.prefix+play, song.ArtistName, song.Title)
		} else {
			msg += fmt.Sprintf("`%s%s`\n", s.prefix+play, song.Title)
		}
	}
	s.sendComplexMessage(ds, m.ChannelID, strmsg(msg), infoLevel)
}

func (s *Service) sendInternalErrorMessage(ds *dg.Session, m *dg.MessageCreate, level int) {
	s.sendComplexMessage(ds, m.ChannelID, strmsg(messageInternalError), level)
}

// func (s *Service) sendStringMessage(ds *dg.Session, m *dg.MessageCreate, msg string, level int) {
//	s.sendComplexMessage(ds, m.ChannelID, strmsg(msg), level)
// }

func (s *Service) toDelete(channelID string, level int) bool {
	s.channelsMx.RLock()
	name := s.allChannels[channelID]
	_, status := s.statusChannels[name]
	_, open := s.openChannels[name]
	s.channelsMx.RUnlock()
	if level <= infoLevel && !(open || status) {
		return true
	}
	if level <= statusLevel && !status {
		return true
	}
	return false
}

func (s *Service) loadChannelsID(ds *dg.Session, guildID string) {
	s.channelsMx.Lock()
	defer s.channelsMx.Unlock()
	if len(s.allChannels) != 0 {
		return
	}

	channels, err := ds.GuildChannels(guildID)
	if err != nil {
		return
	}
	for _, v := range channels {
		s.allChannels[v.ID] = v.Name
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

func strmsg(msg string) *dg.MessageSend {
	return &dg.MessageSend{Content: msg}
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
