package rest

import (
	"github.com/gin-gonic/gin"

	"github.com/HalvaPovidlo/discordBotGo/internal/discord/audio"
	"github.com/HalvaPovidlo/discordBotGo/internal/pkg"
)

type Player interface {
	Enqueue(s *pkg.SongRequest)
	Skip()
	SetLoop(b bool)
	LoopStatus() bool
	NowPlaying() pkg.SongRequest
	Stats() audio.SessionStats
}

type YouTube interface {
	FindSong(query string) (*pkg.SongRequest, error)
}

// Handler TODO: Auth
type Handler struct {
	player  Player
	youtube YouTube
	super   *gin.RouterGroup
}

func NewHandler(player Player, tube YouTube, superGroup *gin.RouterGroup) *Handler {
	return &Handler{
		player:  player,
		youtube: tube,
		super:   superGroup,
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
