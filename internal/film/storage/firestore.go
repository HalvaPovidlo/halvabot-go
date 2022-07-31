package storage

import (
	"context"
	"github.com/pkg/errors"

	"cloud.google.com/go/firestore"
	"github.com/HalvaPovidlo/halvabot-go/internal/pkg/item"
	fire "github.com/HalvaPovidlo/halvabot-go/pkg/storage/firestore"
)

type Firestore struct {
	*firestore.Client
}

var ErrNoUserScore = errors.New("empty user score")

func (f *Firestore) NewFilm(ctx context.Context, film *item.Film, userID string) error {
	batch := f.Batch()
	batch.Create(f.Collection(fire.FilmsCollection).Doc(film.ID), film)
	if film.UserScore == nil {
		return ErrNoUserScore
	}
	batch.Create(f.Collection(fire.UsersCollection).Doc(userID).Collection(fire.FilmsCollection).Doc(film.ID), film)
	_, err := batch.Commit(ctx)
	return err
}

// WIP case of concurrent editing. We can mark changed fields.
func (f *Firestore) EditFilm(ctx context.Context, film *item.Film) error {
	_, err := f.Collection(fire.FilmsCollection).Doc(film.ID).Set(ctx, film)
	return err
}

func (f *Firestore) Score(ctx context.Context, filmID, userID string, score int) error {
	filmRef := f.Collection(fire.FilmsCollection).Doc(filmID)
	userRef := f.Collection(fire.UsersCollection).Doc(userID).Collection(fire.FilmsCollection).Doc(filmID)
	err := f.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(filmRef)
		if err != nil {
			return err
		}
		var film item.Film
		if err := doc.DataTo(&film); err != nil {
			return errors.Wrap(err, "parse film doc")
		}
		oldScore := film.Scores[userID]
		film.Score = score - oldScore
		film.Scores[userID] = score
		film.Average = float64(score) / float64(len(film.Scores))
		if err := tx.Set(filmRef, film); err != nil {
			return err
		}
		film.UserScore = &score
		return tx.Set(userRef, film)
	})
	return err
}
