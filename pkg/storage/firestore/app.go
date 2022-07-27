package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
)

func NewFirestoreClient(ctx context.Context, creds string) (*firestore.Client, error) {
	opts := option.WithCredentialsFile(creds)
	app, err := firebase.NewApp(ctx, nil, opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create firebase app")
	}
	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create firestore client")
	}
	return client, nil
}
