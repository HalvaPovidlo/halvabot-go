package main

import (
	musicrest "github.com/HalvaPovidlo/discordBotGo/internal/discord/music/api/rest"
	"github.com/HalvaPovidlo/discordBotGo/internal/storage/firestore"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	ytdl "github.com/kkdai/youtube/v2"
	"github.com/pkg/errors"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"

	"github.com/HalvaPovidlo/discordBotGo/cmd/config"
	"github.com/HalvaPovidlo/discordBotGo/docs"
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
// @BasePath  /api/v1
func main() {
	// TODO: all magic vars to config
	cfg, err := config.InitConfig()
	if err != nil {
		panic(errors.Wrap(err, "config read failed"))
	}
	logger := zap.NewLogger(cfg.General.Debug)
	ctx, cancel := contexts.WithLogger(contexts.Background(), logger)

	// Initialize discord session
	session, err := discordpkg.OpenSession(cfg.Discord.Token, cfg.General.Debug, logger)
	if err != nil {
		panic(errors.Wrap(err, "discord open session failed"))
	}
	defer func() {
		err = session.Close()
		if err != nil {
			logger.Error(errors.Wrap(err, "close session"))
		} else {
			logger.Infow("Bot session closed")
		}
	}()

	// YouTube services
	ytService, err := youtube.NewService(ctx, option.WithCredentialsFile("halvabot-google.json"))
	if err != nil {
		panic(errors.Wrap(err, "youtube init failed"))
	}
	ytClient := search.NewYouTubeClient(ctx,
		&ytdl.Client{
			Debug:      cfg.General.Debug,
			HTTPClient: http.DefaultClient,
		},
		ytService)

	// Firestore stage
	fireSongsCache := firestore.NewSongsCache(ctx, 12*time.Hour)
	fireStorage, err := firestore.NewFirestoreClient(ctx, "halvabot-firebase.json", cfg.General.Debug)
	if err != nil {
		panic(err)
	}
	fireService, err := firestore.NewFirestoreService(ctx, fireStorage, fireSongsCache)
	if err != nil {
		panic(err)
	}

	// Music stage
	voiceClient := audio.NewVoiceClient(session)
	rawAudioPlayer := audio.NewPlayer(&cfg.Discord.Voice.EncodeOptions, logger)
	musicPlayer := player.NewPlayer(ctx, voiceClient, rawAudioPlayer, fireService, cfg.Discord.Player)

	// Discord commands
	musicCog := music.NewCog(musicPlayer, ytClient, cfg.Discord.Prefix, logger)
	musicCog.RegisterCommands(session, cfg.General.Debug)

	// Http routers
	router := gin.Default()
	docs.SwaggerInfo.BasePath = "/api/v1"
	apiRouter := router.Group("/api/v1")
	discordRouter := discord.NewHandler(apiRouter).Router()
	musicrest.NewHandler(musicPlayer, ytClient, discordRouter).Router()
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	go func() {
		err := router.Run(":9091")
		if err != nil {
			logger.Error(err)
			return
		}
	}()

	// TODO: Graceful shutdown
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	cancel()

	logger.Infow("Graceful shutdown")
	_ = logger.Sync()
}
