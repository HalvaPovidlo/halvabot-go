package storage

import (
	"context"

	"cloud.google.com/go/firestore"

	"github.com/HalvaPovidlo/halvabot-go/internal/pkg/item"
)

type Service struct {
	storage *Firestore
	cache   *Cache
}

func NewStorage(client *firestore.Client) *Service {
	return &Service{
		storage: NewFirestore(client),
		cache:   NewCache(),
	}
}

func (s *Service) EditUser(ctx context.Context, user *item.User) error {
	err := s.storage.EditUser(ctx, user)
	if err != nil {
		return err
	}
	s.cache.SetUser(*user)
	return nil
}

func (s *Service) User(ctx context.Context, userID string) (*item.User, error) {
	if cached := s.cache.User(userID); cached != nil {
		return cached, nil
	}
	user, err := s.storage.User(ctx, userID)
	if err != nil {
		return nil, err
	}
	s.cache.SetUser(*user)
	return user, nil
}

func (s *Service) Films(ctx context.Context, userID string) (item.Films, error) {
	films, err := s.storage.Films(ctx, userID)
	if err != nil {
		return nil, err
	}
	return films, nil
}

func (s *Service) Songs(ctx context.Context, userID string) ([]item.Song, error) {
	songs, err := s.storage.Songs(ctx, userID)
	if err != nil {
		return nil, err
	}
	return songs, nil
}
