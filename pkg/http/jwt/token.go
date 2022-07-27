package jwt

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

const (
	tokenExpiration = time.Hour * 24
)

type service struct {
	secretKey string
}

type Claims struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

func NewJWTokenizer(secretKey string) *service {
	if secretKey == "" {
		secretKey = "secret"
	}
	return &service{
		secretKey: secretKey,
	}
}

func (s *service) Generate(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(tokenExpiration).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	})

	t, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", err
	}
	return t, nil
}

func (s *service) Validate(encodedToken string) (*jwt.Token, error) {
	return jwt.Parse(encodedToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("token is invalid")
		}
		return []byte(s.secretKey), nil
	})
}
