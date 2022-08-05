package login

import (
	"context"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	v1 "github.com/HalvaPovidlo/halvabot-go/internal/api/v1"
	"github.com/HalvaPovidlo/halvabot-go/internal/login"
)

const (
	authorizationHeader = "Authorization"
	bearerSchema        = "Bearer "
)

type accountsStorage interface {
	GetAccount(ctx context.Context, login string) (*login.AccountInfo, error)
}

type tokenizer interface {
	Generate(userID string) (string, error)
	Validate(encodedToken string) (*jwt.Token, error)
}

type Service struct {
	accounts  accountsStorage
	tokenizer tokenizer
}

func NewLoginHandler(loginService accountsStorage, jWtService tokenizer) *Service {
	return &Service{
		accounts:  loginService,
		tokenizer: jWtService,
	}
}

func (s *Service) PostAuthToken(c *gin.Context) {
	var input v1.PostAuthTokenJSONRequestBody
	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, v1.Error{Msg: errors.Wrap(err, "no login or password").Error()})
		return
	}
	input.Login = strings.ToLower(input.Login)
	user, err := s.accounts.GetAccount(c, input.Login)
	if err != nil {
		c.JSON(http.StatusInternalServerError, v1.Error{Msg: errors.Wrap(err, "failed to get account info").Error()})
		return
	}
	if user != nil && user.Password == input.Password {
		token, err := s.tokenizer.Generate(user.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, v1.Error{Msg: errors.Wrap(err, "failed to generate token").Error()})
			return
		}
		c.JSON(http.StatusOK, v1.Token{Token: token})
		return
	}
	c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid login or password"})
}

func (s *Service) Authorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(authorizationHeader)
		if len(authHeader) < len(bearerSchema)+1 {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenString := authHeader[len(bearerSchema):]
		token, err := s.tokenizer.Validate(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if token.Valid {
			claims := token.Claims.(jwt.MapClaims)
			if id, ok := claims[login.UserID]; ok {
				value, valid := id.(string)
				if !valid {
					c.AbortWithStatus(http.StatusInternalServerError)
				}
				c.Set(login.UserID, value)
			}
		} else {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}
