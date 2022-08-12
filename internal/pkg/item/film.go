package item

import (
	"math"
	"math/rand"
	"sort"
	"time"
)

type SortKey int8

const (
	// Ascending
	TitleSort SortKey = iota
	RandomSort
	ScoreSort
	AverageSort
	KinopoiskSort
	ImdbSort
	HalvaSort
)

type Film struct {
	ID                       string             `firestore:"-" json:"id"`
	Title                    string             `firestore:"title,omitempty" json:"title"`
	TitleOriginal            string             `firestore:"title_original,omitempty" json:"title_original,omitempty"`
	Poster                   string             `firestore:"cover,omitempty" json:"cover,omitempty"`
	Cover                    string             `firestore:"poster,omitempty" json:"poster,omitempty"`
	Director                 string             `firestore:"director,omitempty" json:"director,omitempty"`
	Description              string             `firestore:"description,omitempty" json:"description,omitempty"`
	Duration                 string             `firestore:"duration,omitempty" json:"duration,omitempty"`
	Score                    int                `firestore:"score" json:"score"`
	UserScore                *int               `firestore:"user_score" json:"user_score,omitempty"`
	Average                  float64            `firestore:"average,omitempty" json:"average,omitempty"`
	Scores                   map[string]int     `firestore:"scores" json:"scores,omitempty"`
	Comments                 map[string]Comment `firestore:"-" json:"comments,omitempty"`
	WithComments             bool               `firestore:"-" json:"-"`
	URL                      string             `firestore:"kinopoisk,omitempty" json:"kinopoisk,omitempty"`
	RatingKinopoisk          float64            `firestore:"rating_kinopoisk,omitempty" json:"rating_kinopoisk,omitempty"`
	RatingKinopoiskVoteCount int                `firestore:"rating_kinopoisk_vote_count,omitempty" json:"rating_kinopoisk_vote_count,omitempty"`
	RatingImdb               float64            `firestore:"rating_imdb,omitempty" json:"rating_imdb,omitempty"`
	RatingImdbVoteCount      int                `firestore:"rating_imdb_vote_count,omitempty" json:"rating_imdb_vote_count,omitempty"`
	RatingHalva              float64            `firestore:"rating_halva" json:"rating_halva"`
	Year                     int                `firestore:"year,omitempty" json:"year,omitempty"`
	FilmLength               int                `firestore:"film_length,omitempty" json:"film_length,omitempty"`
	Serial                   bool               `firestore:"serial" json:"serial"`
	ShortFilm                bool               `firestore:"short_film" json:"short_film"`
	Genres                   []string           `firestore:"genres,omitempty" json:"genres,omitempty"`
}

type Comment struct {
	UserID    string    `firestore:"user_id" json:"user_id"`
	Text      string    `firestore:"text" json:"text"`
	CreatedAt time.Time `firestore:"created_at" json:"created_at"`
}

func (f *Film) Rate(score int, userID string) {
	oldScore := f.Scores[userID]
	f.Score += score - oldScore
	f.Scores[userID] = score
	f.Average = float64(f.Score) / float64(len(f.Scores))
	f.RatingHalva = float64(f.Score) * math.Abs(f.Average)
}

type Films []Film

func (f Films) Sort(code SortKey) Films {
	// TitleSort
	sort.Slice(f, func(i, j int) bool {
		return f[i].Title < f[j].Title
	})
	switch code {
	case RandomSort:
		rand.Shuffle(len(f), func(i, j int) {
			f[i], f[j] = f[j], f[i]
		})
	case ScoreSort:
		sort.SliceStable(f, func(i, j int) bool {
			return f[i].Score < f[j].Score
		})
	case AverageSort:
		sort.SliceStable(f, func(i, j int) bool {
			return f[i].Average < f[j].Average
		})
	case KinopoiskSort:
		sort.SliceStable(f, func(i, j int) bool {
			return f[i].RatingKinopoisk < f[j].RatingKinopoisk
		})
	case ImdbSort:
		sort.SliceStable(f, func(i, j int) bool {
			return f[i].RatingImdb < f[j].RatingImdb
		})
	case HalvaSort:
		sort.SliceStable(f, func(i, j int) bool {
			return f[i].RatingHalva < f[j].RatingHalva
		})
	}
	return f
}
