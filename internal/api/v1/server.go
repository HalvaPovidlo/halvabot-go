package v1

import (
	"github.com/AlekSi/pointer"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/HalvaPovidlo/halvabot-go/internal/pkg/item"
)

//go:generate oapi-codegen -old-config-style -generate gin -package v1 -o server.gen.go ../../../docs/swagger/swagger.yaml
//go:generate oapi-codegen -old-config-style -generate types -package v1 -o types.gen.go ../../../docs/swagger/swagger.yaml

const basePath = "/api/v1"

type LoginHandler interface {
	PostAuthToken(c *gin.Context)
	HardAuthorization() gin.HandlerFunc
	SoftAuthorization() gin.HandlerFunc
}

type MusicHandler interface {
	PostMusicEnqueueServiceIdentifier(c *gin.Context, service string, kind string)
	PostMusicLoop(c *gin.Context)
	PostMusicRadio(c *gin.Context)
	PostMusicSkip(c *gin.Context)
	GetMusicStatus(c *gin.Context)
}

type FilmsHandler interface {
	GetFilmsAll(c *gin.Context, params GetFilmsAllParams)
	PostFilmsKinopoisk(c *gin.Context)
	PostFilmsNew(c *gin.Context, params PostFilmsNewParams)
	GetFilms(c *gin.Context, id FilmId)
	PatchFilmsId(c *gin.Context, id FilmId)
	PostFilmsIdComment(c *gin.Context, id FilmId)
	PostFilmsIdScore(c *gin.Context, id FilmId)
}

type UserHandler interface {
	GetUserFilms(c *gin.Context, params GetUserFilmsParams)
	GetUserInfo(c *gin.Context)
	PatchUserInfo(c *gin.Context)
	GetUserSongs(c *gin.Context)
}

type Server struct {
	router *gin.Engine
	MusicHandler
	LoginHandler
	FilmsHandler
	UserHandler
}

func NewServer(login LoginHandler, music MusicHandler, films FilmsHandler, users UserHandler) *Server {
	return &Server{
		LoginHandler: login,
		MusicHandler: music,
		FilmsHandler: films,
		UserHandler:  users,
		router:       gin.New(),
	}
}

func (s *Server) Run(ip, port, swagger string, debug bool) {
	if !debug {
		gin.SetMode(gin.ReleaseMode)
		gin.DisableConsoleColor()
	}
	s.router.Use(CORS())
	s.router.StaticFile(swagger, "."+swagger)
	s.router.GET(
		"/swagger/*any",
		ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("http://"+ip+":"+port+swagger)),
	)
	s.RegisterHandlers()
	go func() {
		err := s.router.Run(":" + port)
		if err != nil {
			panic(errors.Wrap(err, "run router"))
		}
	}()
}

