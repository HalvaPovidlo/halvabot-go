package discord

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"

	"github.com/HalvaPovidlo/discordBotGo/internal/audio"
	"github.com/HalvaPovidlo/discordBotGo/internal/music/player"
	"github.com/HalvaPovidlo/discordBotGo/internal/pkg"
	"github.com/HalvaPovidlo/discordBotGo/pkg/discord/command"
	"github.com/HalvaPovidlo/discordBotGo/pkg/util"
	"github.com/HalvaPovidlo/discordBotGo/pkg/zap"
)

const (
	play       = "play "
	skip       = "skip"
	disconnect = "disconnect"
	hello      = "hello"
)

type Player interface {
	Play(s *pkg.Song) (int, error)
	Skip()
	SetLoop(b bool)
	LoopStatus() bool
	NowPlaying() pkg.Song
	Stats() audio.SessionStats
	Connect(guildID, channelID string)
	Disconnect()
	SubscribeOnErrors(h player.ErrorHandler)
	// Enqueue(s *pkg.SongRequest)
	// Stop()
	// Radio()
}

type YouTube interface {
	FindSong(query string) (*pkg.Song, error)
}

type Service struct {
	player  Player
	youtube YouTube
	prefix  string
	logger  *zap.Logger
}

func NewCog(player Player, youtube YouTube, prefix string, logger *zap.Logger) *Service {
	s := Service{
		player:  player,
		youtube: youtube,
		prefix:  prefix,
		logger:  logger,
	}
	s.player.SubscribeOnErrors(&s)
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

func (s *Service) RegisterCommands(session *discordgo.Session, debug bool) {
	registerSlashBasicCommand(session, debug)
	command.NewMessageCommand(s.prefix+play, s.playMessageHandler, debug).RegisterCommand(session)
	command.NewMessageCommand(s.prefix+skip, s.skipMessageHandler, debug).RegisterCommand(session)
	command.NewMessageCommand(s.prefix+disconnect, s.disconnectMessageHandler, debug).RegisterCommand(session)
	command.NewMessageCommand(s.prefix+hello, s.helloMessageHandler, debug).RegisterCommand(session)
}

func (s *Service) helloMessageHandler(session *discordgo.Session, m *discordgo.MessageCreate) {
	_, _ = session.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Hello, %s %s!", m.Author.Token, m.Author.Username))
}

func (s *Service) playMessageHandler(session *discordgo.Session, m *discordgo.MessageCreate) {
	s.logger.Debug("$play command Handled")
	query := strings.TrimPrefix(m.Content, s.prefix+play)
	query = util.StandardizeSpaces(query)

	s.logger.Debug("finding author's voice channel ID")
	id, err := findAuthorVoiceChannelID(session, m)
	if err != nil {
		s.logger.Error(err, "failed to find author's voice channel")
		return
	}
	sp := sendParams{S: session, M: m, L: s.logger}
	sendSearchingMessage(sp)
	s.logger.Debug("Finding song")
	song, err := s.youtube.FindSong(query)
	if err != nil {
		s.logger.Errorw("find song", "err", err, "query", query)
		return
	}

	s.logger.Debug("connecting")
	s.player.Connect(m.GuildID, id)
	playbacks, err := s.player.Play(song)
	if err != nil {
		s.logger.Error(errors.Wrap(err, "play song"))
	}
	sendFoundMessage(song.ArtistName, song.Title, playbacks, sp)
}

func (s *Service) skipMessageHandler(session *discordgo.Session, m *discordgo.MessageCreate) {
	s.logger.Debug("$skip command Handled")
	s.player.Skip()
}

func (s *Service) disconnectMessageHandler(session *discordgo.Session, m *discordgo.MessageCreate) {
	s.logger.Debug("$disconnect command Handled")
	s.player.Skip()
}

func (s *Service) HandleError(err error) {
	s.logger.Error(err)
}

// func updateListeningStatus(session *discordgo.Session, name string, title string) {
//		err := session.UpdateListeningStatus(name + " - " + title)
//		if err != nil {
//			return
//		}
//	}

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
