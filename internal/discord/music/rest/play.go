package rest

//go:generate swag fmt -d ./ -g play.go

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type songQuery struct {
	Song string `json:"song" binding:"required"`
}

// play godoc
// @summary  Play song from youtube by name or url
// @accept   json
// @produce  json
// @param    query  body      songQuery      true  "Название песни или url"
// @success  200    {object}  voice.QueueEntry
// @failure  400    {object}  Response          "Неверные параметры"
// @router   /discord/music/play [post]
func (h *Handler) play(c *gin.Context) {
	var json songQuery
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entry, err := h.player.PlayYoutube(json.Song)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
	}

	c.JSON(http.StatusOK, entry)
}
