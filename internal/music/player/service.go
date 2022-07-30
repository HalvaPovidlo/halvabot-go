package player

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/HalvaPovidlo/halvabot-go/internal/music/audio"
	"github.com/HalvaPovidlo/halvabot-go/internal/pkg/item"
	"github.com/HalvaPovidlo/halvabot-go/pkg/contexts"
)

const maxRadioSongDuration = 10000

type Firestore interface {
	UpsertSongIncPlaybacks(ctx context.Context, new *item.Song) (int, error)
	IncrementUserRequests(ctx context.Context, song *item.Song, userID string)
	GetRandomSongs(ctx context.Context, n int) ([]*item.Song, error)
}

type YouTube interface {
	FindSong(ctx context.Context, query string) (*item.Song, error)
	EnsureStreamInfo(ctx context.Context, song *item.Song) (*item.Song, error)
}

type Service struct {
	*Player
	storage Firestore
	youtube YouTube

	radioMutex sync.Mutex
	isRadio    bool
}

func NewMusicService(ctx context.Context, storage Firestore, youtube YouTube, voice VoiceClient, audio MediaPlayer) *Service {
	s := &Service{
		Player:  NewPlayer(ctx, voice, audio),
		storage: storage,
		youtube: youtube,
	}
	s.Player.SubscribeOnErrors(s.handleError)
	return s
}

func (s *Service) Play(ctx context.Context, query, userID, guildID, channelID string) (*item.Song, error) {
	if !s.Player.voice.IsConnected() && (channelID == "" || guildID == "") {
		return nil, ErrNotConnected
	}

	contexts.GetLogger(ctx).Info("finding song")
	song, err := s.youtube.FindSong(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "find and load song from youtube")
	}

	if channelID != "" || guildID != "" {
		s.Connect(ctx, guildID, channelID)
	}

	song.LastPlay = time.Now()
	_, err = s.storage.UpsertSongIncPlaybacks(ctx, song)
	if err != nil {
		err = errors.Wrap(err, "upsert song with increment")
	}

	if userID != "" {
		s.storage.IncrementUserRequests(ctx, song, userID)
	}

	go s.Player.Play(ctx, song)
	return song, err
}

func (s *Service) Random(ctx context.Context, n int) ([]*item.Song, error) {
	return s.storage.GetRandomSongs(ctx, n)
}

func (s *Service) SetRadio(ctx context.Context, b bool, guildID, channelID string) error {
	if !b {
		s.setRadio(b)
		return nil
	}
	if !s.Player.voice.IsConnected() {
		if guildID == "" || channelID == "" {
			return ErrNotConnected
		}
		s.Player.Connect(ctx, guildID, channelID)
	}
	s.setRadio(b)
	if s.NowPlaying() == nil {
		return s.playRandomSong(ctx)
	}
	return nil
}

func (s *Service) setRadio(b bool) {
	s.radioMutex.Lock()
	s.isRadio = b
	s.radioMutex.Unlock()
}

func (s *Service) playRandomSong(ctx context.Context) error {
	songs, err := s.storage.GetRandomSongs(ctx, 1)
	if err != nil {
		return errors.Wrap(err, "get 1 random song from bd")
	}
	song := songs[0]
	if song.StreamURL == "" {
		song, err = s.youtube.EnsureStreamInfo(ctx, song)
		if err != nil {
			contexts.GetLogger(ctx).Error("ensure stream info for radio", zap.Error(err))
			return s.playRandomSong(ctx)
		}
		if song.Duration > maxRadioSongDuration {
			contexts.GetLogger(ctx).Info("too long song found - skipping")
			return s.playRandomSong(ctx)
		}
	}
	s.Player.Play(ctx, song)
	return nil
}

func (s *Service) RadioStatus() bool {
	s.radioMutex.Lock()
	b := s.isRadio
	s.radioMutex.Unlock()
	return b
}

func (s *Service) handleError(err error) {
	if errors.Is(err, ErrQueueEmpty) {
		if s.RadioStatus() {
			err := s.playRandomSong(context.Background())
			if err != nil {
				s.setRadio(false)
			}
		}
		return
	}
	if !errors.Is(err, audio.ErrManualStop) && !errors.Is(err, io.EOF) {
		s.setRadio(false)
	}
}

func (s *Service) SubscribeOnErrors(h ErrorHandler) {
	s.Player.SubscribeOnErrors(func(err error) {
		if errors.Is(err, io.EOF) || errors.Is(err, audio.ErrManualStop) || errors.Is(err, ErrQueueEmpty) {
			return
		}
		h(err)
	})
}

func (s *Service) Stop(ctx context.Context) {
	s.setRadio(false)
	s.Player.Stop(ctx)
}

func (s *Service) Disconnect(ctx context.Context) {
	s.setRadio(false)
	s.Player.Disconnect(ctx)
}

func (s *Service) Status() item.PlayerStatus {
	return item.PlayerStatus{
		Loop:  s.LoopStatus(),
		Radio: s.RadioStatus(),
		Song:  s.SongStatus(),
		Now:   s.NowPlaying(),
	}
}
