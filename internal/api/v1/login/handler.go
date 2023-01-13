package login

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	v1 "github.com/HalvaPovidlo/halvabot-go/internal/api/v1"
	"github.com/HalvaPovidlo/halvabot-go/internal/login"
)

const (
	authorizationHeader = "Authorization"
)

type loginService interface {
	GetUserID(tokenString string) (string, error)
	LoginToken(ctx context.Context, login, password string) (string, error)
}

type Handler struct {
	loginer loginService
}

func NewLoginHandler(loginService loginService) *Handler {
	return &Handler{
		loginer: loginService,
	}
}

func (h *Handler) PostAuthToken(c *gin.Context) {
	var input v1.PostAuthTokenJSONRequestBody
	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, v1.Error{Msg: err.Error()})
		return
	}
	token, err := h.loginer.LoginToken(c, input.Login, input.Password)
	switch {
	case errors.Is(err, login.ErrUnauthorized):
		c.JSON(http.StatusUnauthorized, v1.Error{Msg: "invalid login or password"})
	case err != nil:
		c.JSON(http.StatusInternalServerError, v1.Error{Msg: err.Error()})
	default:
		c.JSON(http.StatusOK, v1.Token{Token: token})
	}
}

func (h *Handler) Authorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(authorizationHeader)
		if len(authHeader) < len(login.BearerSchema)+1 {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		id, err := h.loginer.GetUserID(authHeader[len(login.BearerSchema):])
		switch {
		case errors.Is(err, login.ErrUnauthorized):
			c.AbortWithStatus(http.StatusUnauthorized)
		case err != nil:
			c.AbortWithStatusJSON(http.StatusInternalServerError, v1.Error{Msg: err.Error()})
		default:
			c.Set(login.UserID, id)
		}
	}
}

func (h *Handler) SoftAuthorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(authorizationHeader)
		if len(authHeader) < len(login.BearerSchema)+1 {
			return
		}
		id, err := h.loginer.GetUserID(authHeader[len(login.BearerSchema):])
		if err == nil && id != "" {
			c.Set(login.UserID, id)
		}
	}
}

func (h *Handler) HardAuthorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString(login.UserID) == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}
