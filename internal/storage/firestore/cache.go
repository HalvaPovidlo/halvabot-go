package firestore

import (
	"os"
	"sync"
	"time"

	"github.com/HalvaPovidlo/discordBotGo/internal/pkg"
	"github.com/HalvaPovidlo/discordBotGo/pkg/contexts"
)

type Item struct {
	song    pkg.Song
	updated time.Time
}

type CacheKey string

type SongsCache struct {
	sync.Mutex
	songs map[string]Item
}

func NewSongsCache(ctx contexts.Context, expirationTime time.Duration) *SongsCache {
	c := &SongsCache{
		songs: make(map[string]Item),
	}
	c.expireProcess(ctx, expirationTime)
	return c
}

func (c *SongsCache) Get(k string) (*pkg.Song, bool) {
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

func (c *SongsCache) Set(k string, song *pkg.Song) {
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

func (c *SongsCache) KeyFromID(s pkg.SongID) string {
	return s.String()
}

func (c *SongsCache) expireProcess(ctx contexts.Context, expirationTime time.Duration) {
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
						// If the song has not played before the expirationTime passed, there will be an error because
						// we will delete it. But since the time is very long, we score we don't really care about such case.
						_ = os.Remove(v.song.StreamURL)
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
