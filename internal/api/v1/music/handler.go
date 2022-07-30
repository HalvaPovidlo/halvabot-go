package music

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	v1 "github.com/HalvaPovidlo/halvabot-go/internal/api/v1"
	"github.com/HalvaPovidlo/halvabot-go/internal/api/v1/login"
	"github.com/HalvaPovidlo/halvabot-go/internal/music/player"
	"github.com/HalvaPovidlo/halvabot-go/internal/pkg"
	"github.com/HalvaPovidlo/halvabot-go/pkg/contexts"
)

type playerService interface {
	Play(ctx context.Context, query, userID, guildID, channelID string) (*pkg.Song, error)
	Skip(ctx context.Context)
	SetLoop(ctx context.Context, b bool)
	SetRadio(ctx context.Context, b bool, guildID, channelID string) error
	Status() pkg.PlayerStatus
}

type Handler struct {
	player playerService
	logger *zap.Logger
}

func NewMusicHandler(player playerService, logger *zap.Logger) *Handler {
	return &Handler{
		player: player,
		logger: logger,
	}
}

func (h *Handler) PostMusicEnqueueServiceIdentifier(c *gin.Context, service string, kind string) {
	var json v1.PostMusicEnqueueServiceIdentifierJSONRequestBody
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, v1.Error{Msg: err.Error()})
		return
	}
	if v1.SongService(service) == v1.Youtube {
		if kind != "query" {
			c.Status(http.StatusNotImplemented)
			return
		}
		ctx := contexts.WithValues(c, h.logger, "")
		song, err := h.player.Play(ctx, json.Input, c.GetString(login.UserID), "", "")
		switch {
		case errors.Is(err, player.ErrNotConnected):
			c.Status(http.StatusConflict)
			return
		case status.Convert(err).Code() != codes.Unknown:
			c.JSON(http.StatusInsufficientStorage, gin.H{"song": song, "msg": err.Error()})
			return
		case err != nil:
			c.JSON(http.StatusInternalServerError, v1.Error{Msg: err.Error()})
			return
		}
		c.JSON(http.StatusOK, buildSong(song))
		return
	}
	c.Status(http.StatusNotImplemented)
}

func buildSong(song *pkg.Song) *v1.Song {
	return &v1.Song{
		ArtistName:   song.ArtistName,
		ArtistUrl:    song.ArtistURL,
		ArtworkUrl:   song.ArtworkURL,
		LastPlay:     song.LastPlay,
		Playbacks:    song.Playbacks,
		Service:      convertService(song.Service),
		ThumbnailUrl: song.ThumbnailURL,
		Title:        song.Title,
		Url:          song.URL,
	}
}

func convertService(service pkg.ServiceName) v1.SongService {
	if service == pkg.ServiceYouTube {
		return v1.Youtube
	}
	return v1.Unknown
}

func (h *Handler) PostMusicSkip(c *gin.Context) {
	ctx := contexts.WithValues(c, h.logger, "")
	h.player.Skip(ctx)
	c.Status(http.StatusOK)
}

func (h *Handler) PostMusicLoop(c *gin.Context) {
	var json v1.EnableMode
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, v1.Error{Msg: err.Error()})
		return
	}
	c.Status(http.StatusNotImplemented)
}

func (h *Handler) PostMusicRadio(c *gin.Context) {
	var json v1.EnableMode
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, v1.Error{Msg: err.Error()})
		return
	}
	ctx := contexts.WithValues(c, h.logger, "")
	err := h.player.SetRadio(ctx, json.Enable, "", "")
	switch {
	case errors.Is(err, player.ErrNotConnected):
		c.Status(http.StatusConflict)
	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.String(http.StatusOK, "")
}

func (h *Handler) GetMusicStatus(c *gin.Context) {
	st := h.player.Status()
	if st.Now == nil {
		st.Now = &pkg.Song{}
	}
	c.JSON(http.StatusOK, st)
}
