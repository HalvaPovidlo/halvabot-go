package storage

import (
	"sync"

	"github.com/HalvaPovidlo/halvabot-go/internal/pkg/item"
)

type Cache struct {
	sync.Mutex
	films map[string]item.Film
}

func NewCache() *Cache {
	return &Cache{
		films: make(map[string]item.Film),
	}
}

func (c *Cache) Fill(films item.Films) {
	c.Lock()
	for i := range films {
		c.films[films[i].ID] = films[i]
	}
	c.Unlock()
}

func (c *Cache) NewFilm(film item.Film) {
	c.Lock()
	if _, ok := c.films[film.ID]; !ok {
		c.films[film.ID] = film
	}
	c.Unlock()
}

func (c *Cache) EditFilm(film item.Film) {
	c.Lock()
	if old, ok := c.films[film.ID]; ok && old.WithComments {
		film.Comments = old.Comments
		film.WithComments = true
	}
	c.films[film.ID] = film
	c.Unlock()
}

func (c *Cache) SetFilm(film item.Film) {
	c.Lock()
	c.films[film.ID] = film
	c.Unlock()
}

func (c *Cache) AllFilms() item.Films {
	c.Lock()
	var films item.Films
	films = make(item.Films, 0, len(c.films))
	for _, f := range c.films {
		films = append(films, f)
	}
	c.Unlock()
	return films
}

func (c *Cache) Film(filmID string) *item.Film {
	c.Lock()
	defer c.Unlock()
	if film, ok := c.films[filmID]; ok {
		return &film
	}
	return nil
}

func (c *Cache) Comments(filmID string) map[string]item.Comment {
	c.Lock()
	defer c.Unlock()
	film, ok := c.films[filmID]
	if !ok {
		return nil
	}
	return film.Comments
}

func (c *Cache) Comment(filmID, commentID string, comment item.Comment) {
	c.Lock()
	if film, ok := c.films[filmID]; ok {
		if film.Comments == nil {
			film.Comments = make(map[string]item.Comment)
		}
		film.Comments[commentID] = comment
	}
	c.Unlock()
}
