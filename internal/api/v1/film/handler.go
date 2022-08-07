//nolint:revive // generated interface
package film

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	v1 "github.com/HalvaPovidlo/halvabot-go/internal/api/v1"
	"github.com/HalvaPovidlo/halvabot-go/internal/login"
	"github.com/HalvaPovidlo/halvabot-go/internal/pkg/item"
)

type internalService interface {
	NewFilm(ctx context.Context, film *item.Film, userID string, withKP bool) (*item.Film, error)
	NewKinopoiskFilm(ctx context.Context, uri, userID string, score int) (*item.Film, error)
	EditFilm(ctx context.Context, film *item.Film) (*item.Film, error)
	AllFilms(ctx context.Context) ([]item.Film, error)
	Film(ctx context.Context, filmID string) (*item.Film, error)
	Comment(ctx context.Context, text, filmID, userID string) error
	Score(ctx context.Context, filmID, userID string, score int) (*item.Film, error)
}

type Films struct {
	Items []item.Film `json:"items"`
}

type Handler struct {
	service internalService
}

func NewFilmHandler(service internalService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) GetFilmsAll(c *gin.Context) {
	films, err := h.service.AllFilms(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, v1.Error{Msg: err.Error()})
		return
	}
	items := make([]v1.Film, 0, len(films))
	for i := range films {
		items = append(items, v1.ConvertFilm(setUserScore(&films[i], c.GetString(login.UserID))))
	}
	c.JSON(http.StatusOK, v1.Films{Items: items})
}

func (h *Handler) PostFilmsKinopoisk(c *gin.Context) {
	var json v1.PostFilmsKinopoiskJSONRequestBody
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, v1.Error{Msg: err.Error()})
		return
	}
	film, err := h.service.NewKinopoiskFilm(c, json.Url, c.GetString(login.UserID), json.Score)
	if err != nil {
		c.JSON(http.StatusInternalServerError, v1.Error{Msg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, v1.ConvertFilm(setUserScore(film, c.GetString(login.UserID))))
}

func (h *Handler) PostFilmsNew(c *gin.Context, params v1.PostFilmsNewParams) {
	var json v1.PostFilmsNewJSONRequestBody
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, v1.Error{Msg: err.Error()})
		return
	}
	useKinopoisk := params.Kinopoisk == nil || *params.Kinopoisk
	film, err := h.service.NewFilm(c, v1.ConvertItem(&json), c.GetString(login.UserID), useKinopoisk)
	if err != nil {
		c.JSON(http.StatusInternalServerError, v1.Error{Msg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, v1.ConvertFilm(film))
}

func (h *Handler) GetFilms(c *gin.Context, id v1.FilmId) {
	film, err := h.service.Film(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, v1.Error{Msg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, v1.ConvertFilm(setUserScore(film, c.GetString(login.UserID))))
}

func (h *Handler) PatchFilmsId(c *gin.Context, id v1.FilmId) {
	var json v1.PatchFilmsIdJSONRequestBody
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, v1.Error{Msg: err.Error()})
		return
	}
	json.FilmId = id
	film, err := h.service.EditFilm(c, v1.ConvertItem(&json))
	if err != nil {
		c.JSON(http.StatusInternalServerError, v1.Error{Msg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, v1.ConvertFilm(setUserScore(film, c.GetString(login.UserID))))
}

func (h *Handler) PostFilmsIdComment(c *gin.Context, id v1.FilmId) {
	var json v1.PostFilmsIdCommentJSONRequestBody
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, v1.Error{Msg: err.Error()})
		return
	}
	err := h.service.Comment(c, json.Text, id, c.GetString(login.UserID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, v1.Error{Msg: err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (h *Handler) PostFilmsIdScore(c *gin.Context, id v1.FilmId) {
	var json v1.PostFilmsIdScoreJSONRequestBody
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, v1.Error{Msg: err.Error()})
		return
	}
	if json.Score < -1 || json.Score > 1 {
		c.JSON(http.StatusBadRequest, v1.Error{Msg: "score must be an integer in the range [-1,1]"})
		return
	}
	film, err := h.service.Score(c, id, c.GetString(login.UserID), json.Score)
	if err != nil {
		c.JSON(http.StatusInternalServerError, v1.Error{Msg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, v1.ConvertFilm(setUserScore(film, c.GetString(login.UserID))))
}

func setUserScore(film *item.Film, userID string) *item.Film {
	film.UserScore = nil
	if userID == "" {
		return film
	}
	if score, ok := film.Scores[userID]; ok {
		film.UserScore = &score
	}
	return film
}
