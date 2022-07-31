package storage

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"

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

func (f *Firestore) AllFilms(ctx context.Context, userID string) ([]item.Film, error) {
	var films []item.Film
	iter := f.Collection(fire.FilmsCollection).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		film, err := parseFilm(doc, userID)
		if err != nil {
			return nil, err
		}
		films = append(films, *film)
	}
	return films, nil
}

func (f *Firestore) Film(ctx context.Context, filmID, userID string) (*item.Film, error) {
	doc, err := f.Collection(fire.FilmsCollection).Doc(filmID).Get(ctx)
	if err != nil {
		return nil, err
	}
	film, err := parseFilm(doc, userID)
	if err != nil {
		return nil, err
	}
	comments, err := f.Comments(ctx, filmID)
	if err != nil {
		return nil, errors.Wrapf(err, "get comments film=%s", filmID)
	}
	film.Comments = comments
	return film, nil
}

func parseFilm(doc *firestore.DocumentSnapshot, userID string) (*item.Film, error) {
	var film item.Film
	if err := doc.DataTo(&film); err != nil {
		return nil, errors.Wrap(err, "parse film doc")
	}
	film.ID = doc.Ref.ID
	if userID != "" {
		if userScore, ok := film.Scores[userID]; ok {
			film.UserScore = &userScore
		}
	}
	return &film, nil
}

func (f *Firestore) Comments(ctx context.Context, filmID string) (map[string]item.Comment, error) {
	var comments map[string]item.Comment
	iter := f.Collection(fire.FilmsCollection).Doc(filmID).Collection(fire.CommentsCollection).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var comment item.Comment
		if err := doc.DataTo(&comment); err != nil {
			return nil, errors.Wrap(err, "parse comment doc")
		}
		comments[doc.Ref.ID] = comment
	}
	return comments, nil
}

func (f *Firestore) Comment(ctx context.Context, text, filmID, userID string) (string, error) {
	comment := item.Comment{
		UserID:    userID,
		Text:      text,
		CreatedAt: time.Now(),
	}
	doc, _, err := f.Collection(fire.FilmsCollection).Doc(filmID).Collection(fire.CommentsCollection).Add(ctx, comment)
	if err != nil {
		return "", err
	}
	return doc.ID, nil
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
