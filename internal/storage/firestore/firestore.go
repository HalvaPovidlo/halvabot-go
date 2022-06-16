package firestore

import (
	"context"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"

	"github.com/HalvaPovidlo/discordBotGo/internal/pkg"
	"github.com/HalvaPovidlo/discordBotGo/pkg/contexts"
	"google.golang.org/grpc/status"
)

const (
	songsCollection = "songs"
	usersCollection = "users"
	// Maximum batch size by firestore docs
	batchSize              = 500
	approximateSongsNumber = 1000
)

type Client struct {
	*firestore.Client
	updateMx  sync.Mutex
	songs     map[string]*pkg.Song
	userSongs map[string]map[string]*pkg.Song
	debug     bool
}

var ErrNotFound = errors.New("no docs found")

func NewFirestoreClient(ctx contexts.Context, creds string, debug bool) (*Client, error) {
	sa := option.WithCredentialsFile(creds)
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create firebase app")
	}
	c, err := app.Firestore(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create firestore client")
	}
	client := &Client{
		Client:    c,
		songs:     make(map[string]*pkg.Song),
		userSongs: make(map[string]map[string]*pkg.Song),
		debug:     debug,
	}
	client.updateSongs(ctx)
	client.updateUserSongs(ctx)
	return client, nil
}

func (c *Client) GetSongByID(ctx contexts.Context, id pkg.SongID) (*pkg.Song, error) {
	doc, err := c.Collection(songsCollection).Doc(id.String()).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, ErrNotFound
		}
		return nil, errors.Wrapf(err, "failed to get %s from %s", id.String(), songsCollection)
	}
	var s pkg.Song
	err = doc.DataTo(&s)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse doc into struct")
	}
	return &s, nil
}

func (c *Client) SetSong(ctx contexts.Context, song *pkg.Song) error {
	if c.debug {
		return nil
	}
	c.updateMx.Lock()
	c.songs[song.ID.String()] = song
	c.updateMx.Unlock()
	return nil
}

func (c *Client) SetSongForced(ctx contexts.Context, song *pkg.Song) error {
	if c.debug {
		return nil
	}
	ctx.LoggerFromContext().Infof("DB: SetSongForced %s", song.ID)
	_, err := c.Collection(songsCollection).Doc(song.ID.String()).Set(ctx, song)
	if err != nil {
		return errors.Wrapf(err, "failed to set %s from %s", song.ID.String(), songsCollection)
	}
	return nil
}

func (c *Client) GetUserSong(ctx contexts.Context, id pkg.SongID, user string) (*pkg.Song, error) {
	ctx.LoggerFromContext().Infof("DB: GetUserSong id:%s user:%s", id, user)
	doc, err := c.Collection(usersCollection).Doc(user).Collection(songsCollection).Doc(id.String()).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, ErrNotFound
		}
		return nil, errors.Wrapf(err, "failed to get %s from %s", id.String(), usersCollection)
	}
	var s pkg.Song
	err = doc.DataTo(&s)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse doc into struct")
	}
	return &s, nil
}

func (c *Client) SetUserSong(ctx contexts.Context, song *pkg.Song, user string) error {
	if c.debug {
		return nil
	}
	c.updateMx.Lock()
	if _, ok := c.userSongs[user]; !ok {
		c.userSongs[user] = make(map[string]*pkg.Song)
	}
	c.userSongs[user][song.ID.String()] = song
	c.updateMx.Unlock()
	return nil
}

func (c *Client) GetAllSongsID(ctx contexts.Context) ([]pkg.SongID, error) {
	if c.debug {
		return nil, nil
	}
	ctx.LoggerFromContext().Info("DB: GetAllSongsID")
	iter := c.Collection(songsCollection).Documents(ctx)
	res := make([]pkg.SongID, 0, approximateSongsNumber)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, errors.Wrap(err, "iteration failed")
		}
		var s pkg.Song
		err = doc.DataTo(&s)
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal data")
		}
		if s.ID.ID == "" {
			s.ID = pkg.GetIDFromURL(s.URL)
		}
		res = append(res, s.ID)
	}
	return res, nil
}

