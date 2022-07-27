package login

import (
	"context"
)

type mock struct {
}

func NewMockAuthenticator() *mock {
	return &mock{}
}

func (a *mock) AuthUser(ctx context.Context, login, password string) (string, error) {
	return "012345678901234567", nil
}
