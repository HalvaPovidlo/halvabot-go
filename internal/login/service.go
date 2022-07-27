package login

import (
	"context"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

const (
	UserID              = "user_id"
	authorizationHeader = "Authorization"
	bearerSchema        = "Bearer "
)

type accountsStorage interface {
	GetAccount(ctx context.Context, login string) (*AccountInfo, error)
}

type tokenizer interface {
	Generate(userID string) (string, error)
	Validate(encodedToken string) (*jwt.Token, error)
}

type Credentials struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type Service struct {
	accounts  accountsStorage
	tokenizer tokenizer
}

type Response struct {
	Token string `json:"token"`
}

func NewLoginService(loginService accountsStorage, jWtService tokenizer) *Service {
	return &Service{
		accounts:  loginService,
		tokenizer: jWtService,
	}
}

// Login godoc
// @summary Validates your login and password. Returns JWT.
// @accept  json
// @produce json
// @tags    auth
// @param   query body     Credentials true "Login and password"
// @success 200   {object} Response    "JWT of your session"
// @router  /login [post]
func (s *Service) LoginHandler(c *gin.Context) {
	var input Credentials
	err := c.ShouldBind(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrap(err, "no login or password").Error()})
		return
	}
	input.Login = strings.ToLower(input.Login)
	user, err := s.accounts.GetAccount(c, input.Login)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.Wrap(err, "failed to get account info").Error()})
		return
	}
	if user != nil && user.Password == input.Password {
		token, err := s.tokenizer.Generate(user.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errors.Wrap(err, "failed to generate token").Error()})
			return
		}
		c.JSON(http.StatusOK, Response{Token: token})
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
			if id, ok := claims[UserID]; ok {
				value, valid := id.(string)
				if !valid {
					c.AbortWithStatus(http.StatusInternalServerError)
				}
				c.Set(UserID, value)
			}
		} else {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}
