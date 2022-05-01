package player

import (
	"github.com/HalvaPovidlo/discordBotGo/internal/audio"
	"io"
	"sync"
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

	radioMutex sync.Mutex
	isRadio    bool
	logger     zap.Logger
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

	song.LastPlay = pkg.PlayDate{Time: time.Now()}
	playbacks, err := s.storage.UpsertSongIncPlaybacks(ctx, song)

	s.logger.Debug("sending command play")
	s.Player.Play(song)
	if err != nil {
		return song, playbacks, ErrStorageQueryFailed.Wrap(errors.Wrap(err, "upsert song with increment").Error())
	}
	return song, playbacks, nil
}

func (s *Service) Random(ctx contexts.Context, n int) ([]*pkg.Song, error) {
	return s.storage.GetRandomSongs(ctx, n)
}

func (s *Service) SetRadio(b bool) {
	s.radioMutex.Lock()
	s.isRadio = b
	s.radioMutex.Unlock()
}

func (s *Service) RadioStatus() bool {
	s.radioMutex.Lock()
	b := s.isRadio
	s.radioMutex.Unlock()
	return b
}

func (s *Service) handleError(err error) {

	if err == ErrQueueEmpty && s.RadioStatus() {
		songs, err := s.storage.GetRandomSongs(contexts.Context{Context: contexts.Background()}, 1)
		if err != nil {
			s.logger.Error(errors.Wrap(err, "radio failed"))
			s.SetRadio(false)
			return
		}
		s.Player.Play(songs[0])
	}
}

func (s *Service) SubscribeOnErrors(h ErrorHandler) {
	s.Player.SubscribeOnErrors(func(err error) {
		if err == io.EOF || err == audio.ErrManualStop || err == ErrQueueEmpty {
			return
		}
		h(err)
	})
}
