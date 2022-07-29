package rest

import (
	"context"
	"github.com/HalvaPovidlo/halvabot-go/internal/pkg/item"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type player interface {
	Play(ctx context.Context, query, userID, guildID, channelID string) (*item.Song, int, error)
	Skip(ctx context.Context)
	SetLoop(ctx context.Context, b bool)
	LoopStatus() bool
	SetRadio(ctx context.Context, b bool, guildID, channelID string) error
	RadioStatus() bool
	NowPlaying() *item.Song
	SongStatus() item.SessionStats
	Status() item.PlayerStatus
}

type authorizer interface {
	Authorization() gin.HandlerFunc
}

type Handler struct {
	player player
	auth   authorizer
	super  *gin.RouterGroup
	logger *zap.Logger
}

func NewHandler(player player, auth authorizer, superGroup *gin.RouterGroup, logger *zap.Logger) *Handler {
	return &Handler{
		player: player,
		auth:   auth,
		super:  superGroup,
		logger: logger,
	}
}

func (h *Handler) Route() {
	music := h.super.Group("/music")
	// Unauthorized endpoints
	music.GET("/loopstatus", h.loopStatusHandler)
	music.GET("/radiostatus", h.radioStatusHandler)
	music.GET("/songstatus", h.songStatusHandler)
	music.GET("/now", h.nowPlayingHandler)
	music.GET("/status", h.statusHandler)

	// Authorized endpoints
	music.Use(h.auth.Authorization())
	music.POST("/enqueue", h.enqueueHandler)
	music.GET("/skip", h.skipHandler)
	music.GET("/stop", h.stopHandler)
	music.POST("/setloop", h.setLoopHandler)
	music.POST("/setradio", h.setRadioHandler)
}

type Response struct {
	Message string `json:"message"`
}
