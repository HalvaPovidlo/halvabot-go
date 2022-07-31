package login

import (
	"context"
	login2 "github.com/HalvaPovidlo/halvabot-go/internal/api/v1/login"
)

type mock struct{}

func NewMockStorage() *mock {
	return &mock{}
}

func (a *mock) GetAccount(ctx context.Context, login string) (*AccountInfo, error) {
	return &login.AccountInfo{
		Password:      "password",
		login2.UserID: "user_id",
	}, nil
}
