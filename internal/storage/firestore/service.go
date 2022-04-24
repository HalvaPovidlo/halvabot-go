package firestore

import (
	"math/rand"
	"sync"
	"time"

	firebase "firebase.google.com/go"
	"github.com/pkg/errors"
	"google.golang.org/api/option"

	"github.com/HalvaPovidlo/discordBotGo/internal/storage"
	"github.com/HalvaPovidlo/discordBotGo/pkg/contexts"
)

type shortCache struct {
	sync.RWMutex
	List []storage.SongID
	// if we want to update this map in realtime
	// Map map[string]struct{}
}

type Firestore struct {
	songs SongsCache
	app   *firebase.App

	songsShort   shortCache
	updatesMutex sync.Mutex
	updated      bool
}

func NewFirestoreService(ctx contexts.Context, credsPath string, songs SongsCache) (*Firestore, error) {
	sa := option.WithCredentialsFile(credsPath)
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		return nil, errors.Wrap(err, "firebase.NewApp failed")
	}
	f := Firestore{
		songs:      songs,
		songsShort: shortCache{},
		app:        app,
	}
	f.updateShortCache(ctx)
	return &f, nil
}

func (f *Firestore) GetSong(ctx contexts.Context, id storage.SongID) (*storage.Song, error) {
	if s, ok := f.songs.Get(f.songs.KeyFromID(id)); ok {
		return s, nil
	}
	client, err := NewClient(ctx, f.app)
	if err != nil {
		return nil, err
	}
	song, err := client.GetSongByID(ctx, id)
	if err != nil {
		return nil, err
	}
	f.songs.Set(f.songs.KeyFromID(id), song)
	return song, nil
}

func (f *Firestore) SetSong(ctx contexts.Context, song *storage.Song) error {
	f.setUpdate(true)
	client, err := NewClient(ctx, f.app)
	if err != nil {
		return err
	}
	if err := client.SetSong(ctx, song); err != nil {
		return errors.Wrap(err, "firestore set song")
	}
	f.songs.Set(f.songs.KeyFromID(song.ID), song)
	return nil
}

func (f *Firestore) updateShortCache(ctx contexts.Context) {
	ticker := time.NewTicker(6 * time.Hour)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if f.needUpdate() {
					f.setUpdate(false)
					client, err := NewClient(ctx, f.app)
					if err != nil {
						f.setUpdate(true)
						continue
					}
					list, err := client.GetAllSongsID(ctx)
					if err != nil {
						f.setUpdate(true)
						ctx.LoggerFromContext().Error(errors.Wrap(err, "getting all songs"))
					}
					f.songsShort.Lock()
					f.songsShort.List = list
					f.songsShort.Unlock()
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (f *Firestore) GetRandomSongs(ctx contexts.Context, n int) ([]storage.Song, error) {
	set := make(map[string]storage.SongID)
	max := len(f.songsShort.List)

	cooldown := n * 10
	for len(set) < n && cooldown > 0 {
		cooldown--
		rand.Seed(time.Now().UnixNano())
		i := rand.Intn(max)
		f.songsShort.RLock()
		set[f.songsShort.List[i].ID] = f.songsShort.List[i]
		f.songsShort.RUnlock()
	}

	result := make([]storage.Song, 0, len(set))
	for _, v := range set {
		song, err := f.GetSong(ctx, v)
		if err != nil {
			return nil, errors.Wrap(err, "get random songs failed")
		}
		result = append(result, *song)
	}
	return result, nil
}

func (f *Firestore) setUpdate(b bool) {
	f.updatesMutex.Lock()
	f.updated = b
	f.updatesMutex.Unlock()
}

func (f *Firestore) needUpdate() bool {
	f.updatesMutex.Lock()
	defer f.updatesMutex.Unlock()
	return f.updated
}
