package discord

import (
	"github.com/bwmarrin/discordgo"

	"github.com/HalvaPovidlo/discordBotGo/pkg/zap"
)

const (
	// ChannelDebugID TODO: load from session and guildID
	ChannelDebugID = "747743319039148062"
)

func OpenSession(token string, debug bool, logger zap.Logger) (*discordgo.Session, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		logger.Errorw("error creating Discord session",
			"err", err)
		return nil, err
	}
	logger.Infow("Bot initialized")

	// session.AddHandler(func(s *discordgo.Session, r *discordgo.GuildCreate) {
	//	logger.Infof("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
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
		logger.Errorw("error opening connection", "err", err)
		return nil, err
	}

	logger.Infow("Bot session opened", "SessionID", session.State.SessionID)
	return session, nil
}
