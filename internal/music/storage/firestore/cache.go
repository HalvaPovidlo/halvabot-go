package firestore

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/HalvaPovidlo/halvabot-go/internal/pkg/item"
)

type Item struct {
	song    item.Song
	updated time.Time
}

type CacheKey string

// SongsCache TODO: there is no need in this cache because firestore.client has it's own that can be reused
type SongsCache struct {
	sync.Mutex
	songs map[string]Item
}

func NewSongsCache(ctx context.Context, expirationTime time.Duration) *SongsCache {
	c := &SongsCache{
		songs: make(map[string]Item),
	}
	c.expireProcess(ctx, expirationTime)
	return c
}

func (c *SongsCache) Get(k string) (*item.Song, bool) {
	c.Lock()
	defer c.Unlock()
	s, ok := c.songs[k]
	if !ok {
		return nil, false
	}
	s.updated = time.Now()
	c.songs[k] = s
	return &s.song, true
}

func (c *SongsCache) Set(k string, song *item.Song) {
	if song == nil {
		return
	}
	c.Lock()
	c.songs[k] = Item{
		song:    *song,
		updated: time.Now(),
	}
	c.Unlock()
}

func (c *SongsCache) KeyFromID(s item.SongID) string {
	return s.String()
}

func (c *SongsCache) expireProcess(ctx context.Context, expirationTime time.Duration) {
	ticker := time.NewTicker(expirationTime)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				c.Lock()
				now := time.Now()
				for k, v := range c.songs {
					if v.updated.Before(now.Add(-expirationTime)) {
						delete(c.songs, k)
					}
				}
				c.Unlock()
			}
		}
	}()
}

func (c *SongsCache) Clear() {
	c.Lock()
	for k, v := range c.songs {
		_ = os.Remove(v.song.StreamURL)
		delete(c.songs, k)
	}
	c.Unlock()
}
