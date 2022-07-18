package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	"github.com/HalvaPovidlo/discordBotGo/cmd/config"
	"github.com/HalvaPovidlo/discordBotGo/docs"
	v1 "github.com/HalvaPovidlo/discordBotGo/internal/api/v1"
	musicrest "github.com/HalvaPovidlo/discordBotGo/internal/music/api/rest"
	"github.com/HalvaPovidlo/discordBotGo/internal/music/player"
	"github.com/HalvaPovidlo/discordBotGo/pkg/log"
)

// @title           HalvaBot mock for testing
// @version         1.0
// @description     A music discord bot.

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:9090
// @BasePath  /api/v1
func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		panic(errors.Wrap(err, "config read failed"))
	}
	logger := log.NewLogger(cfg.General.Debug)

	mock := &player.MockPlayer{}

	// Http routers
	if !cfg.General.Debug {
		gin.SetMode(gin.ReleaseMode)
		gin.DisableConsoleColor()
	}
	router := gin.Default()
	router.Use(v1.CORS())
	docs.SwaggerInfo.Host = cfg.Host.IP + ":" + cfg.Host.Mock
	docs.SwaggerInfo.BasePath = "/api/v1"

	apiRouter := v1.NewAPI(router.Group("/api/v1")).Router()
	musicrest.NewHandler(mock, apiRouter, logger).Router()
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	go func() {
		err := router.Run(":" + cfg.Host.Mock)
		if err != nil {
			logger.Error("run router", zap.Error(err))
			return
		}
	}()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	logger.Info("Graceful shutdown")
	_ = logger.Sync()
}
