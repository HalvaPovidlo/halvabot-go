package user

import (
	"context"
	"time"

	"github.com/AlekSi/pointer"

	"github.com/HalvaPovidlo/halvabot-go/internal/pkg/item"
)

type Mock struct{}

var userMock = &item.User{
	ID:       "mockman_id",
	Username: "Boss",
	Avatar:   "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcROxu17ejFxPV3LuOm6BHqNdrCaY06zZkgprQs80ZYKEw&s",
}

func (s *Mock) EditUser(ctx context.Context, user *item.User) (*item.User, error) {
	userMock = user
	return userMock, nil
}

func (s *Mock) User(ctx context.Context, userID string) (*item.User, error) {
	return userMock, nil
}

func (s *Mock) Films(ctx context.Context, userID string) ([]item.Film, error) {
	return []item.Film{
		{
			ID:                       "12434",
			Title:                    "wasdo;nlk",
			TitleOriginal:            "adkadva",
			Poster:                   "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcROxu17ejFxPV3LuOm6BHqNdrCaY06zZkgprQs80ZYKEw&s",
			Cover:                    "",
			Director:                 "qwffw",
			Description:              "dspavdk;mqm",
			Duration:                 "",
			Score:                    1,
			UserScore:                pointer.ToInt(1),
			Average:                  1,
			Scores:                   map[string]int{"mockman_id": 1},
			URL:                      "https://agwf",
			RatingKinopoisk:          1,
			RatingKinopoiskVoteCount: 1,
			RatingImdb:               1,
			RatingImdbVoteCount:      1,
			Serial:                   false,
			ShortFilm:                true,
			Genres:                   []string{"cook"},
		},
	}, nil
}

func (s *Mock) Songs(ctx context.Context, userID string) ([]item.Song, error) {
	return []item.Song{
		{
			Title:        "daekw",
			URL:          "https://asdd",
			Service:      "youtube",
			ArtistName:   "qaefda",
			ArtistURL:    "https://asdd",
			ArtworkURL:   "https://asdd",
			ThumbnailURL: "https://asdd",
			Playbacks:    1,
			LastPlay:     time.Now(),
			Duration:     111,
		},
	}, nil
}
