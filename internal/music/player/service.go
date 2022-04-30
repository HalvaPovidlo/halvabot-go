package player

import (
	"time"

	"github.com/pkg/errors"

	"github.com/HalvaPovidlo/discordBotGo/internal/pkg"
	"github.com/HalvaPovidlo/discordBotGo/pkg/contexts"
	"github.com/HalvaPovidlo/discordBotGo/pkg/zap"
)

type Firestore interface {
	UpsertSongIncPlaybacks(ctx contexts.Context, new *pkg.Song) (int, error)
	GetSong(ctx contexts.Context, id pkg.SongID) (*pkg.Song, error)
	SetSong(ctx contexts.Context, song *pkg.Song) error
	GetRandomSongs(ctx contexts.Context, n int) ([]*pkg.Song, error)
}

type YouTube interface {
	FindSong(ctx contexts.Context, query string) (*pkg.Song, error)
}

type Service struct {
	*Player
	storage Firestore
	youtube YouTube

	logger zap.Logger
}

func NewMusicService(ctx contexts.Context, storage Firestore, youtube YouTube, voice VoiceClient, audio MediaPlayer, logger zap.Logger) *Service {
	return &Service{
		Player:  NewPlayer(ctx, voice, audio, logger),
		storage: storage,
		youtube: youtube,
		logger:  logger,
	}
}

func (s *Service) Play(ctx contexts.Context, query, guildID, channelID string) (*pkg.Song, int, error) {
	s.logger.Debug("Finding song")
	song, err := s.youtube.FindSong(ctx, query)
	if err != nil {
		return nil, 0, ErrSongNotFound.Wrap(err.Error())
	}

	s.Connect(guildID, channelID)

	song.LastPlay = pkg.PlayDate{Time: time.Now()}
	playbacks, err := s.storage.UpsertSongIncPlaybacks(ctx, song)
	if err != nil {
		err = ErrStorageQueryFailed.Wrap(errors.Wrap(err, "upsert song with increment").Error())
	}

	s.logger.Debug("sending command play")
	s.Player.Play(song)
	return song, playbacks, err
}

func (s *Service) Enqueue(ctx contexts.Context, query string) (*pkg.Song, int, error) {
	logger := ctx.LoggerFromContext()
	logger.Debug("Finding song")
	song, err := s.youtube.FindSong(ctx, query)
	if err != nil {
		return nil, 0, ErrSongNotFound.Wrap(err.Error())
	}

	var err2 *Error
	song.LastPlay = pkg.PlayDate{Time: time.Now()}
	playbacks, err := s.storage.UpsertSongIncPlaybacks(ctx, song)
	if err != nil {
		err2 = ErrStorageQueryFailed.Wrap(errors.Wrap(err, "upsert song with increment").Error())
	}

	s.logger.Debug("sending command play")
	s.Player.Play(song)
	return song, playbacks, err2
}
