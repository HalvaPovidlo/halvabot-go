package command

import "github.com/bwmarrin/discordgo"

type Slash struct {
	handler func(s *discordgo.Session, i *discordgo.InteractionCreate)
	*discordgo.ApplicationCommand
}

func NewSlashCommand(
	command *discordgo.ApplicationCommand,
	handler func(s *discordgo.Session, i *discordgo.InteractionCreate),
) *Slash {
	return &Slash{
		handler:            handler,
		ApplicationCommand: command,
	}
}

// RegisterCommand TODO: Handle errors
func (c *Slash) RegisterCommand(s *discordgo.Session) (unregisterCommand func()) {
	s.AddHandler(c.handler)
	cmd, err := s.ApplicationCommandCreate(s.State.User.ID, "", c.ApplicationCommand)
	if err != nil {
		return nil
	}
	return func() {
		s.ApplicationCommandDelete(s.State.User.ID, "", cmd.ID)
	}
}
