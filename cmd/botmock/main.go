package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"

	"github.com/HalvaPovidlo/halvabot-go/cmd/config"
	v1 "github.com/HalvaPovidlo/halvabot-go/internal/api/v1"
	v1film "github.com/HalvaPovidlo/halvabot-go/internal/api/v1/film"
	v1login "github.com/HalvaPovidlo/halvabot-go/internal/api/v1/login"
	"github.com/HalvaPovidlo/halvabot-go/internal/api/v1/music"
	v1user "github.com/HalvaPovidlo/halvabot-go/internal/api/v1/user"
	"github.com/HalvaPovidlo/halvabot-go/internal/film"
	"github.com/HalvaPovidlo/halvabot-go/internal/login"
	"github.com/HalvaPovidlo/halvabot-go/internal/user"
	"github.com/HalvaPovidlo/halvabot-go/pkg/http/jwt"
	"github.com/HalvaPovidlo/halvabot-go/pkg/log"
)

func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		panic(errors.Wrap(err, "config read failed"))
	}
	logger := log.NewLogger(cfg.General.Debug)

	loginHandler := v1login.NewLoginHandler(login.NewLoginService(login.NewMockStorage(), jwt.NewJWTokenizer("mock_secret")))
	musicHandler := music.NewMusicHandler(&music.MockPlayer{}, logger)
	filmHandler := v1film.NewFilmHandler(film.NewMock())
	userHandler := v1user.NewUserHandler(&user.Mock{})

	// Http routers
	server := v1.NewServer(loginHandler, musicHandler, filmHandler, userHandler)
	server.Run(cfg.Host.IP, cfg.Host.Mock, config.SwaggerPath, cfg.General.Debug)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	logger.Info("Graceful shutdown")
	_ = logger.Sync()
}
