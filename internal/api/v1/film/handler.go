package film

import (
	"github.com/gin-gonic/gin"

	"github.com/HalvaPovidlo/halvabot-go/internal/pkg/item"
)

type FilmService interface {
	Add(film *item.Film) (string, error)
	Comment(filmID string, userID string, comment string)
	Score(filmID string, userID string, score int)
	GetFilm(filmID string) *item.Film
}

// Handler TODO: Auth
type Handler struct {
	service FilmService
	super   *gin.RouterGroup
}

func NewHandler(service FilmService, superGroup *gin.RouterGroup) *Handler {
	return &Handler{
		service: service,
		super:   superGroup,
	}
}

func (h *Handler) Router() *gin.RouterGroup {
	music := h.super.Group("/film")
	return music
}

type Response struct {
	Message string `json:"message"`
}
