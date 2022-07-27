package main

import (
	"context"
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
	"go.uber.org/zap"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"

	"github.com/HalvaPovidlo/halvabot-go/cmd/config"
	docs "github.com/HalvaPovidlo/halvabot-go/docs/botapp"
	v1 "github.com/HalvaPovidlo/halvabot-go/internal/api/v1"
	capi "github.com/HalvaPovidlo/halvabot-go/internal/chess/api/discord"
	"github.com/HalvaPovidlo/halvabot-go/internal/chess/lichess"
	dapi "github.com/HalvaPovidlo/halvabot-go/internal/music/api/discord"
	musicrest "github.com/HalvaPovidlo/halvabot-go/internal/music/api/rest"
	"github.com/HalvaPovidlo/halvabot-go/internal/music/audio"
	"github.com/HalvaPovidlo/halvabot-go/internal/music/player"
	ytsearch "github.com/HalvaPovidlo/halvabot-go/internal/music/search/youtube"
	"github.com/HalvaPovidlo/halvabot-go/internal/music/storage/firestore"
	"github.com/HalvaPovidlo/halvabot-go/internal/pkg"
	"github.com/HalvaPovidlo/halvabot-go/pkg/contexts"
	dpkg "github.com/HalvaPovidlo/halvabot-go/pkg/discord"
	"github.com/HalvaPovidlo/halvabot-go/pkg/http/jwt"
	"github.com/HalvaPovidlo/halvabot-go/pkg/http/login"
	"github.com/HalvaPovidlo/halvabot-go/pkg/log"
	pfirestore "github.com/HalvaPovidlo/halvabot-go/pkg/storage/firestore"
)

// @title       HalvaBot for Discord
// @version     1.0
// @description A music discord bot.

// @license.name Apache 2.0
// @license.url  http://www.apache.org/licenses/LICENSE-2.0.html

// @host     localhost:9091
// @BasePath /api/v1
func main() {
	// TODO: all magic vars to config
	cfg, err := config.InitConfig()
	if err != nil {
		panic(errors.Wrap(err, "config read failed"))
	}
	logger := log.NewLogger(cfg.General.Debug)
	ctx := contexts.WithLogger(context.Background(), logger)
	ctx, cancel := context.WithCancel(ctx)

	// Initialize discord session
	session, err := dpkg.OpenSession(cfg.Discord.Token, cfg.General.Debug, logger)
	if err != nil {
		logger.Panic("discord open session failed", zap.Error(err))
	}
	defer func() {
		err = session.Close()
		if err != nil {
			logger.Error("close session", zap.Error(err))
		} else {
			logger.Info("Bot session closed")
		}
	}()
	// Load master
	loadMaster := pkg.NewLoadMaster(ctx, 12*time.Hour)

	// Cache
	songsCache := firestore.NewSongsCache(ctx, 12*time.Hour)
	defer songsCache.Clear()

	// YouTube services
	ytService, err := youtube.NewService(ctx, option.WithCredentialsFile("halvabot-google.json"))
	if err != nil {
		logger.Panic("youtube init failed", zap.Error(err))
	}
	ytdlClient := ytdl.Client{Debug: cfg.General.Debug, HTTPClient: http.DefaultClient}
	ytClient := ytsearch.NewYouTubeClient(
		&ytdlClient,
		ytService,
		ytsearch.NewDownloader(ytdlClient, cfg.Youtube.OutputDir, loadMaster),
		cfg.Youtube,
	)

	// Firestore stage
	fireClient, err := pfirestore.NewFirestoreClient(ctx, "halvabot-firebase.json")
	if err != nil {
		logger.Panic("new firestore client", zap.Error(err))
	}
	fireStorage, err := firestore.NewFirestoreClient(ctx, fireClient, cfg.General.Debug)
	if err != nil {
		logger.Panic("new firestore storage", zap.Error(err))
	}
	fireService, err := firestore.NewFirestoreService(ctx, fireStorage, songsCache)
	if err != nil {
		logger.Panic("new firestore service", zap.Error(err))
	}

	// Music stage
	voiceClient := audio.NewVoiceClient(session)
	rawAudioPlayer := audio.NewPlayer(loadMaster, &cfg.Discord.Voice.EncodeOptions)
	musicPlayer := player.NewMusicService(ctx, fireService, ytClient, voiceClient, rawAudioPlayer)

	// Chess
	lichessClient := lichess.NewClient()

	// Discord commands
	musicCog := dapi.NewCog(musicPlayer, cfg.Discord.Prefix, cfg.Discord.API)
	musicCog.RegisterCommands(ctx, session, cfg.General.Debug, logger)
	chessCog := capi.NewCog(cfg.Discord.Prefix, lichessClient)
	chessCog.RegisterCommands(session, cfg.General.Debug, logger)

	// Auth stage
	loginer := login.NewLoginService(login.NewFirebaseAuthenticator(fireClient), jwt.NewJWTokenizer(cfg.Secret))

	// Http routers
	if !cfg.General.Debug {
		gin.SetMode(gin.ReleaseMode)
		gin.DisableConsoleColor()
	}
	router := gin.New()
	router.Use(v1.CORS())
	docs.SwaggerInfo.Host = cfg.Host.IP + ":" + cfg.Host.Bot
	docs.SwaggerInfo.BasePath = "/api/v1"
	apiRouter := v1.NewAPIRouter(router, loginer)
	musicrest.NewHandler(musicPlayer, loginer, apiRouter, logger).Route()
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	go func() {
		err := router.Run(":" + cfg.Host.Bot)
		if err != nil {
			logger.Error("run router", zap.Error(err))
			return
		}
	}()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	cancel()
	if err := os.RemoveAll("/" + cfg.Youtube.OutputDir + "/"); err != nil {
		logger.Error(err.Error())
	}
	if err := os.MkdirAll("/"+cfg.Youtube.OutputDir+"/", 0o755); err != nil {
		logger.Error(err.Error())
	}

	logger.Info("Graceful shutdown")
	_ = logger.Sync()
}
