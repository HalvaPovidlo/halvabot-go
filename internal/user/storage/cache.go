package storage

import (
	"sync"

	"github.com/HalvaPovidlo/halvabot-go/internal/pkg/item"
)

// TODO: cache user films and songs (problem with rate and play)
type Cache struct {
	sync.Mutex
	users map[string]item.User
}

func NewCache() *Cache {
	return &Cache{
		users: make(map[string]item.User),
	}
}

func (c *Cache) SetUser(user item.User) {
	c.Lock()
	c.users[user.ID] = user
	c.Unlock()
}

func (c *Cache) User(userID string) *item.User {
	c.Lock()
	defer c.Unlock()
	if film, ok := c.users[userID]; ok {
		return &film
	}
	return nil
}