func (s *Server) RegisterHandlers() {
	wrapper := ServerInterfaceWrapper{
		Handler: s,
		HandlerMiddlewares: []MiddlewareFunc{
			MiddlewareFunc(gin.LoggerWithConfig(gin.LoggerConfig{SkipPaths: []string{"/api/v1/music/status"}})),
			MiddlewareFunc(gin.Recovery()),
		},
	}
	api := s.router.Group(basePath)
	api.Use(s.SoftAuthorization())
	api.POST("/auth/token", wrapper.PostAuthToken)
	api.GET("/music/status", wrapper.GetMusicStatus)
	api.GET("/films/all", wrapper.GetFilmsAll)
	api.GET("/films/:id", wrapper.GetFilms)

	api.Use(s.HardAuthorization())
	api.POST("/music/enqueue/:service/:kind", wrapper.PostMusicEnqueueServiceIdentifier)
	api.POST("/music/loop", wrapper.PostMusicLoop)
	api.POST("/music/radio", wrapper.PostMusicRadio)
	api.POST("/music/skip", wrapper.PostMusicSkip)

	api.POST("/films/kinopoisk", wrapper.PostFilmsKinopoisk)
	api.POST("/films/new", wrapper.PostFilmsNew)
	api.POST("/films/:id", wrapper.PatchFilmsId)
	api.POST("/films/:id/comment", wrapper.PostFilmsIdComment)
	api.POST("/films/:id/score", wrapper.PostFilmsIdScore)

	api.GET("/user/films", wrapper.GetUserFilms)
	api.GET("/user/info", wrapper.GetUserInfo)
	api.PATCH("/user/info", wrapper.PatchUserInfo)
	api.GET("/user/songs", wrapper.GetUserSongs)
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func ConvertItem(f *Film) *item.Film {
	scores := make(map[string]int)
	if f.Scores != nil && len(*f.Scores) != 0 {
		for k, v := range *f.Scores {
			if value, ok := v.(int); ok {
				scores[k] = value
			}
		}
	}
	var genres []string
	if f.Genres != nil {
		genres = *f.Genres
	}
	return &item.Film{
		ID:                       f.FilmId,
		Title:                    f.Title,
		TitleOriginal:            pointer.GetString(f.TitleOriginal),
		Poster:                   pointer.GetString(f.Poster),
		Cover:                    pointer.GetString(f.Cover),
		Director:                 pointer.GetString(f.Director),
		Description:              pointer.GetString(f.Description),
		Duration:                 pointer.GetString(f.Duration),
		Score:                    f.Score,
		UserScore:                f.UserScore,
		Average:                  float64(f.Average),
		Scores:                   scores,
		WithComments:             false,
		URL:                      f.Kinopoisk,
		RatingKinopoisk:          float64(f.RatingKinopoisk),
		RatingKinopoiskVoteCount: f.RatingKinopoiskVoteCount,
		RatingImdb:               float64(f.RatingImdb),
		RatingImdbVoteCount:      f.RatingImdbVoteCount,
		Year:                     pointer.GetInt(f.Year),
		FilmLength:               pointer.GetInt(f.FilmLength),
		Serial:                   f.Serial,
		ShortFilm:                f.ShortFilm,
		Genres:                   genres,
	}
}

func ConvertFilm(f *item.Film) Film {
	comments := make(map[string]interface{})
	for k, v := range f.Comments {
		comments[k] = v
	}
	scores := make(map[string]interface{})
	for k, v := range f.Scores {
		scores[k] = v
	}
	return Film{
		Average:                  float32(f.Average),
		Comments:                 &comments,
		Cover:                    pointer.ToStringOrNil(f.Cover),
		Description:              pointer.ToStringOrNil(f.Description),
		Director:                 pointer.ToStringOrNil(f.Director),
		Duration:                 pointer.ToStringOrNil(f.Duration),
		FilmId:                   f.ID,
		FilmLength:               pointer.ToIntOrNil(f.FilmLength),
		Genres:                   &f.Genres,
		Kinopoisk:                f.URL,
		Poster:                   pointer.ToStringOrNil(f.Poster),
		RatingImdb:               float32(f.RatingImdb),
		RatingImdbVoteCount:      f.RatingImdbVoteCount,
		RatingKinopoisk:          float32(f.RatingKinopoisk),
		RatingKinopoiskVoteCount: f.RatingKinopoiskVoteCount,
		Score:                    f.Score,
		Scores:                   &scores,
		Serial:                   f.Serial,
		ShortFilm:                f.ShortFilm,
		Title:                    f.Title,
		TitleOriginal:            pointer.ToStringOrNil(f.TitleOriginal),
		UserScore:                f.UserScore,
		Year:                     pointer.ToIntOrNil(f.Year),
	}
}

func BuildSong(song *item.Song) *Song {
	return &Song{
		ArtistName:   song.ArtistName,
		ArtistUrl:    song.ArtistURL,
		ArtworkUrl:   song.ArtworkURL,
		LastPlay:     song.LastPlay,
		Playbacks:    song.Playbacks,
		Service:      convertService(song.Service),
		ThumbnailUrl: song.ThumbnailURL,
		Title:        song.Title,
		Url:          song.URL,
	}
}

func convertService(service item.ServiceName) SongService {
	if service == item.ServiceYouTube {
		return Youtube
	}
	return Unknown
}

func ConvertSortKet(k Sort) item.SortKey {
	switch k {
	case SortAverage:
		return item.AverageSort
	case SortHalva:
		return item.HalvaSort
	case SortImdb:
		return item.ImdbSort
	case SortKinopoisk:
		return item.KinopoiskSort
	case SortRandom:
		return item.RandomSort
	case SortScore:
		return item.ScoreSort
	case SortTitle:
		return item.TitleSort
	}
	return item.TitleSort
}
