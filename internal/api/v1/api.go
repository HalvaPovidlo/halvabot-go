package v1

import (
	"github.com/gin-gonic/gin"
)

type LoginService interface {
	LoginHandler(c *gin.Context)
}

func NewAPIRouter(super *gin.Engine, login LoginService) *gin.RouterGroup {
	api := super.Group("/api/v1")
	api.Use(
		gin.LoggerWithConfig(gin.LoggerConfig{
			SkipPaths: []string{
				"/api/v1/music/now",
				"/api/v1/music/status",
				"/api/v1/music/songstatus",
			},
		}),
		gin.Recovery(),
		CORS(),
	)
	api.POST("/login", login.LoginHandler)
	return api
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
