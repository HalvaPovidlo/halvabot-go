package rest

import (
	"github.com/gin-gonic/gin"

	"github.com/HalvaPovidlo/discordBotGo/internal/audio"
	"github.com/HalvaPovidlo/discordBotGo/internal/pkg"
	"github.com/HalvaPovidlo/discordBotGo/pkg/contexts"
)

type Player interface {
	Enqueue(ctx contexts.Context, query string) (*pkg.Song, int, error)
	Skip()
	SetLoop(b bool)
	LoopStatus() bool
	NowPlaying() *pkg.Song
	Stats() audio.SessionStats
}

// Handler TODO: Auth
type Handler struct {
	player Player
	super  *gin.RouterGroup
}

func NewHandler(player Player, superGroup *gin.RouterGroup) *Handler {
	return &Handler{
		player: player,
		super:  superGroup,
	}
}

func (h *Handler) Router() *gin.RouterGroup {
	music := h.super.Group("/music")
	music.POST("/enqueue", h.enqueueHandler)
	music.GET("/skip", h.skipHandler)
	music.GET("/stop", h.stopHandler)
	music.GET("/loopstatus", h.loopStatusHandler)
	music.POST("/setloop", h.setLoopHandler)
	music.GET("/stats", h.statsHandler)
	music.GET("/now", h.nowPlayingHandler)
	return music
}

type Response struct {
	Message string `json:"message"`
}
