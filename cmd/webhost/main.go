package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/HalvaPovidlo/halvabot-go/cmd/config"
	"github.com/HalvaPovidlo/halvabot-go/pkg/log"
)

const (
	homePage = "/"
	// serverURL = "/server"
	// mockURL   = "/mock"
	rootPath = "./www/"
)

func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		panic(err)
	}
	logger := log.NewLogger(cfg.General.Debug)
	if !cfg.General.Debug {
		gin.SetMode(gin.ReleaseMode)
		gin.DisableConsoleColor()
	}

	// serverIPHandler := func(c *gin.Context) {
	//	c.String(http.StatusOK, cfg.Host.IP+":"+cfg.Host.Bot)
	// }
	// mockIPHandler := func(c *gin.Context) {
	//	c.String(http.StatusOK, cfg.Host.IP+":"+cfg.Host.Mock)
	// }

	router := gin.Default()
	router.Use(CORSMiddleware())
	router.Static(homePage, rootPath)
	// router.GET("/", func(c *gin.Context) {
	//	location := url.URL{Path: homePage}
	//	c.Redirect(http.StatusMovedPermanently, location.RequestURI())
	// })
	// router.GET(serverURL, serverIPHandler)
	// router.GET(mockURL, mockIPHandler)
	go func() {
		err := router.Run(":" + cfg.Host.Web)
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

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		c.Next()
	}
}
