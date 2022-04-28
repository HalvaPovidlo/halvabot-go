package command

import "github.com/bwmarrin/discordgo"

type slashHandler func(s *discordgo.Session, i *discordgo.InteractionCreate)

type Slash struct {
	handler slashHandler
	*discordgo.ApplicationCommand
	debug bool
}

func NewSlashCommand(
	command *discordgo.ApplicationCommand,
	handler slashHandler,
	debug bool,
) *Slash {
	return &Slash{
		handler:            handler,
		ApplicationCommand: command,
		debug:              debug,
	}
}

// RegisterCommand TODO: Handle errors
func (c *Slash) RegisterCommand(s *discordgo.Session) (unregisterCommand func()) {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		channel, _ := s.Channel(i.ChannelID)
		if (channel.Name == "debug") != c.debug {
			return
		}
	})
	cmd, err := s.ApplicationCommandCreate(s.State.User.ID, "", c.ApplicationCommand)
	if err != nil {
		return nil
	}
	return func() {
		s.ApplicationCommandDelete(s.State.User.ID, "", cmd.ID)
	}
}
