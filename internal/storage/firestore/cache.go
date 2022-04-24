package firestore

import (
	"time"

	"github.com/patrickmn/go-cache"

	"github.com/HalvaPovidlo/discordBotGo/internal/storage"
)

type CacheKey string

type SongsCache struct {
	cache *cache.Cache
}

func NewSongsCache(expirationTime, cleanUpTime time.Duration) *SongsCache {
	return &SongsCache{
		cache: cache.New(expirationTime, cleanUpTime),
	}
}

func (c *SongsCache) Get(k CacheKey) (*storage.Song, bool) {
	if v, ok := c.cache.Get(string(k)); !ok {
		s, ok := v.(*storage.Song)
		return s, ok
	}
	return nil, false
}

func (c *SongsCache) Set(k CacheKey, song *storage.Song) {
	c.cache.Set(string(k), song, cache.DefaultExpiration)
}

func (c *SongsCache) KeyFromID(s storage.SongID) CacheKey {
	return CacheKey(s.String())
}

func (c *SongsCache) Items() []*storage.Song {
	items := c.cache.Items()
	songs := make([]*storage.Song, 0, len(items))
	for _, v := range items {
		songs = append(songs, v.Object.(*storage.Song))
	}
	return songs
}
