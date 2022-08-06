package login

import (
	"context"
)

type mock struct{}

func NewMockStorage() *mock {
	return &mock{}
}

func (a *mock) GetAccount(ctx context.Context, login string) (*AccountInfo, error) {
	return &AccountInfo{
		Password: "password",
		UserID:   "mockman_id",
	}, nil
}
