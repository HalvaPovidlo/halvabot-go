package rest

import (
	"github.com/HalvaPovidlo/halvabot-go/internal/pkg/item"
	"github.com/gin-gonic/gin"
)

type FilmService interface {
	Add(film *item.Film) (string, error)
	Comment(filmID string, userID string, comment string)
	Score(filmID string, userID string, score int)
	GetFilm(filmID string) *item.Film
	//GetFilms(code film.SortCode) []film.ShortFilm
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
	//music.POST("/add", h.addHandler)
	//music.POST("/comment", h.commentHandler)
	//music.POST("/score", h.commentHandler)
	//music.POST("/get", h.getHandler)
	//music.GET("/all", h.allHandler)
	return music
}

type Response struct {
	Message string `json:"message"`
}
