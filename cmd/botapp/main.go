package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	ytdl "github.com/kkdai/youtube/v2"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"

	"github.com/HalvaPovidlo/halvabot-go/cmd/config"
	v1 "github.com/HalvaPovidlo/halvabot-go/internal/api/v1"
	v1chess "github.com/HalvaPovidlo/halvabot-go/internal/api/v1/chess"
	v1film "github.com/HalvaPovidlo/halvabot-go/internal/api/v1/film"
	v1login "github.com/HalvaPovidlo/halvabot-go/internal/api/v1/login"
	v1music "github.com/HalvaPovidlo/halvabot-go/internal/api/v1/music"
	"github.com/HalvaPovidlo/halvabot-go/internal/chess/lichess"
	"github.com/HalvaPovidlo/halvabot-go/internal/film"
	fstorage "github.com/HalvaPovidlo/halvabot-go/internal/film/storage"
	"github.com/HalvaPovidlo/halvabot-go/internal/login"
	"github.com/HalvaPovidlo/halvabot-go/internal/music"
	"github.com/HalvaPovidlo/halvabot-go/internal/music/audio"
	"github.com/HalvaPovidlo/halvabot-go/internal/music/player"
	ytsearch "github.com/HalvaPovidlo/halvabot-go/internal/music/search/youtube"
	"github.com/HalvaPovidlo/halvabot-go/internal/music/storage/firestore"
	"github.com/HalvaPovidlo/halvabot-go/pkg/contexts"
	dpkg "github.com/HalvaPovidlo/halvabot-go/pkg/discord"
	"github.com/HalvaPovidlo/halvabot-go/pkg/http/jwt"
	"github.com/HalvaPovidlo/halvabot-go/pkg/log"
	pfirestore "github.com/HalvaPovidlo/halvabot-go/pkg/storage/firestore"
)

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
	defer session.Close()
	// Load master
	loadMaster := music.NewLoadMaster(ctx, 12*time.Hour)

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
	musicService := v1music.NewMusicService(
		ctx,
		fireService,
		ytClient,
		player.NewPlayer(ctx, audio.NewVoiceClient(session), audio.NewPlayer(loadMaster, &cfg.Discord.Voice.EncodeOptions)),
	)

	// Chess
	lichessClient := lichess.NewClient()

	// Discord commands
	v1music.NewDiscordMusicHandler(musicService, cfg.Discord.Prefix, cfg.Discord.API).RegisterCommands(ctx, session, cfg.General.Debug, logger)
	v1chess.NewDiscordChessHandler(cfg.Discord.Prefix, lichessClient).RegisterCommands(session, cfg.General.Debug, logger)

	// Auth stage
	loginHandler := v1login.NewLoginHandler(login.NewLoginService(login.NewAccountStorage(fireClient), jwt.NewJWTokenizer(cfg.Secret)))

	// Films stage
	filmHandler := v1film.NewFilmHandler(film.NewService(fstorage.NewStorage(fireClient), cfg.Kinopoisk))

	// Http routers
	server := v1.NewServer(loginHandler, v1music.NewMusicHandler(musicService, logger), filmHandler)
	server.Run(cfg.Host.IP, cfg.Host.Bot, config.SwaggerPath, cfg.General.Debug)

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
