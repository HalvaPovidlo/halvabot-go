package v1

import (
	"github.com/gin-gonic/gin"
)

type API struct {
	super *gin.RouterGroup
}

func NewAPI(superGroup *gin.RouterGroup) *API {
	return &API{
		super: superGroup,
	}
}

func (h *API) Router() *gin.RouterGroup {
	h.super.Use(
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
	return h.super
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
