package film

import (
	"context"
	"net/http"

	"github.com/AlekSi/pointer"
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
		items = append(items, convertFilm(setUserScore(&films[i], c.GetString(login.UserID))))
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
	c.JSON(http.StatusOK, convertFilm(setUserScore(film, c.GetString(login.UserID))))
}

func (h *Handler) PostFilmsNew(c *gin.Context, params v1.PostFilmsNewParams) {
	var json v1.PostFilmsNewJSONRequestBody
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, v1.Error{Msg: err.Error()})
		return
	}
	useKinopoisk := params.Kinopoisk == nil || *params.Kinopoisk == true
	film, err := h.service.NewFilm(c, convertItem(&json), c.GetString(login.UserID), useKinopoisk)
	if err != nil {
		c.JSON(http.StatusInternalServerError, v1.Error{Msg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, convertFilm(film))
}

func (h *Handler) GetFilms(c *gin.Context, id v1.FilmId) {
	film, err := h.service.Film(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, v1.Error{Msg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, convertFilm(setUserScore(film, c.GetString(login.UserID))))
}

func (h *Handler) PostFilmsId(c *gin.Context, id v1.FilmId) {
	var json v1.PostFilmsIdJSONRequestBody
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, v1.Error{Msg: err.Error()})
		return
	}
	json.FilmId = id
	film, err := h.service.EditFilm(c, convertItem(&json))
	if err != nil {
		c.JSON(http.StatusInternalServerError, v1.Error{Msg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, convertFilm(setUserScore(film, c.GetString(login.UserID))))
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
	film, err := h.service.Score(c, id, c.GetString(login.UserID), json.Score)
	if err != nil {
		c.JSON(http.StatusInternalServerError, v1.Error{Msg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, convertFilm(setUserScore(film, c.GetString(login.UserID))))
}

func convertFilm(f *item.Film) v1.Film {
	comments := make(map[string]interface{})
	for k, v := range f.Comments {
		comments[k] = v
	}
	scores := make(map[string]interface{})
	for k, v := range f.Scores {
		scores[k] = v
	}
	return v1.Film{
		Average:                  float32(f.Average),
		Comments:                 &comments,
		Cover:                    pointer.ToStringOrNil(f.Cover),
		Description:              pointer.ToStringOrNil(f.Description),
		Director:                 pointer.ToStringOrNil(f.Director),
		Duration:                 pointer.ToStringOrNil(f.Duration),
		FilmId:                   f.ID,
		FilmLength:               pointer.ToIntOrNil(f.FilmLength),
		Genres:                   &f.Genres,
		Kinopoisk:                f.URL,
		Poster:                   pointer.ToStringOrNil(f.Poster),
		RatingImdb:               float32(f.RatingImdb),
		RatingImdbVoteCount:      f.RatingImdbVoteCount,
		RatingKinopoisk:          float32(f.RatingKinopoisk),
		RatingKinopoiskVoteCount: f.RatingKinopoiskVoteCount,
		Score:                    f.Score,
		Scores:                   &scores,
		Serial:                   f.Serial,
		ShortFilm:                f.ShortFilm,
		Title:                    f.Title,
		TitleOriginal:            pointer.ToStringOrNil(f.TitleOriginal),
		UserScore:                f.UserScore,
		Year:                     pointer.ToIntOrNil(f.Year),
	}
}

func convertItem(f *v1.Film) *item.Film {
	scores := make(map[string]int)
	if f.Scores != nil && len(*f.Scores) != 0 {
		for k, v := range *f.Scores {
			if value, ok := v.(int); ok {
				scores[k] = value
			}
		}
	}
	var genres []string
	if f.Genres != nil {
		genres = *f.Genres
	}
	return &item.Film{
		ID:                       f.FilmId,
		Title:                    f.Title,
		TitleOriginal:            pointer.GetString(f.TitleOriginal),
		Poster:                   pointer.GetString(f.Poster),
		Cover:                    pointer.GetString(f.Cover),
		Director:                 pointer.GetString(f.Director),
		Description:              pointer.GetString(f.Description),
		Duration:                 pointer.GetString(f.Duration),
		Score:                    f.Score,
		UserScore:                f.UserScore,
		Average:                  float64(f.Average),
		Scores:                   scores,
		WithComments:             false,
		URL:                      f.Kinopoisk,
		RatingKinopoisk:          float64(f.RatingKinopoisk),
		RatingKinopoiskVoteCount: f.RatingKinopoiskVoteCount,
		RatingImdb:               float64(f.RatingImdb),
		RatingImdbVoteCount:      f.RatingImdbVoteCount,
		Year:                     pointer.GetInt(f.Year),
		FilmLength:               pointer.GetInt(f.FilmLength),
		Serial:                   f.Serial,
		ShortFilm:                f.ShortFilm,
		Genres:                   genres,
	}
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
