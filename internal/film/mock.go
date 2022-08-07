package film

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/HalvaPovidlo/halvabot-go/internal/pkg/item"
)

type Mock struct {
	films map[string]item.Film
}

func NewMock() *Mock {
	m := &Mock{}
	m.films = make(map[string]item.Film)
	m.films[mockFilm.ID] = mockFilm
	return m
}

func (m *Mock) NewFilm(ctx context.Context, film *item.Film, userID string, withKP bool) (*item.Film, error) {
	if film.UserScore == nil {
		return nil, ErrNoUserScore
	}
	film.Score = *film.UserScore
	film.Average = float64(*film.UserScore)
	film.Scores = make(map[string]int)
	film.Scores[userID] = *film.UserScore

	if withKP {
		film = MergeFilm(&mockKPFilm, film)
	}
	m.films[film.ID] = *film
	return film, nil
}

func (m *Mock) NewKinopoiskFilm(ctx context.Context, uri, userID string, score int) (*item.Film, error) {
	return &mockFilm, nil
}

func (m *Mock) EditFilm(ctx context.Context, film *item.Film) (*item.Film, error) {
	if _, ok := m.films[film.ID]; !ok {
		return nil, errors.New("film not found")
	}
	m.films[film.ID] = *film
	return film, nil
}

func (m *Mock) AllFilms(ctx context.Context) ([]item.Film, error) {
	films := make([]item.Film, 0, len(m.films))
	for _, v := range m.films {
		v.Comments = nil
		films = append(films, v)
	}
	return films, nil
}

func (m *Mock) Film(ctx context.Context, filmID string) (*item.Film, error) {
	f, ok := m.films[filmID]
	if !ok {
		return nil, errors.New("film not found")
	}
	return &f, nil
}

func (m *Mock) Comment(ctx context.Context, text, filmID, userID string) error {
	f, ok := m.films[filmID]
	if !ok {
		return errors.New("film not found")
	}

	f.Comments[uuid.New().String()] = item.Comment{
		UserID:    userID,
		Text:      text,
		CreatedAt: time.Now(),
	}
	return nil
}

func (m *Mock) Score(ctx context.Context, filmID, userID string, score int) (*item.Film, error) {
	f, ok := m.films[filmID]
	if !ok {
		return nil, errors.New("film not found")
	}
	oldScore := f.Scores[userID]
	f.Score += score - oldScore
	f.Scores[userID] = score
	f.Average = float64(f.Score) / float64(len(f.Scores))
	m.films[filmID] = f
	return &f, nil
}

var mockScores = map[string]int{
	"320309512697413633": 1,
	"397466273157480448": 1,
	"339482443943772160": 1,
	"mockman_id":         1,
	"242030987536629760": 0,
}

