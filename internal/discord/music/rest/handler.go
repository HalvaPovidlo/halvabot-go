package rest

import "github.com/gin-gonic/gin"

type Handler struct {
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
