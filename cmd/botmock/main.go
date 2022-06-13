package botmock

import (
	"github.com/HalvaPovidlo/discordBotGo/cmd/config"
	"github.com/HalvaPovidlo/discordBotGo/docs"
	musicrest "github.com/HalvaPovidlo/discordBotGo/internal/music/api/rest"
	"github.com/HalvaPovidlo/discordBotGo/internal/music/player"
	"github.com/HalvaPovidlo/discordBotGo/pkg/zap"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"os"
	"os/signal"
	"syscall"
)

// @title           HalvaBot for Discord
// @version         1.0
// @description     A music discord bot.

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:9090
// @BasePath  /api/v1
func main() {
	// TODO: all magic vars to config
	cfg, err := config.InitConfig()
	if err != nil {
		panic(errors.Wrap(err, "config read failed"))
	}
	logger := zap.NewLogger(cfg.General.Debug)
	//ctx, cancel := contexts.WithLogger(contexts.Background(), logger)

	mock := &player.MockPlayer{}

	// Http routers
	if !cfg.General.Debug {
		gin.SetMode(gin.ReleaseMode)
		gin.DisableConsoleColor()
	}
	router := gin.Default()
	docs.SwaggerInfo.Host = cfg.Host.IP + ":" + cfg.Host.Bot
	docs.SwaggerInfo.BasePath = "/api/v1"
	apiRouter := router.Group("/api/v1")
	musicrest.NewHandler(mock, apiRouter).Router()
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	go func() {
		err := router.Run(":" + cfg.Host.Bot)
		if err != nil {
			logger.Error(err)
			return
		}
	}()

	// TODO: Graceful shutdown
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	logger.Infow("Graceful shutdown")
	_ = logger.Sync()
}
