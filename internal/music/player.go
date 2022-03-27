package music

import (
	"github.com/bwmarrin/discordgo"

	"github.com/HalvaPovidlo/discordBotGo/pkg/discord"
	"github.com/HalvaPovidlo/discordBotGo/pkg/zap"
)

type Player struct {
	logger *zap.Logger
}

func NewPlayer(logger *zap.Logger) *Player {
	p := Player{
		logger: logger,
	}
	return &p
}

func registerBasicCommand(s *discordgo.Session) (unregisterCommand func()) {
	command := discord.NewSlashCommand(
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
	return command.RegisterCommand(s)
}

func (p *Player) RegisterCommands(s *discordgo.Session) {
	registerBasicCommand(s)
}
