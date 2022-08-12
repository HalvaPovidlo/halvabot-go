package user

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
	EditUser(ctx context.Context, user *item.User) (*item.User, error)
	User(ctx context.Context, userID string) (*item.User, error)
	Films(ctx context.Context, userID string) (item.Films, error)
	Songs(ctx context.Context, userID string) ([]item.Song, error)
}

type handler struct {
	service internalService
}

func NewUserHandler(service internalService) *handler {
	return &handler{service: service}
}

func (h *handler) GetUserFilms(c *gin.Context, params v1.GetUserFilmsParams) {
	films, err := h.service.Films(c, c.GetString(login.UserID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, v1.Error{Msg: err.Error()})
		return
	}
	sortKey := v1.SortTitle
	if params.Sort == nil {
		sortKey = v1.Sort(*params.Sort)
	}
	films.Sort(v1.ConvertSortKet(sortKey))
	items := make([]v1.Film, 0, len(films))
	for i := range films {
		items = append(items, v1.ConvertFilm(&films[i]))
	}
	c.JSON(http.StatusOK, v1.Films{Items: items})
}

func (h *handler) GetUserInfo(c *gin.Context) {
	user, err := h.service.User(c, c.GetString(login.UserID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, v1.Error{Msg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, buildUser(user))
}

func (h *handler) PatchUserInfo(c *gin.Context) {
	var json v1.PatchUserInfoJSONBody
	err := c.ShouldBindJSON(&json)
	if err != nil {
		c.JSON(http.StatusBadRequest, v1.Error{Msg: err.Error()})
		return
	}
	user, err := h.service.EditUser(c, buildItem(&json))
	if err != nil {
		c.JSON(http.StatusInternalServerError, v1.Error{Msg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, buildUser(user))
}

func (h *handler) GetUserSongs(c *gin.Context) {
	songs, err := h.service.Songs(c, c.GetString(login.UserID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, v1.Error{Msg: err.Error()})
		return
	}
	items := make([]v1.Song, 0, len(songs))
	for i := range songs {
		items = append(items, *v1.BuildSong(&songs[i]))
	}
	c.JSON(http.StatusOK, v1.Songs{Items: items})
}

func buildUser(user *item.User) *v1.User {
	return &v1.User{
		Avatar:   pointer.ToString(user.Avatar),
		Username: user.Username,
	}
}

func buildItem(user *v1.User) *item.User {
	return &item.User{
		Avatar:   pointer.GetString(user.Avatar),
		Username: user.Username,
	}
}
