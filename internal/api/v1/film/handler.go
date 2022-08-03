package film

import (
	"context"
	v1 "github.com/HalvaPovidlo/halvabot-go/internal/api/v1"
	"github.com/HalvaPovidlo/halvabot-go/internal/pkg/item"
	"github.com/gin-gonic/gin"
)

type FilmService interface {
	NewFilm(ctx context.Context, film *item.Film, userID string, withKP bool) (*item.Film, error)
	NewKinopoiskFilm(ctx context.Context, uri, userID string, score int) (*item.Film, error)
	EditFilm(ctx context.Context, film *item.Film) (*item.Film, error)
	AllFilms(ctx context.Context) ([]item.Film, error)
	Film(ctx context.Context, filmID string) (*item.Film, error)
	Comment(ctx context.Context, text, filmID, userID string) error
	Score(ctx context.Context, filmID, userID string, score int) (*item.Film, error)
}

type Handler struct {
	service FilmService
}

func NewHandler(service FilmService, superGroup *gin.RouterGroup) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) GetFilmsAll(c *gin.Context) {
	films, err := h.service.AllFilms(c)
	if err != nil {
		return
	}
}
func (h *Handler) PostFilmsKinopoisk(c *gin.Context) {

}
func (h *Handler) PostFilmsNew(c *gin.Context, params v1.PostFilmsNewParams) {

}
func (h *Handler) GetFilms(c *gin.Context, id v1.FilmId) {

}
func (h *Handler) PostFilmsId(c *gin.Context, id v1.FilmId) {

}
func (h *Handler) PostFilms(c *gin.Context, id v1.FilmId) {

}
func (h *Handler) PostFilmsIdScore(c *gin.Context, id v1.FilmId) {

}