var mockComments = map[string]item.Comment{
	"1232safsaf": {
		UserID:    "397466273157480448",
		Text:      "Отличный фильм 😀",
		CreatedAt: time.Now().Add(-4 * time.Hour),
	},
	"123skjlaf": {
		UserID:    "397466273157480448",
		Text:      "Передумал, фильм не очень. ![https://cdn-0.emojis.wiki/emoji-pics/twitter/zany-face-twitter.png]",
		CreatedAt: time.Now().Add(-3 * time.Hour),
	},
	"d12d44sdfs": {
		UserID:    "mockman_id",
		Text:      "Вот смотрю я это аниме серию за серией и в голову приходит только, что создатели аниме очень рисковые ребята. Тут очень тонкая и глубокая сатира, не каждый взрослый фильм или сериал затрагивает такое количество важных тем, как это вроде бы детское произведение. Темы освящаются как бы мимоходом, при этом весьма поучительно. Подано больше в комедийном жанре, но в некоторые моменты прям совсем не до смеха, правда их стараются смягчить и свести все к шутке.",
		CreatedAt: time.Now().Add(-2 * time.Hour),
	},
	"bfdoowwe": {
		UserID:    "mockman_id",
		Text:      "Здраствуйте. Я, Кирилл. Хотел бы чтобы вы сделали игру, 3Д-экшон суть такова... Пользователь может играть лесными эльфами, охраной дворца и злодеем. И если пользователь играет эльфами то эльфы в лесу, домики деревяные набигают солдаты дворца и злодеи. Можно грабить корованы... И эльфу раз лесные то сделать так что там густой лес... А движок можно поставить так что вдали деревья картинкой, когда подходиш они преобразовываются в 3-хмерные деревья[1]. Можно покупать и т.п. возможности как в Daggerfall. И враги 3-хмерные тоже, и труп тоже 3д. Можно прыгать и т.п. Если играть за охрану дворца то надо слушаться командира, и защищать дворец от злого (имя я не придумал) и шпионов, партизанов эльфов, и ходит на набеги на когото из этих (эльфов, злого...). Ну а если за злого... то значит шпионы или партизаны эльфов иногда нападают, пользователь сам себе командир может делать что сам захочет прикажет своим войскам с ним самим напасть на дворец и пойдет в атаку. Всего в игре 4 зоны. Т.е. карта и на ней есть 4 зоны, 1 - зона людей (нейтрал), 2- зона императора (где дворец), 3-зона эльфов, 4 - зона злого... (в горах, там есть старый форт...)\n\nТак же чтобы в игре могли не только убить но и отрубить руку и если пользователя не вылечат то он умрет, так же выколоть глаз но пользователь может не умереть а просто пол экрана не видеть, или достать или купить протез, если ногу тоже либо умреш либо будеш ползать либо на коляске котаться, или самое хорошее... поставить протез. Сохранятся можно...\n\nP.S. Я джва года хочу такую игру.",
		CreatedAt: time.Now().Add(-1 * time.Hour),
	},
}

var mockFilm = item.Film{
	ID:                       "841681",
	Title:                    "Токийский гуль",
	TitleOriginal:            "Tokyo Ghoul",
	Poster:                   "https://kinopoiskapiunofficial.tech/images/posters/kp/841681.jpg",
	Cover:                    "",
	Director:                 "",
	Description:              "С обычным студентом Кэном Канэки случается беда, парень попадает в больницу. Но на этом неприятности не заканчиваются: ему пересаживают органы гулей – существ, поедающих плоть людей. После злосчастной операции Канэки становится одним из чудовищ, пытается стать своим, но для людей он теперь изгой, обреченный на уничтожение.",
	Duration:                 "",
	Score:                    4,
	UserScore:                nil,
	Average:                  0.8,
	Scores:                   mockScores,
	Comments:                 mockComments,
	URL:                      "https://www.kinopoisk.ru/film/841681/",
	RatingKinopoisk:          7.2,
	RatingKinopoiskVoteCount: 69196,
	RatingImdb:               7.8,
	RatingImdbVoteCount:      50272,
	Year:                     2014,
	FilmLength:               0,
	Serial:                   true,
	ShortFilm:                false,
	Genres:                   []string{"триллер", "драма", "боевик", "фэнтези", "ужасы", "мультфильм", "аниме"},
}

var mockKPFilm = KinopoiskFilm{
	KinopoiskID:              841681,
	ImdbID:                   "tt3741634",
	NameRu:                   "Токийский гуль",
	NameOriginal:             "Tokyo Ghoul",
	PosterURL:                "https://kinopoiskapiunofficial.tech/images/posters/kp/841681.jpg",
	CoverURL:                 "",
	RatingKinopoisk:          7.2,
	RatingKinopoiskVoteCount: 69196,
	RatingImdb:               7.8,
	RatingImdbVoteCount:      50272,
	Year:                     2014,
	FilmLength:               0,
	Description:              "С обычным студентом Кэном Канэки случается беда, парень попадает в больницу. Но на этом неприятности не заканчиваются: ему пересаживают органы гулей – существ, поедающих плоть людей. После злосчастной операции Канэки становится одним из чудовищ, пытается стать своим, но для людей он теперь изгой, обреченный на уничтожение.",
	Genres: []KinopoiskGenre{
		{Genre: "триллер"},
		{Genre: "драма"},
		{Genre: "боевик"},
		{Genre: "фэнтези"},
		{Genre: "ужасы"},
		{Genre: "мультфильм"},
		{Genre: "аниме"},
	},
	Serial:    true,
	Completed: false,
	WebURL:    "https://www.kinopoisk.ru/film/841681/",
}
