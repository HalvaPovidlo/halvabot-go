package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"

	"github.com/HalvaPovidlo/halvabot-go/cmd/config"
	v1 "github.com/HalvaPovidlo/halvabot-go/internal/api/v1"
	"github.com/HalvaPovidlo/halvabot-go/internal/api/v1/login"
	"github.com/HalvaPovidlo/halvabot-go/internal/api/v1/music"
	"github.com/HalvaPovidlo/halvabot-go/internal/music/player"
	"github.com/HalvaPovidlo/halvabot-go/pkg/http/jwt"
	"github.com/HalvaPovidlo/halvabot-go/pkg/log"
)

func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		panic(errors.Wrap(err, "config read failed"))
	}
	logger := log.NewLogger(cfg.General.Debug)

	mock := &player.MockPlayer{}
	loginService := login.NewLoginService(login.NewMockStorage(), jwt.NewJWTokenizer("mock_secret"))
	musicService := music.NewMusicHandler(mock, logger)
	// Http routers
	server := v1.NewServer(musicService, loginService)
	server.Run(cfg.Host.IP, cfg.Host.Mock, config.SwaggerPath, cfg.General.Debug)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	logger.Info("Graceful shutdown")
	_ = logger.Sync()
}
