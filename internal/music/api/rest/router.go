package rest

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/HalvaPovidlo/halvabot-go/internal/pkg"
)

type Player interface {
	Play(ctx context.Context, query, userID, guildID, channelID string) (*pkg.Song, int, error)
	Skip(ctx context.Context)
	SetLoop(ctx context.Context, b bool)
	LoopStatus() bool
	SetRadio(ctx context.Context, b bool, guildID, channelID string) error
	RadioStatus() bool
	NowPlaying() *pkg.Song
	SongStatus() pkg.SessionStats
	Status() pkg.PlayerStatus
}

// Handler TODO: Auth
type Handler struct {
	player Player
	super  *gin.RouterGroup
	logger *zap.Logger
}

func NewHandler(player Player, superGroup *gin.RouterGroup, logger *zap.Logger) *Handler {
	return &Handler{
		player: player,
		super:  superGroup,
		logger: logger,
	}
}

func (h *Handler) Router() *gin.RouterGroup {
	music := h.super.Group("/music")
	music.POST("/enqueue", h.enqueueHandler)
	music.GET("/skip", h.skipHandler)
	music.GET("/stop", h.stopHandler)
	music.GET("/loopstatus", h.loopStatusHandler)
	music.POST("/setloop", h.setLoopHandler)
	music.GET("/radiostatus", h.radioStatusHandler)
	music.POST("/setradio", h.setRadioHandler)
	music.GET("/songstatus", h.songStatusHandler)
	music.GET("/status", h.statusHandler)
	music.GET("/now", h.nowPlayingHandler)
	return music
}

type Response struct {
	Message string `json:"message"`
}
