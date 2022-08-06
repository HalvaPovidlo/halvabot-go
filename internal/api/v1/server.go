package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	GetFilmsAll(c *gin.Context)
	PostFilmsKinopoisk(c *gin.Context)
	PostFilmsNew(c *gin.Context, params PostFilmsNewParams)
	GetFilms(c *gin.Context, id FilmId)
	PostFilmsId(c *gin.Context, id FilmId)
	PostFilmsIdComment(c *gin.Context, id FilmId)
	PostFilmsIdScore(c *gin.Context, id FilmId)
}

type Server struct {
	router *gin.Engine
	MusicHandler
	LoginHandler
	FilmsHandler
}

func NewServer(music MusicHandler, login LoginHandler) *Server {
	return &Server{
		MusicHandler: music,
		LoginHandler: login,
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
	api.POST("/films/:id", wrapper.PostFilmsId)
	api.POST("/films/:id/comment", wrapper.PostFilmsIdComment)
	api.POST("/films/:id/score", wrapper.PostFilmsIdScore)
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
