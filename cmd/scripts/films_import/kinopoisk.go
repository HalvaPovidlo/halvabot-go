//nolint:revive // blame Kinopoisk
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"

	"github.com/HalvaPovidlo/halvabot-go/internal/pkg/item"
)

func buildFilmWithKinopoisk(ctx context.Context, film *item.Film) *item.Film {
	kf, err := getKinopoiskFilm(ctx, film.ID)
	if err != nil {
		fmt.Println("GET KINOPOISK ", err)
		return nil
	}
	var genres []string
	for i := range kf.Genres {
		genres = append(genres, kf.Genres[i].Genre)
	}
	return &item.Film{
		ID:                       film.ID,
		Title:                    kf.NameRu,
		TitleOriginal:            kf.NameOriginal,
		Poster:                   kf.PosterURL,
		Cover:                    kf.CoverURL,
		Director:                 film.Director,
		Description:              kf.Description,
		Duration:                 film.Duration,
		Score:                    film.Score,
		UserScore:                film.UserScore,
		Average:                  film.Average,
		Scores:                   film.Scores,
		Comments:                 film.Comments,
		URL:                      kf.WebUrl,
		RatingKinopoisk:          kf.RatingKinopoisk,
		RatingKinopoiskVoteCount: kf.RatingKinopoiskVoteCount,
		RatingImdb:               kf.RatingImdb,
		RatingImdbVoteCount:      kf.RatingImdbVoteCount,
		Year:                     kf.Year,
		FilmLength:               kf.FilmLength,
		Serial:                   kf.Serial,
		ShortFilm:                kf.ShortFilm,
		Genres:                   genres,
	}
}

func getKinopoiskFilm(ctx context.Context, id string) (*KinopoiskFilm, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://kinopoiskapiunofficial.tech/api/v2.2/films/"+id, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-API-KEY", "558d7fdc-bda9-4da8-a421-5e6bb865bf1e")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("status not ok")
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var film KinopoiskFilm
	if err := json.Unmarshal(data, &film); err != nil {
		return nil, err
	}

	return &film, nil
}

type KinopoiskFilm struct {
	KinopoiskID              int              `json:"kinopoiskId"`
	ImdbID                   string           `json:"imdbId"`
	NameRu                   string           `json:"nameRu"`
	NameOriginal             string           `json:"nameOriginal"`
	PosterURL                string           `json:"posterUrl"`
	CoverURL                 string           `json:"coverUrl"`
	RatingKinopoisk          float64          `json:"ratingKinopoisk"`
	RatingKinopoiskVoteCount int              `json:"ratingKinopoiskVoteCount"`
	RatingImdb               float64          `json:"ratingImdb"`
	RatingImdbVoteCount      int              `json:"ratingImdbVoteCount"`
	Year                     int              `json:"year"`
	FilmLength               int              `json:"filmLength"`
	Description              string           `json:"description"`
	Genres                   []KinopoiskGenre `json:"genres"`
	Serial                   bool             `json:"serial"`
	ShortFilm                bool             `json:"shortFilm"`
	Completed                bool             `json:"completed"`
	WebUrl                   string           `json:"webUrl"`
}

type KinopoiskGenre struct {
	Genre string `json:"genre"`
}
