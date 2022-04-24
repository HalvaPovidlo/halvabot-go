package firestore

import (
	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/HalvaPovidlo/discordBotGo/internal/storage"
	"github.com/HalvaPovidlo/discordBotGo/pkg/contexts"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	songsCollection = "songs"
	// Maximum batch size by firestore docs
	batchSize              = 500
	approximateSongsNumber = 1000
)

type Client struct {
	*firestore.Client
}

var ErrNotFound = errors.New("no docs found")

func NewClient(ctx contexts.Context, app *firebase.App) (*Client, error) {
	c, err := app.Firestore(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create firestore client")
	}
	return &Client{c}, nil
}

func (c *Client) GetSongByID(ctx contexts.Context, id storage.SongID) (*storage.Song, error) {
	doc, err := c.Collection(songsCollection).Doc(id.String()).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, ErrNotFound
		}
		return nil, errors.Wrapf(err, "failed to get %s from %s", id.String(), songsCollection)
	}
	var s storage.Song
	err = doc.DataTo(&s)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse doc into struct")
	}
	return &s, nil
}

func (c *Client) SetSong(ctx contexts.Context, song *storage.Song) error {
	_, err := c.Collection(songsCollection).Doc(song.ID.String()).Set(ctx, song)
	if err != nil {
		return errors.Wrapf(err, "failed to set %s from %s", song.ID.String(), songsCollection)
	}
	return nil
}

func (c *Client) GetAllSongsID(ctx contexts.Context) ([]storage.SongID, error) {
	iter := c.Collection(songsCollection).Documents(ctx)
	res := make([]storage.SongID, 0, approximateSongsNumber)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, errors.Wrap(err, "iteration failed")
		}
		var s storage.Song
		err = doc.DataTo(&s)
		if err != nil {
			return nil, errors.Wrap(err, "unable to marshal data")
		}
		res = append(res, s.ID)
	}
	return res, nil
}

func (c *Client) WriteBatch(ctx contexts.Context, songs []*storage.Song) error {
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

func (c *Client) doBatch(ctx contexts.Context, songs []*storage.Song) error {
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
