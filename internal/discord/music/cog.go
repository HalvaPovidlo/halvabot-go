package music

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"

	"github.com/HalvaPovidlo/discordBotGo/internal/discord/voice"
	"github.com/HalvaPovidlo/discordBotGo/pkg/discord/command"
	"github.com/HalvaPovidlo/discordBotGo/pkg/util"
	"github.com/HalvaPovidlo/discordBotGo/pkg/zap"
)

const (
	play = "play "
)

// Player TODO: loop, disconnect commands
type Player interface {
	PlayYoutube(query string) (*voice.QueueEntry, error)
	Skip() (*voice.QueueEntry, error)
	Stop() error
	Connect(guildID, channelID string) error
}

type Cog struct {
	player Player
	prefix string
	logger *zap.Logger
}

func NewCog(player Player, prefix string, logger *zap.Logger) *Cog {
	c := Cog{
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

func (c *Cog) registerMessagePlayCommand(s *discordgo.Session) {
	mc := command.NewMessageCommand(c.prefix+play, c.playMessageHandler)
	mc.RegisterCommand(s)
}

func (c *Cog) RegisterCommands(s *discordgo.Session) {
	registerSlashBasicCommand(s)
	c.registerMessagePlayCommand(s)
}

func (c *Cog) playMessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	query := strings.TrimPrefix(m.Content, c.prefix+play)
	query = util.StandardizeSpaces(query)

	id, err := findAuthorVoiceChannelID(s, m)
	if err != nil {
		// TODO: log error
		return
	}
	err = c.player.Connect(m.GuildID, id)
	if err != nil {
		return
	}

	_, err = c.player.PlayYoutube(query)
	if err != nil {
		fmt.Println(err)
		return
	}
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
		return "", errors.New("cog: Unable to find voice channel")
	}

	return id, nil
}
