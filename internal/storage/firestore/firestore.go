package firestore

import (
	"context"

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
	// Maximum batch size by firestore docs
	batchSize              = 500
	approximateSongsNumber = 1000
)

type Client struct {
	*firestore.Client
	debug bool
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
	return &Client{c, debug}, nil
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
	_, err := c.Collection(songsCollection).Doc(song.ID.String()).Set(ctx, song)
	if err != nil {
		return errors.Wrapf(err, "failed to set %s from %s", song.ID.String(), songsCollection)
	}
	return nil
}

func (c *Client) GetAllSongsID(ctx contexts.Context) ([]pkg.SongID, error) {
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
		res = append(res, s.ID)
	}
	return res, nil
}

// UpsertSongIncPlaybacks We don't use it because our cash of songs is always consistent
// As we have only one writer to the song db - this bot
func (c *Client) UpsertSongIncPlaybacks(ctx contexts.Context, new *pkg.Song) (int, error) {
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
