package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
	ytdl "github.com/kkdai/youtube/v2"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"

	"github.com/HalvaPovidlo/discordBotGo/cmd/config"
	"github.com/HalvaPovidlo/discordBotGo/internal/discord"
	"github.com/HalvaPovidlo/discordBotGo/internal/discord/audio"
	"github.com/HalvaPovidlo/discordBotGo/internal/discord/music"
	"github.com/HalvaPovidlo/discordBotGo/internal/discord/music/player"
	"github.com/HalvaPovidlo/discordBotGo/internal/search"
	"github.com/HalvaPovidlo/discordBotGo/pkg/contexts"
	discordpkg "github.com/HalvaPovidlo/discordBotGo/pkg/discord"
	"github.com/HalvaPovidlo/discordBotGo/pkg/zap"
)

// @title           HalvaBot for Discord
// @version         1.0
// @description     A music discord bot.

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:9091
// @BasePath  /api
func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		panic(errors.Wrap(err, "config read failed"))
	}
	logger := zap.NewLogger()
	ctx, cancel := contexts.WithLogger(contexts.Background(), logger)

	// Initialize discord session
	session, err := discordpkg.OpenSession(cfg.Discord.Token, logger)
	if err != nil {
		panic(errors.Wrap(err, "discord open session failed"))
	}
	defer func(session *discordgo.Session) {
		_ = session.Close()
		logger.Infow("Bot session closed")
	}(session)

	session.Debug = true
	session.LogLevel = 100

	// YouTube services
	ytService, err := youtube.NewService(ctx, option.WithCredentialsFile("halvabot-google.json"))
	if err != nil {
		panic(errors.Wrap(err, "youtube init failed"))
	}
	ytClient := search.NewYouTubeClient(ctx,
		&ytdl.Client{
			Debug:      true,
			HTTPClient: http.DefaultClient,
		},
		ytService)

	// Music stage
	voiceClient := audio.NewVoiceClient(session)
	rawAudioPlayer := audio.NewPlayer(&cfg.Discord.Voice.EncodeOptions, logger)
	musicPlayer := player.NewPlayer(ctx, voiceClient, rawAudioPlayer, cfg.Discord.Player)

	// Discord commands
	musicCog := music.NewCog(musicPlayer, ytClient, cfg.Discord.Prefix, logger)
	musicCog.RegisterCommands(session)

	// Http routers
	router := gin.Default()
	apiRouter := router.Group("/api")
	discordRouter := discord.NewHandler(apiRouter).Router()
	discord.NewHandler(discordRouter).Router()
	go func() {
		err := router.Run(":9091")
		if err != nil {
			logger.Error(err)
			return
		}
	}()

	// Graceful shutdown
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	cancel()

	logger.Infow("Graceful shutdown")
	_ = logger.Sync()
}
