package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gocarina/gocsv"

	"github.com/HalvaPovidlo/halvabot-go/internal/pkg/item"
	fire "github.com/HalvaPovidlo/halvabot-go/pkg/storage/firestore"
)

// kinopoisk API
// 843650
// 1228254
// COVER, POSTER

const (
	slava   = "242030987536629760"
	dima    = "257456911270674433"
	khodand = "320309512697413633"
	sanek   = "320310971593916416"
	leha    = "320311179245256706"
	artem   = "339482443943772160"
	vlad    = "397466273157480448"
	ivan    = "407858784354959361"
)

//
// var nameToID = map[string]string{
//	"slava":   "242030987536629760",
//	"dima":    "257456911270674433",
//	"khodand": "320309512697413633",
//	"sanek":   "320310971593916416",
//	"leha":    "320311179245256706",
//	"artem":   "339482443943772160",
//	"vlad":    "397466273157480448",
//	"ivan":    "407858784354959361",
// }

const FILE = "films.csv"

type CSVFilm struct {
	Title       string `csv:"title" json:"title,omitempty"`
	URL         string `csv:"url" json:"-"`
	Director    string `csv:"director" json:"director,omitempty"`
	Description string `csv:"description" json:"description,omitempty"`
	Duration    string `csv:"duration" json:"duration,omitempty"`
	Khodand     *int   `csv:"khodand,omitempty"`
	Vlad        *int   `csv:"vlad,omitempty"`
	Artem       *int   `csv:"artem,omitempty"`
	Leha        *int   `csv:"leha,omitempty"`
	Slava       *int   `csv:"slava,omitempty"`
	Sanek       *int   `csv:"sanek,omitempty"`
	Ivan        *int   `csv:"ivan,omitempty"`
	Dima        *int   `csv:"dima,omitempty"`
}

func main() {
	ctx := context.Background()
	app, err := fire.NewFirestoreClient(ctx, "halvabot-firebase.json")
	if err != nil {
		panic(err)
	}
	filmsFile, err := os.OpenFile(FILE, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer filmsFile.Close()

	var films []CSVFilm

	if err := gocsv.UnmarshalFile(filmsFile, &films); err != nil { // Load clients from file
		panic(err)
	}
	for i := 0; i < len(films); i++ {
		film := &films[i]
		fullFilm := buildFilmWithKinopoisk(ctx, filmFromCSV(film))
		if fullFilm == nil {
			fmt.Println("ERROR")
			fmt.Println(film)
			fmt.Println("----------------------------------------------------------------")
			continue
		}
		time.Sleep(time.Second / 10)
		if err := addToFirestore(ctx, app, fullFilm); err != nil {
			panic(err)
		}
	}
	fmt.Println("Success")
}

type fireScore struct {
	Score int `firestore:"score"`
}

func addToFirestore(ctx context.Context, app *firestore.Client, film *item.Film) error {
	filmID := string(film.ID)
	batch := app.Batch()

	filmDoc := app.Collection(fire.FilmsCollection).Doc(filmID)
	batch.Set(filmDoc, film)
	for user, score := range film.Scores {
		userID := string(user)
		batch.Set(filmDoc.Collection(fire.ScoresCollection).Doc(userID), fireScore{Score: score})
		film.UserScore = &score
		batch.Set(app.Collection(fire.UsersCollection).Doc(userID).Collection(fire.FilmsCollection).Doc(filmID), film)
	}
	_, err := batch.Commit(ctx)
	return err
}

func filmFromCSV(csv *CSVFilm) *item.Film {
	if _, err := url.Parse(csv.URL); err != nil {
		panic(err)
	}
	score, average, scores := getScores(csv)
	film := &item.Film{
		ID:          item.FilmID(kinopoiskURLToID(csv.URL)),
		Title:       csv.Title,
		Director:    csv.Director,
		Description: csv.Description,
		Duration:    csv.Duration,
		Score:       &score,
		Average:     average,
		Scores:      scores,
		URL:         csv.URL,
	}

	return film
}

func kinopoiskURLToID(uri string) string {
	const filmURL = `https://www.kinopoisk.ru/film/`
	const seriesURL = `https://www.kinopoisk.ru/series/`
	id := strings.TrimPrefix(uri, filmURL)
	id = strings.TrimPrefix(id, seriesURL)
	return strings.Trim(id, "/")
}

func getScores(csv *CSVFilm) (int, float64, map[item.UserID]int) {
	var score int
	var number int
	scores := make(map[item.UserID]int)
	if csv.Khodand != nil {
		number++
		score += *csv.Khodand
		scores[khodand] = *csv.Khodand
	}
	if csv.Vlad != nil {
		number++
		score += *csv.Vlad
		scores[vlad] = *csv.Vlad
	}
	if csv.Artem != nil {
		number++
		score += *csv.Artem
		scores[artem] = *csv.Artem
	}
	if csv.Leha != nil {
		number++
		score += *csv.Leha
		scores[leha] = *csv.Leha
	}
	if csv.Slava != nil {
		number++
		score += *csv.Slava
		scores[slava] = *csv.Slava
	}
	if csv.Sanek != nil {
		number++
		score += *csv.Sanek
		scores[sanek] = *csv.Sanek
	}
	if csv.Ivan != nil {
		number++
		score += *csv.Ivan
		scores[ivan] = *csv.Ivan
	}
	if csv.Dima != nil {
		number++
		score += *csv.Dima
		scores[dima] = *csv.Dima
	}
	return score, float64(score) / float64(number), scores
}
