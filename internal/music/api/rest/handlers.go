package rest

//go:generate swag fmt -d ./ -g play.go

import (
	"github.com/HalvaPovidlo/halvabot-go/internal/pkg/item"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/HalvaPovidlo/halvabot-go/internal/login"
	"github.com/HalvaPovidlo/halvabot-go/pkg/contexts"
)

type songQuery struct {
	Song string `json:"song" binding:"required"`
}

type enableQuery struct {
	Enable bool `json:"enable" binding:"exists"`
}

type EnqueueResponse struct {
	Song           item.Song `json:"song"`
	PlaybacksCount int       `json:"playbacks_count"`
}

// enqueue godoc
// @summary                    Play the song from YouTube by name or url
// @securityDefinitions.apiKey JWT
// @in                         header
// @name                       Authorization
// @Tags    music,jwtAuth
// @accept                     json
// @produce                    json
// @param                      Authorization header   string      true "Bearer"
// @param                      query         body     songQuery       true "Song name or url"
// @success                    200           {object} EnqueueResponse "The song that was added to the queue"
// @failure                    400           {object} Response        "Incorrect input"
// @failure                    500           {object} Response        "Internal error. This does not necessarily mean that the song will not play. For example, if there is a database error, the song will still be added to the queue."
// @router                     /music/enqueue [post]
func (h *Handler) enqueueHandler(c *gin.Context) {
	var json songQuery
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := contexts.WithValues(c, h.logger, "")
	song, playbacks, err := h.player.Play(ctx, json.Song, c.GetString(login.UserID), "", "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Message: err.Error()})
	}
	c.JSON(http.StatusOK, EnqueueResponse{Song: *song, PlaybacksCount: playbacks})
}

// skip godoc
// @summary                    Skip the current song and play next from the queue
// @securityDefinitions.apiKey JWT
// @in                         header
// @name                       Authorization
// @Tags    music,jwtAuth
// @produce                    json
// @param                      Authorization header string true "Bearer"
// @success                    200           string string
// @router                     /music/skip [get]
func (h *Handler) skipHandler(c *gin.Context) {
	ctx := contexts.WithValues(c, h.logger, "")
	h.player.Skip(ctx)
	c.Status(http.StatusOK)
}

// stop godoc
// @summary                    Skip the current song and play next from the queue
// @securityDefinitions.apiKey JWT
// @in                         header
// @name                       Authorization
// @Tags    music,jwtAuth
// @produce                    plain
// @param                      Authorization header string true "Bearer"
// @success                    200           string string
// @router                     /music/stop [get]
func (h *Handler) stopHandler(c *gin.Context) {
	// h.player.Stop()
	c.String(http.StatusNotImplemented, "")
}

// loopStatus godoc
// @summary Is loop mode enabled
// @Tags    music
// @produce plain
// @success 200 string string "Returns true or false as string"
// @router  /music/loopstatus [get]
func (h *Handler) loopStatusHandler(c *gin.Context) {
	resp := ""
	if h.player.LoopStatus() {
		resp = "true"
	} else {
		resp = "false"
	}
	c.String(http.StatusOK, resp)
}

// setLoop godoc
// @summary                    Set loop mode
// @securityDefinitions.apiKey JWT
// @in                         header
// @name                       Authorization
// @Tags    music,jwtAuth
// @accept                     json
// @produce                    json
// @param                      Authorization header   string      true "Bearer"
// @param                      query         body     enableQuery true "Send true to enable and false to disable"
// @success                    200           string   string
// @failure                    400           {object} Response "Incorrect input"
// @router                     /music/setloop [post]
func (h *Handler) setLoopHandler(c *gin.Context) {
	var json enableQuery
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.String(http.StatusOK, "")
}

// SongStatus godoc
// @summary SongStatus of player on the current song
// @Tags                       music
// @produce json
// @success 200 {object} pkg.SessionStats "The song that is playing right now"
// @router  /music/songstatus [get]
func (h *Handler) songStatusHandler(c *gin.Context) {
	entry := h.player.SongStatus()
	c.JSON(http.StatusOK, entry)
}

// Status godoc
// @summary Status of the player
// @Tags                       music
// @produce json
// @success 200 {object} pkg.PlayerStatus "Status of the player"
// @router  /music/status [get]
func (h *Handler) statusHandler(c *gin.Context) {
	status := h.player.Status()
	if status.Now == nil {
		status.Now = &item.Song{}
	}
	c.JSON(http.StatusOK, status)
}

// nowPlaying godoc
// @summary Song that is playing now
// @Tags                       music
// @produce json
// @success 200 {object} pkg.Song "The song that is playing right now"
// @router  /music/now [get]
func (h *Handler) nowPlayingHandler(c *gin.Context) {
	entry := h.player.NowPlaying()
	if entry == nil {
		c.JSON(http.StatusOK, &item.Song{})
		return
	}
	c.JSON(http.StatusOK, entry)
}

// radiostatus godoc
// @summary Is radio mode enabled
// @Tags                       music
// @produce plain
// @success 200 string string "Returns true or false as string"
// @router  /music/radiostatus [get]
func (h *Handler) radioStatusHandler(c *gin.Context) {
	resp := ""
	if h.player.RadioStatus() {
		resp = "true"
	} else {
		resp = "false"
	}
	c.String(http.StatusOK, resp)
}

// setLoop godoc
// @summary                    Set radio mode
// @securityDefinitions.apiKey JWT
// @in                         header
// @name                       Authorization
// @Tags                       music,jwtAuth
// @accept                     json
// @produce                    json
// @param                      Authorization header   string          true "Bearer"
// @param                      query         body     enableQuery true "Send true to enable and false to disable"
// @success                    200           string   string
// @failure                    400           {object} Response "Incorrect input"
// @router                     /music/setradio [post]
func (h *Handler) setRadioHandler(c *gin.Context) {
	var json enableQuery
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := contexts.WithValues(c, h.logger, "")
	if err := h.player.SetRadio(ctx, json.Enable, "", ""); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.String(http.StatusOK, "")
}
