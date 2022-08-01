package film

import (
	"context"
	"github.com/pkg/errors"
	"time"

	"github.com/HalvaPovidlo/halvabot-go/internal/pkg/item"
)

type storage interface {
	NewFilm(ctx context.Context, film *item.Film, userID string) error
	EditFilm(ctx context.Context, film *item.Film) error
	AllFilms(ctx context.Context) ([]item.Film, error)
	Film(ctx context.Context, filmID string) (*item.Film, error)
	Comment(ctx context.Context, filmID string, comment *item.Comment) error
	Score(ctx context.Context, filmID, userID string, score int) (*item.Film, error)
}

type Service struct {
	storage   storage
	kinopoisk *Kinopoisk
}

var ErrNoUserScore = errors.New("no user score")

func (s Service) NewFilm(ctx context.Context, film *item.Film, userID string) (*item.Film, error) {
	if film.UserScore == nil {
		return nil, ErrNoUserScore
	}
	film.Score = *film.UserScore
	film.Scores[userID] = *film.UserScore
	err := s.storage.NewFilm(ctx, film, userID)
	if err != nil {
		return nil, errors.Wrap(err, "new film to storage")
	}
	return film, nil
}

func (s Service) EditFilm(ctx context.Context, film *item.Film) (*item.Film, error) {
	err := s.storage.EditFilm(ctx, film)
	if err != nil {
		return nil, errors.Wrap(err, "edit film in storage")
	}
	return film, nil
}

func (s Service) AllFilms(ctx context.Context) ([]item.Film, error) {
	films, err := s.storage.AllFilms(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "all films from storage")
	}
	return films, nil
}

func (s Service) Film(ctx context.Context, filmID string) (*item.Film, error) {
	film, err := s.storage.Film(ctx, filmID)
	if err != nil {
		return nil, errors.Wrap(err, "get film from storage")
	}
	return film, nil
}

func (s Service) Comment(ctx context.Context, text, filmID, userID string) error {
	comment := &item.Comment{
		UserID:    userID,
		Text:      text,
		CreatedAt: time.Now(),
	}
	err := s.storage.Comment(ctx, filmID, comment)
	if err != nil {
		return errors.Wrap(err, "add comment to storage")
	}
	return nil
}

func (s Service) Score(ctx context.Context, filmID, userID string, score int) (*item.Film, error) {
	film, err := s.storage.Score(ctx, filmID, userID, score)
	if err != nil {
		return nil, errors.Wrap(err, "score film in storage")
	}
	return film, nil
}

//
//func setUserScore(film *item.Film, userID string) *item.Film {
//	film.UserScore = nil
//	if score, ok := film.Scores[userID]; ok {
//		film.UserScore = &score
//	}
//	return film
//}
