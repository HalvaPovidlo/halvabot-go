package login

import (
	"context"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

const (
	authorizationHeader = "Authorization"
	bearerSchema        = "Bearer "
)

type loginer interface {
	AuthUser(ctx context.Context, login, password string) (string, error)
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
	loginer   loginer
	tokenizer tokenizer
}

func NewLoginService(loginService loginer, jWtService tokenizer) *Service {
	return &Service{
		loginer:   loginService,
		tokenizer: jWtService,
	}
}

// Login godoc
// @summary Validates your login and password. Returns JWT.
// @accept  json
// @produce json
// @param   query body Credentials true "Login and password"
// @router  /login [post]
func (s *Service) LoginHandler(c *gin.Context) {
	var creds Credentials
	err := c.ShouldBind(&creds)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.Wrap(err, "no login or password").Error()})
		return
	}
	userID, err := s.loginer.AuthUser(c, creds.Login, creds.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.Wrap(err, "failed to validate creds").Error()})
		return
	}
	if userID != "" {
		token, err := s.tokenizer.Generate(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errors.Wrap(err, "failed to generate token").Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": token})
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
			if id, ok := claims["user_id"]; ok {
				value, valid := id.(string)
				if !valid {
					c.AbortWithStatus(http.StatusInternalServerError)
				}
				c.Set("user_id", value)
			}
		} else {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}
