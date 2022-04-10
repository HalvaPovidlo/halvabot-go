package player

import (
	"fmt"
	"github.com/HalvaPovidlo/discordBotGo/internal/discord/search"
	"github.com/HalvaPovidlo/discordBotGo/internal/discord/voice"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/HalvaPovidlo/discordBotGo/cmd/config"
	"github.com/HalvaPovidlo/discordBotGo/pkg/discord/command"
	"github.com/HalvaPovidlo/discordBotGo/pkg/util"
	"github.com/HalvaPovidlo/discordBotGo/pkg/zap"
)

const (
	play = "play "
)

type Player struct {
	logger  *zap.Logger
	config  config.DiscordConfig
	voice   *voice.Voice
	youtube *search.YouTube
}

func NewPlayer(youtube *search.YouTube, voiceClient *voice.Voice, config config.DiscordConfig, logger *zap.Logger) *Player {
	p := Player{
		youtube: youtube,
		voice:   voiceClient,
		logger:  logger,
		config:  config,
	}
	return &p
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

func (p *Player) registerMessagePlayCommand(s *discordgo.Session) {
	mc := command.NewMessageCommand(p.config.Prefix+play, p.playMessageHandler)
	mc.RegisterCommand(s)
}

func (p *Player) RegisterCommands(s *discordgo.Session) {
	registerSlashBasicCommand(s)
	p.registerMessagePlayCommand(s)
}

func (p *Player) playMessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	query := strings.TrimPrefix(m.Content, p.config.Prefix+play)
	query = util.StandardizeSpaces(query)
	fmt.Println("query parsed", query)

	fmt.Println(m.GuildID)
	guild, err := s.State.Guild(m.GuildID)
	fmt.Println(err)
	fmt.Println(guild.Name)
	id := ""
	fmt.Println(len(guild.VoiceStates))
	for _, voiceState := range guild.VoiceStates {
		fmt.Println(voiceState.UserID)
		if voiceState.UserID == m.Author.ID {
			id = voiceState.ChannelID
			break
		}
	}
	fmt.Println(id)
	err = p.voice.Connect(m.GuildID, id)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = p.playYoutube(query)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (p *Player) playYoutube(query string) error {
	song, err := p.youtube.FindSong(query)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("Found song", song.Metadata.Title)
	// TODO: Announce according to debug settings
	err = p.voice.Play(song, true)
	return err
}
