package firestore

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/HalvaPovidlo/discordBotGo/internal/pkg"
	"github.com/HalvaPovidlo/discordBotGo/pkg/contexts"
)

type shortCache struct {
	sync.RWMutex
	List []pkg.SongID
}

type Service struct {
	cache  *SongsCache
	client *Client

	songsShort   shortCache
	updatesMutex sync.Mutex
	updated      bool
}

func NewFirestoreService(ctx context.Context, client *Client, songs *SongsCache) (*Service, error) {
	f := Service{
		cache:      songs,
		client:     client,
		songsShort: shortCache{},
	}
	go f.updateShortCache(ctx)
	f.updateShortCacheProcess(ctx)
	return &f, nil
}

func (s *Service) GetSong(ctx context.Context, id pkg.SongID) (*pkg.Song, error) {
	key := s.cache.KeyFromID(id)
	logger := contexts.GetLogger(ctx)
	logger.Debug("get song from cache", zap.String("id", id.String()))
	if s, ok := s.cache.Get(key); ok {
		return s, nil
	}

	logger.Debug("get song from db", zap.String("id", id.String()))

	song, err := s.client.GetSongByID(ctx, id)
	if err != nil {
		if err == ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, errors.Wrapf(err, "get song by id %s", id)
	}
	logger.Debug("set song to cache", zap.String("id", id.String()))
	s.cache.Set(key, song)
	return song, nil
}

func (s *Service) SetSong(ctx context.Context, song *pkg.Song) error {
	s.setUpdate(true)
	if err := s.client.SetSong(ctx, song); err != nil {
		return errors.Wrap(err, "firestore set song")
	}
	s.cache.Set(s.cache.KeyFromID(song.ID), song)
	return nil
}

func (s *Service) UpsertSongIncPlaybacks(ctx context.Context, new *pkg.Song) (int, error) {
	old, err := s.GetSong(ctx, new.ID)
	if err != nil && err != ErrNotFound {
		return 0, errors.Wrap(err, "failed to get song from db")
	}
	playbacks := 0
	new.MergeNoOverride(old)
	new.Playbacks++
	playbacks = new.Playbacks
	if err = s.SetSong(ctx, new); err != nil {
		return 0, errors.Wrap(err, "failed to set song into db")
	}
	return playbacks, nil
}

func (s *Service) IncrementUserRequests(ctx context.Context, song *pkg.Song, userID string) {
	userSong, err := s.client.GetUserSong(ctx, song.ID, userID)
	if err != nil {
		if err == ErrNotFound {
			song.Playbacks = 1
		} else {
			return
		}
	} else {
		song.Playbacks = userSong.Playbacks + 1
	}
	err = s.client.SetUserSong(ctx, song, userID)
	if err != nil {
		return
	}
}

func (s *Service) GetRandomSongs(ctx context.Context, n int) ([]*pkg.Song, error) {
	set := make(map[string]pkg.SongID)
	max := len(s.songsShort.List)
	if max == 0 {
		return nil, errors.New("no preloaded songs")
	}

	cooldown := n * 10
	for len(set) < n && cooldown > 0 {
		cooldown--
		rand.Seed(time.Now().UnixNano())
		time.Sleep(time.Nanosecond * 2)
		i := rand.Intn(max)
		s.songsShort.RLock()
		set[s.songsShort.List[i].ID] = s.songsShort.List[i]
		s.songsShort.RUnlock()
	}

	result := make([]*pkg.Song, 0, len(set))
	for _, v := range set {
		song, err := s.GetSong(ctx, v)
		if err != nil {
			return nil, errors.Wrap(err, "get song failed")
		}
		result = append(result, song)
	}
	return result, nil
}

func (s *Service) updateShortCacheProcess(ctx context.Context) {
	// TODO: in config
	ticker := time.NewTicker(3 * time.Hour)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if s.needUpdate() {
					s.setUpdate(false)
					s.updateShortCache(ctx)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (s *Service) updateShortCache(ctx context.Context) {
	list, err := s.client.GetAllSongsID(ctx)
	logger := contexts.GetLogger(ctx)
	if err != nil {
		s.setUpdate(true)
		logger.Error("getting all songs", zap.Error(err))
	}
	s.songsShort.Lock()
	s.songsShort.List = list
	size := len(list)
	s.songsShort.Unlock()
	logger.Info("short cache updated with songs", zap.Int("number", size))
}

func (s *Service) setUpdate(b bool) {
	s.updatesMutex.Lock()
	s.updated = b
	s.updatesMutex.Unlock()
}

func (s *Service) needUpdate() bool {
	s.updatesMutex.Lock()
	defer s.updatesMutex.Unlock()
	return s.updated
}
