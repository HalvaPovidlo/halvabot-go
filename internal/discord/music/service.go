package music

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"

	"github.com/HalvaPovidlo/discordBotGo/internal/discord/pkg"
	"github.com/HalvaPovidlo/discordBotGo/pkg/discord/command"
	"github.com/HalvaPovidlo/discordBotGo/pkg/util"
	"github.com/HalvaPovidlo/discordBotGo/pkg/zap"
)

const (
	play = "play "
)

type Player interface {
	PlayYoutube(query string) (*pkg.SongRequest, error)
	Skip() *pkg.SongRequest
	Connect(guildID, channelID string)
	Errors() <-chan error
}

type Service struct {
	player Player
	prefix string
	logger *zap.Logger
}

func NewCog(player Player, prefix string, logger *zap.Logger) *Service {
	c := Service{
		player: player,
		prefix: prefix,
		logger: logger,
	}
	return &c
}

func registerSlashBasicCommand(s *discordgo.Session) (unregisterCommand func()) {
	sc := command.NewSlashCommand(
		&discordgo.ApplicationCommand{Name: "basic2-command", Description: "Basic command"},
		func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Hey there! Congratulations, you just executed your first slash command",
				},
			})
		},
	)
	return sc.RegisterCommand(s)
}

func (s *Service) registerMessagePlayCommand(session *discordgo.Session) {
	mc := command.NewMessageCommand(s.prefix+play, s.playMessageHandler)
	mc.RegisterCommand(session)
}

func (s *Service) RegisterCommands(session *discordgo.Session) {
	registerSlashBasicCommand(session)
	s.registerMessagePlayCommand(session)
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
	s.logger.Debug("connecting")
	s.player.Connect(m.GuildID, id)

	s.logger.Debug("PlayYoutube")
	_, err = s.player.PlayYoutube(query)
	if err != nil {
		s.logger.Error(err)
		// TODO: Disconnect
		return
	}
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
		return "", errors.New("cog: Unable to find voice channel")
	}

	return id, nil
}
