package discord

import (
	"go.uber.org/zap"

	"github.com/bwmarrin/discordgo"
)

const (
	// ChannelDebugID TODO: load from session and guildID
	ChannelDebugID = "747743319039148062"

	MonkaS               = "<:monkaS:817041877718138891>"
	MessageInternalError = ":x: **Internal error** " + MonkaS
)

func OpenSession(token string, debug bool, logger *zap.Logger) (*discordgo.Session, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		logger.Error("error creating Discord session", zap.Error(err))
		return nil, err
	}
	logger.Info("bot initialized")

	// session.AddHandler(func(s *discordgo.Session, r *discordgo.GuildCreate) {
	//	logger.Infof("logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	//	guilds := s.State.Guilds
	//	for _, guild := range guilds {
	//		fmt.Println(guild.ID, len(guild.VoiceStates), guild.Name)
	//		for _, state := range guild.VoiceStates {
	//			fmt.Println(state.UserID, state.ChannelID, state.GuildID)
	//		}
	//	}
	//	fmt.Println("Ready with", len(guilds), "guilds.")
	// })

	session.Identify.Intents = discordgo.IntentsAll
	session.LogLevel = discordgo.LogInformational
	if debug {
		session.LogLevel = discordgo.LogDebug
	}
	err = session.Open()
	if err != nil {
		logger.Error("error opening connection", zap.Error(err))
		return nil, err
	}

	logger.Info("bot session opened", zap.String("SessionID", session.State.SessionID))
	return session, nil
}
