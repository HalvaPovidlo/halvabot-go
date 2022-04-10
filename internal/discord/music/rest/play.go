package rest

//go:generate swag fmt -d ./ -g play.go

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/HalvaPovidlo/discordBotGo/internal/discord/voice"
)

// ShowAccount godoc
// @summary        Swagger Example API
// @produce      json
// @param        song  query     string  false  "Название песни или url"
// @success      200              {object} voice.QueueEntry  "ok"
// @failure      400              {object}  Response "error"
// @Router  /discord/music/play [get]
func (h *Handler) play(c *gin.Context) {
	q := voice.QueueEntry{}
	firstname := c.DefaultQuery("firstname", "Guest")
	q.ServiceName = firstname
	//lastname := c.Query("lastname") // shortcut for c.Request.URL.Query().Get("lastname")

	c.JSON(http.StatusOK, q)

}
