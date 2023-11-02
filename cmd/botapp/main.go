package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	v1user "github.com/HalvaPovidlo/halvabot-go/internal/api/v1/user"
	"github.com/HalvaPovidlo/halvabot-go/internal/chess/lichess"
	"github.com/HalvaPovidlo/halvabot-go/internal/film"
	fstorage "github.com/HalvaPovidlo/halvabot-go/internal/film/storage"
	"github.com/HalvaPovidlo/halvabot-go/internal/login"
	"github.com/HalvaPovidlo/halvabot-go/internal/music"
	"github.com/HalvaPovidlo/halvabot-go/internal/music/player"
	ytsearch "github.com/HalvaPovidlo/halvabot-go/internal/music/search/youtube"
	"github.com/HalvaPovidlo/halvabot-go/internal/music/storage/firestore"
	"github.com/HalvaPovidlo/halvabot-go/internal/music/voice"
	"github.com/HalvaPovidlo/halvabot-go/internal/user"
	ustorage "github.com/HalvaPovidlo/halvabot-go/internal/user/storage"
	"github.com/HalvaPovidlo/halvabot-go/pkg/contexts"
	"github.com/HalvaPovidlo/halvabot-go/pkg/discord"
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
	ctx, cancel := context.WithCancel(contexts.WithLogger(context.Background(), logger))

	// Initialize discord session
	discordService := discord.New(cfg.Discord.Token, cfg.Discord.Prefix, cfg.General.Debug, logger)
	// Load master
	loadMaster := music.NewLoadMaster(ctx, 12*time.Hour)

	// Cache
	songsCache := firestore.NewSongsCache(ctx, 12*time.Hour)
	defer songsCache.Clear()

	// YouTube services
	ytClient, err := youtube.NewService(ctx, option.WithCredentialsFile("halvabot-google.json"))
	if err != nil {
		logger.Panic("youtube init failed", zap.Error(err))
	}
	ytService := ytsearch.NewYouTubeClient(ytClient, ytsearch.NewYouTubeDL(loadMaster), cfg.Youtube)

	// Firestore stage
	fireClient, err := pfirestore.NewFirestoreClient(ctx, "halvabot-firebase.json")
	if err != nil {
		logger.Panic("new firestore client", zap.Error(err))
	}
	fireService, err := firestore.NewFirestoreService(ctx, fireClient, songsCache, cfg.General.Debug)
	if err != nil {
		logger.Panic("new firestore service", zap.Error(err))
	}

	// Music stage
	musicService := v1music.NewMusicService(
		ctx,
		fireService,
		ytService,
		player.NewPlayer(ctx, voice.NewVoiceClient(session), voice.NewPlayer(loadMaster, &cfg.Discord.Voice.EncodeOptions)),
	)

	// Chess
	lichessClient := lichess.NewClient()

	// Discord commands
	v1music.NewDiscordMusicHandler(musicService, cfg.Discord.Prefix, cfg.Discord.API).RegisterCommands(ctx, session, cfg.General.Debug, logger)
	v1chess.NewDiscordChessHandler(cfg.Discord.Prefix, lichessClient).RegisterCommands(discordService)

	// Handlers stage
	loginHandler := v1login.NewLoginHandler(login.NewLoginService(login.NewAccountStorage(fireClient), jwt.NewJWTokenizer(cfg.Secret)))
	filmHandler := v1film.NewFilmHandler(film.NewService(fstorage.NewStorage(fireClient), cfg.Kinopoisk))
	musicHandler := v1music.NewMusicHandler(musicService, logger)
	userHandler := v1user.NewUserHandler(user.NewUserService(ustorage.NewStorage(fireClient)))

	// Http routers
	server := v1.NewServer(loginHandler, musicHandler, filmHandler, userHandler)
	server.Run(cfg.Host.IP, cfg.Host.Bot, config.SwaggerPath, cfg.General.Debug)

	ctx, cancel = signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer cancel()

	if err := discordService.Connect(ctx); err != nil {
		logger.Error("discord connect", zap.Error(err))
	}

	if err := os.RemoveAll("/" + cfg.Youtube.OutputDir + "/"); err != nil {
		logger.Error(err.Error())
	}
	//if err := os.MkdirAll("/"+cfg.Youtube.OutputDir+"/", 0o755); err != nil {
	//	logger.Error(err.Error())
	//}

	logger.Info("Graceful shutdown")
	_ = logger.Sync()
}
