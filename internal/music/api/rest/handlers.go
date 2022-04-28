package rest

//go:generate swag fmt -d ./ -g play.go

import (
	"github.com/HalvaPovidlo/discordBotGo/internal/pkg"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type songQuery struct {
	Song string `json:"song" binding:"required"`
}

type loopQuery struct {
	Enable bool `json:"enable" binding:"required"`
}

type EnqueueResponse struct {
	Song           pkg.SongRequest `json:"song"`
	PlaybacksCount int             `json:"playbacks_count"`
}

// enqueue godoc
// @summary  Play the song from YouTube by name or url
// @accept   json
// @produce  json
// @param    query  body      songQuery        true  "Song name or url"
// @success  200    {object}  EnqueueResponse  "The song that was added to the queue"
// @failure  400    {object}  Response         "Incorrect input"
// @failure  500    {object}  Response         "Unknown internal error occurred"
// @failure  500    {object}  Response         "Internal error. This does not necessarily mean that the song will not play. For example, if there is a database error, the song will still be added to the queue."
// @router   /discord/music/enqueue [post]
func (h *Handler) enqueueHandler(c *gin.Context) {
	var json songQuery
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	song, err := h.youtube.FindSong(json.Song)
	if err != nil {
		c.JSON(http.StatusNotFound, Response{Message: errors.Wrap(err, "song not found").Error()})
		return
	}
	p, err := h.player.Enqueue(song)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Message: errors.Wrap(err, "song added to the queue").Error()})
	}
	c.JSON(http.StatusOK, EnqueueResponse{Song: *song, PlaybacksCount: p})
}

// skip godoc
// @summary  Skip the current song and play next from the queue
// @produce  json
// @success  200  string  string
// @router   /discord/music/skip [get]
func (h *Handler) skipHandler(c *gin.Context) {
	h.player.Skip()
	c.String(http.StatusOK, "")
}

// stop godoc
// @summary  Skip the current song and play next from the queue
// @produce  plain
// @success  200  string  string
// @router   /discord/music/stop [get]
func (h *Handler) stopHandler(c *gin.Context) {
	// h.player.Stop()
	c.String(http.StatusNotImplemented, "")
}

// loopStatus godoc
// @summary  Is loop mode enabled
// @produce  plain
// @success  200  string  string  "Returns true or false as string"
// @router   /discord/music/loopstatus [get]
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
// @summary  Set loop mode
// @accept   json
// @produce  json
// @param    query  body      loopQuery  true  "Song name or url"
// @success  200    string    string
// @failure  400    {object}  Response  "Incorrect input"
// @router   /discord/music/setloop [post]
func (h *Handler) setLoopHandler(c *gin.Context) {
	var json loopQuery
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.String(http.StatusOK, "")
}

// stats godoc
// @summary  Stats of player on current song
// @produce  json
// @success  200  {object}  audio.SessionStats  "The song that is playing right now"
// @router   /discord/music/stats [get]
func (h *Handler) statsHandler(c *gin.Context) {
	entry := h.player.Stats()
	c.JSON(http.StatusOK, entry)
}

// nowPlaying godoc
// @summary  Song that is playing now
// @produce  json
// @success  200  {object}  pkg.SongRequest  "The song that is playing right now"
// @router   /discord/music/now [get]
func (h *Handler) nowPlayingHandler(c *gin.Context) {
	entry := h.player.NowPlaying()
	c.JSON(http.StatusOK, entry)
}
