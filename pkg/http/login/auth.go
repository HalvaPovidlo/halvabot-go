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

type userInfo struct {
	Password string `firestore:"password"`
	UserID   string `firestore:"user_id"`
}

type authenticator struct {
	client *firestore.Client
}

func NewFirebaseAuthenticator(client *firestore.Client) *authenticator {
	return &authenticator{
		client: client,
	}
}

func (a *authenticator) AuthUser(ctx context.Context, login, password string) (string, error) {
	doc, err := a.client.Collection(loginsCollection).Doc(login).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return "", nil
		}
		return "", err
	}
	var info userInfo
	err = doc.DataTo(&info)
	if err != nil || info.Password != password {
		return "", err
	}
	return info.UserID, nil
}
