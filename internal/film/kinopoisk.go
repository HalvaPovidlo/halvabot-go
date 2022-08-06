package film

import (
	"context"
	"io"
	"net/http"
	"strings"

	"encoding/json"
	"github.com/pkg/errors"

	"github.com/HalvaPovidlo/halvabot-go/internal/pkg/item"
)

const (
	filmURL   = `https://www.kinopoisk.ru/film/`
	seriesURL = `https://www.kinopoisk.ru/series/`

	apiFilms      = "https://kinopoiskapiunofficial.tech/api/v2.2/films/"
	xApiKeyHeader = "X-API-KEY"
)

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

type Kinopoisk struct {
	apiKey string
	client *http.Client
}

func NewKinopoisk(apiKey string) *Kinopoisk {
	return &Kinopoisk{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

func (k *Kinopoisk) GetFilm(ctx context.Context, id string) (*KinopoiskFilm, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiFilms+id, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add(xApiKeyHeader, k.apiKey)
	resp, err := k.client.Do(req)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("status not ok: " + string(data))
	}

	var film KinopoiskFilm
	if err := json.Unmarshal(data, &film); err != nil {
		return nil, errors.Wrapf(err, "unmarshall film")
	}
	return &film, nil
}

func BuildFilm(kf *KinopoiskFilm) *item.Film {
	var genres []string
	for i := range kf.Genres {
		genres = append(genres, kf.Genres[i].Genre)
	}
	return &item.Film{
		ID:                       IDFromKinopoiskURL(kf.WebUrl),
		Title:                    kf.NameRu,
		TitleOriginal:            kf.NameOriginal,
		Poster:                   kf.PosterURL,
		Cover:                    kf.CoverURL,
		Description:              kf.Description,
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

func MergeFilm(kf *KinopoiskFilm, f *item.Film) *item.Film {
	if f.Title == "" {
		f.Title = kf.NameRu
	}
	if f.TitleOriginal == "" {
		f.TitleOriginal = kf.NameOriginal
	}
	if f.Poster == "" {
		f.Poster = kf.PosterURL
	}
	if f.Cover == "" {
		f.Cover = kf.CoverURL
	}
	if f.Description == "" {
		f.Director = kf.Description
	}
	if f.URL == "" {
		f.URL = kf.WebUrl
	}
	if f.Year == 0 {
		f.Year = kf.Year
	}
	if f.FilmLength == 0 {
		f.FilmLength = kf.FilmLength
	}
	if len(f.Genres) == 0 {
		for i := range kf.Genres {
			f.Genres = append(f.Genres, kf.Genres[i].Genre)
		}
	}
	return &item.Film{
		ID:                       f.ID,
		Title:                    f.Title,
		TitleOriginal:            f.TitleOriginal,
		Poster:                   f.Poster,
		Cover:                    f.Cover,
		Director:                 f.Director,
		Description:              f.Description,
		Duration:                 f.Duration,
		Score:                    f.Score,
		UserScore:                f.UserScore,
		Average:                  f.Average,
		Scores:                   f.Scores,
		Comments:                 f.Comments,
		WithComments:             f.WithComments,
		URL:                      f.URL,
		RatingKinopoisk:          kf.RatingKinopoisk,
		RatingKinopoiskVoteCount: kf.RatingKinopoiskVoteCount,
		RatingImdb:               kf.RatingImdb,
		RatingImdbVoteCount:      kf.RatingImdbVoteCount,
		Year:                     f.Year,
		FilmLength:               f.FilmLength,
		Serial:                   kf.Serial,
		ShortFilm:                kf.ShortFilm,
		Genres:                   f.Genres,
	}
}

func IDFromKinopoiskURL(uri string) string {
	id := strings.TrimPrefix(uri, filmURL)
	id = strings.TrimPrefix(id, seriesURL)
	return strings.Trim(id, "/")
}
