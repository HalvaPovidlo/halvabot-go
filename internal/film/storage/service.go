package storage

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/HalvaPovidlo/halvabot-go/internal/pkg/item"
)

type Service struct {
	client *firestore.Client
}

func (s *Service) Add(ctx context.Context, film item.Film, userID string) {
	_, err := s.client.Collection("films").Doc(string(film.ID)).Set(ctx, film)
	if err != nil {
		return
	}
}
