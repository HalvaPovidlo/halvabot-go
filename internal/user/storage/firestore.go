package storage

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"

	"github.com/HalvaPovidlo/halvabot-go/internal/pkg/item"
	fire "github.com/HalvaPovidlo/halvabot-go/pkg/storage/firestore"
)

type Firestore struct {
	*firestore.Client
}

func NewFirestore(client *firestore.Client) *Firestore {
	return &Firestore{
		Client: client,
	}
}

func (f *Firestore) EditUser(ctx context.Context, user *item.User) error {
	_, err := f.Collection(fire.UsersCollection).Doc(user.ID).Set(ctx, user)
	return err
}

func (f *Firestore) User(ctx context.Context, userID string) (*item.User, error) {
	doc, err := f.Collection(fire.UsersCollection).Doc(userID).Get(ctx)
	if err != nil {
		return nil, err
	}
	var user item.User
	if err := doc.DataTo(&user); err != nil {
		return nil, errors.Wrap(err, "parse user doc")
	}
	user.ID = doc.Ref.ID
	return &user, nil
}

func (f *Firestore) Films(ctx context.Context, userID string) (item.Films, error) {
	var films item.Films
	iter := f.Collection(fire.UsersCollection).Doc(userID).Collection(fire.FilmsCollection).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var film item.Film
		if err := doc.DataTo(&film); err != nil {
			return nil, errors.Wrap(err, "parse film doc")
		}
		films = append(films, film)
	}
	return films, nil
}

func (f *Firestore) Songs(ctx context.Context, userID string) ([]item.Song, error) {
	var songs []item.Song
	iter := f.Collection(fire.UsersCollection).Doc(userID).Collection(fire.SongsCollection).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var song item.Song
		if err := doc.DataTo(&song); err != nil {
			return nil, errors.Wrap(err, "parse song doc")
		}
		songs = append(songs, song)
	}
	return songs, nil
}
