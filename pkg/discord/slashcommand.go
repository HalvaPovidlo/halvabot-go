package discord

import "github.com/bwmarrin/discordgo"

type SlashCommand struct {
	handler func(s *discordgo.Session, i *discordgo.InteractionCreate)
	*discordgo.ApplicationCommand
}

func NewSlashCommand(
	command *discordgo.ApplicationCommand,
	handler func(s *discordgo.Session, i *discordgo.InteractionCreate),
) *SlashCommand {
	return &SlashCommand{
		handler:            handler,
		ApplicationCommand: command,
	}
}

// RegisterCommand TODO: Handle errors
func (c *SlashCommand) RegisterCommand(s *discordgo.Session) (unregisterCommand func()) {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		c.handler(s, i)
	})
	cmd, err := s.ApplicationCommandCreate(s.State.User.ID, "", c.ApplicationCommand)
	if err != nil {
		return nil
	}
	return func() {
		s.ApplicationCommandDelete(s.State.User.ID, "", cmd.ID)
	}
}
