package rest

import (
	"github.com/HalvaPovidlo/discordBotGo/internal/discord/audio"
	"github.com/HalvaPovidlo/discordBotGo/internal/discord/pkg"
	"github.com/gin-gonic/gin"
)

type Player interface {
	PlayYoutube(query string) (*pkg.SongRequest, error)
	Skip() *pkg.SongRequest
	Stop()
	LoopStatus() bool
	SetLoop(b bool)
	NowPlaying() pkg.SongRequest
	Stats() audio.SessionStats
}

type Handler struct {
	player Player
}

func (h *Handler) Router() *gin.RouterGroup {
	r := gin.Default()
	music := r.Group("/music")
	music.POST("/play", h.playHandler)
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
