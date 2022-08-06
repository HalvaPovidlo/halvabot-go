package login

import (
	"context"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

const (
	UserID       = "user_id"
	BearerSchema = "Bearer "
)

var ErrUnauthorized = errors.New("user unauthorized")

type accountStorage interface {
	GetAccount(ctx context.Context, login string) (*AccountInfo, error)
}

type tokenizer interface {
	Generate(userID string) (string, error)
	Validate(encodedToken string) (*jwt.Token, error)
}

type Service struct {
	accounts  accountStorage
	tokenizer tokenizer
}

func NewLoginService(s accountStorage, t tokenizer) *Service {
	return &Service{
		accounts:  s,
		tokenizer: t,
	}
}

func (s *Service) LoginToken(ctx context.Context, login, password string) (string, error) {
	login = strings.ToLower(login)
	user, err := s.accounts.GetAccount(ctx, login)
	if err != nil {
		return "", errors.Wrap(err, "get account info")
	}
	if user != nil && user.Password == password {
		token, err := s.tokenizer.Generate(user.UserID)
		if err != nil {
			return "", errors.Wrap(err, "generate token")
		}
		return token, nil
	}
	return "", ErrUnauthorized
}

func (s *Service) GetUserID(tokenString string) (string, error) {
	token, err := s.tokenizer.Validate(tokenString)
	if err != nil {
		return "", errors.Wrap(err, "validate token")
	}
	if token.Valid {
		claims := token.Claims.(jwt.MapClaims)
		if id, ok := claims[UserID]; ok {
			if value, ok := id.(string); ok {
				return value, nil
			}
		}
	}
	return "", ErrUnauthorized
}
