package login

import (
	"context"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	loginsCollection = "logins"
)

type AccountInfo struct {
	Password string `firestore:"password"`
	UserID   string `firestore:"user_id"`
}

type storage struct {
	client *firestore.Client
}

func NewAccountStorage(client *firestore.Client) *storage {
	return &storage{
		client: client,
	}
}

func (s *storage) GetAccount(ctx context.Context, login string) (*AccountInfo, error) {
	doc, err := s.client.Collection(loginsCollection).Doc(login).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}
	var info AccountInfo
	err = doc.DataTo(&info)
	return &info, err
}
