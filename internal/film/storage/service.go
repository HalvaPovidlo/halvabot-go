package storage

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/HalvaPovidlo/halvabot-go/internal/pkg/item"
	fire "github.com/HalvaPovidlo/halvabot-go/pkg/storage/firestore"
)

type Service struct {
	Client *firestore.Client
}

func (s *Service) Add(ctx context.Context, film item.Film, userID string) error {
	filmID := string(film.ID)
	filmDoc := s.Client.Collection(fire.FilmsCollection).Doc(filmID)

	batch := s.Client.Batch()
	batch.Set(filmDoc, film)
	batch.Set(filmDoc.Collection(fire.ScoresCollection).Doc(userID), item.FireScore{Score: film.UserScore})
	batch.Set(s.Client.Collection(fire.UsersCollection).Doc(userID).Collection(fire.FilmsCollection).Doc(filmID), film)

	_, err := batch.Commit(ctx)
	return err
}