// UpsertSongIncPlaybacks We don't use it because our cash of songs is always consistent
// As we have only one writer to the song db - this bot
func (c *Client) UpsertSongIncPlaybacks(ctx contexts.Context, new *pkg.Song) (int, error) {
	ctx.LoggerFromContext().Infof("DB: UpsertSongIncPlaybacks %s", new.ID)
	ref := c.Collection(songsCollection).Doc(new.ID.String())
	playbacks := 0
	err := c.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(ref) // tx.Get, NOT ref.Get!
		if err != nil {
			return err
		}
		var old pkg.Song
		if err := doc.DataTo(&old); err != nil {
			return errors.Wrap(err, "parsing data to Song failed")
		}
		playbacks = old.Playbacks + 1
		new.MergeNoOverride(&old)
		new.Playbacks = playbacks
		return tx.Set(ref, new)
	})
	if err != nil {
		return 0, errors.Wrap(err, "transaction failed")
	}
	return playbacks, nil
}

func (c *Client) updateSongs(ctx contexts.Context) {
	ticker := time.NewTicker(time.Second * 30)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.updateMx.Lock()
				if len(c.songs) == 0 {
					c.updateMx.Unlock()
					continue
				}
				toSend := make([]*pkg.Song, 0, len(c.songs))
				for k, v := range c.songs {
					toSend = append(toSend, v)
					delete(c.songs, k)
				}
				c.updateMx.Unlock()
				ctx.LoggerFromContext().Infof("DB: updating songs %d", len(toSend))
				err := c.WriteBatch(ctx, toSend)
				if err != nil {
					ctx.LoggerFromContext().Error(err, "DB: unable to update songs")
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (c *Client) updateUserSongs(ctx contexts.Context) {
	ticker := time.NewTicker(time.Minute)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.updateMx.Lock()
				toSend := make(map[string][]*pkg.Song)
				for user, songs := range c.userSongs {
					toSend[user] = make([]*pkg.Song, 0, len(songs))
					for k, v := range songs {
						toSend[user] = append(toSend[user], v)
						delete(c.songs, k)
					}
					delete(c.userSongs, user)
				}
				c.updateMx.Unlock()
				go func() {
					for user, songs := range toSend {
						ctx.LoggerFromContext().Infof("DB: updateUserSongs user:%s songs:%d", user, len(songs))
						for i := range songs {
							_, err := c.Collection(usersCollection).Doc(user).Collection(songsCollection).Doc(songs[i].ID.String()).Set(ctx, songs[i])
							if err != nil {
								ctx.LoggerFromContext().Error("DB: while updating User:", user, len(songs), "Song:", songs[i], "Error", err)
							}
						}
					}
				}()
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (c *Client) WriteBatch(ctx contexts.Context, songs []*pkg.Song) error {
	size := len(songs)
	for i := 0; i < size; i += batchSize {
		k := i + batchSize
		if k > size {
			k = size
		}
		err := c.doBatch(ctx, songs[i:k])
		if err != nil {
			return errors.Wrapf(err, "faild to send songs batch from %d to %d", i, k)
		}
	}
	return nil
}

func (c *Client) doBatch(ctx contexts.Context, songs []*pkg.Song) error {
	batch := c.Batch()
	for s := range songs {
		batch.Set(c.Collection(songsCollection).Doc(songs[s].ID.String()), songs[s])
	}
	_, err := batch.Commit(ctx)
	return err
}

// Example of NOT FULL REWRITING (WITH DELETING) set (HACK with json)
//
// var inInterface map[string]interface{}
// inrec, _ := json.Marshal(song)
// json.Unmarshal(inrec, &inInterface)
// if lp, ok := inInterface["last_play"]; ok {
// if lp == "0001-01-01T00:00:00Z" {
// delete(inInterface, "last_play")
// }
// }
// _, err = client.Collection("songs").Doc("asda").Set(ctx, inInterface, firestore.MergeAll)
// if err != nil {
//	panic(err)
//	return
// }
