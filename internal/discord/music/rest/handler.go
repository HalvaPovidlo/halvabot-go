package rest

import (
	"github.com/HalvaPovidlo/discordBotGo/internal/discord/voice"
	"github.com/gin-gonic/gin"
)

type Player interface {
	PlayYoutube(query string) (*voice.QueueEntry, error)
}

type Handler struct {
	player Player
}

func (h *Handler) Router() *gin.RouterGroup {
	r := gin.Default()
	music := r.Group("/music")
	music.GET("/play", h.play)
	return music
}

type Response struct {
	Message string `json:"message"`
}
