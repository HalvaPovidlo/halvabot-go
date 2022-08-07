package user

import (
	"context"

	"github.com/pkg/errors"

	"github.com/HalvaPovidlo/halvabot-go/internal/pkg/item"
)

type userStorage interface {
	EditUser(ctx context.Context, user *item.User) error
	User(ctx context.Context, userID string) (*item.User, error)
	Films(ctx context.Context, userID string) ([]item.Film, error)
	Songs(ctx context.Context, userID string) ([]item.Song, error)
}

type service struct {
	storage userStorage
}

func NewUserService(storage userStorage) *service {
	return &service{storage: storage}
}

func (s *service) EditUser(ctx context.Context, user *item.User) (*item.User, error) {
	err := s.storage.EditUser(ctx, user)
	if err != nil {
		return nil, errors.Wrap(err, "edit user from storage")
	}
	return user, nil
}

func (s *service) User(ctx context.Context, userID string) (*item.User, error) {
	user, err := s.storage.User(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "get user from storage")
	}
	return user, nil
}

func (s *service) Films(ctx context.Context, userID string) ([]item.Film, error) {
	films, err := s.storage.Films(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "get user films from storage")
	}
	return films, nil
}

func (s *service) Songs(ctx context.Context, userID string) ([]item.Song, error) {
	songs, err := s.storage.Songs(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "get user songs from storage")
	}
	return songs, nil
}
