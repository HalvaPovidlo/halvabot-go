package storage

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/pkg/errors"

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

func (s *Service) NewFilm(ctx context.Context, film *item.Film, userID string) error {
	err := s.storage.NewFilm(ctx, film, userID)
	if err != nil {
		return errors.Wrap(err, "add new film to firestore")
	}
	film.WithComments = true
	s.cache.NewFilm(*film)
	return nil
}

func (s *Service) EditFilm(ctx context.Context, film *item.Film) error {
	err := s.storage.EditFilm(ctx, film)
	if err != nil {
		return errors.Wrap(err, "edit film in firestore")
	}
	s.cache.EditFilm(*film)
	return nil
}

func (s *Service) AllFilms(ctx context.Context) ([]item.Film, error) {
	films := s.cache.AllFilms()
	if len(films) == 0 {
		all, err := s.storage.AllFilms(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "get all films from firestore")
		}
		s.cache.Fill(all)
		films = all
	}
	return films, nil
}

func (s *Service) Film(ctx context.Context, filmID string) (*item.Film, error) {
	cached := s.cache.Film(filmID)
	if cached != nil || cached.WithComments {
		return cached, nil
	}
	film, err := s.storage.Film(ctx, filmID)
	if err != nil {
		return nil, errors.Wrap(err, "get film from firestore")
	}
	s.cache.SetFilm(*film)
	return film, nil
}

func (s *Service) Comment(ctx context.Context, filmID string, comment *item.Comment) error {
	id, err := s.storage.Comment(ctx, filmID, comment)
	if err != nil {
		return errors.Wrap(err, "add comment to the film in firestore")
	}
	s.cache.Comment(filmID, id, *comment)
	return nil
}

func (s *Service) Score(ctx context.Context, filmID, userID string, score int) (*item.Film, error) {
	film, err := s.storage.Score(ctx, filmID, userID, score)
	if err != nil {
		return nil, errors.Wrap(err, "add score to the film in firestore")
	}
	s.cache.EditFilm(film)
	return &film, nil
}
