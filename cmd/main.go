package main

import (
	"fmt"
	"github.com/HalvaPovidlo/discordBotGo/internal/discord/music/player"
	"github.com/HalvaPovidlo/discordBotGo/internal/discord/search"
	"github.com/HalvaPovidlo/discordBotGo/internal/discord/voice"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	ytdl "github.com/kkdai/youtube/v2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"

	"github.com/HalvaPovidlo/discordBotGo/cmd/config"
	"github.com/HalvaPovidlo/discordBotGo/pkg/context"
	"github.com/HalvaPovidlo/discordBotGo/pkg/zap"
)

// @title           Swagger Example API
// @version         1.0
// @description     This is a sample server celler server.

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api
func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		panic(err)
	}
	logger := zap.NewLogger()
	ctx := context.WithLogger(context.Background(), logger)

	session, err := discordgo.New("Bot " + cfg.Discord.Token)
	if err != nil {
		logger.Errorw("error creating Discord session",
			"err", err)
		return
	}
	logger.Infow("Bot initialized")

	session.AddHandler(func(s *discordgo.Session, r *discordgo.GuildCreate) {
		logger.Infof("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
		//s.UpdateStatus(0, conf.DefaultStatus)
		guilds := s.State.Guilds
		for _, guild := range guilds {
			fmt.Println(guild.ID, len(guild.VoiceStates), guild.Name)
			for _, state := range guild.VoiceStates {
				fmt.Println(state.UserID, state.ChannelID, state.GuildID)
			}
		}
		fmt.Println("Ready with", len(guilds), "guilds.")
	})

	session.Identify.Intents = discordgo.IntentsAll
	err = session.Open()
	if err != nil {
		logger.Errorw("error opening connection", "err", err)
		return
	}
	logger.Infow("Bot session opened", "SessionID", session.State.SessionID)

	defer func(session *discordgo.Session) {
		_ = session.Close()
		logger.Infow("Bot session closed")
	}(session)

	ytService, err := youtube.NewService(ctx, option.WithCredentialsFile("halvabot-google.json"))
	if err != nil {
		panic(err)
	}

	ytClient := search.NewYouTubeClient(&ytdl.Client{
		Debug:      true,
		HTTPClient: http.DefaultClient,
	}, ytService)

	voiceClient := voice.NewVoice(session, cfg.Discord.Voice)
	
	musicPlayer := player.NewPlayer(ytClient, voiceClient, cfg.Discord, logger)
	musicPlayer.RegisterCommands(session)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	logger.Infow("Gracefully shutdowning")
	defer logger.Sync()
}
