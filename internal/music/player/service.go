package player

import (
	"context"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/HalvaPovidlo/halvabot-go/internal/music/audio"
	"github.com/HalvaPovidlo/halvabot-go/internal/pkg/item"
	"github.com/HalvaPovidlo/halvabot-go/pkg/contexts"
)

var ErrNotConnected = errors.New("player not connected")
var ErrQueueEmpty = errors.New("queue is empty")

type MediaPlayer interface {
	Process(requests <-chan *audio.SongRequest) <-chan error
	Stats() item.SessionStats
	IsPlaying() bool
	Stop()
}

type VoiceClient interface {
	Connection() *discordgo.VoiceConnection
	Connect(guildID, channelID string) error
	IsConnected() bool
	Disconnect() error
}

type ErrorHandler func(err error)

type commandType int

const (
	play commandType = iota
	next
	skip
	stop
	connect
	disconnect
	shuffle
	loop
)

func (c commandType) String() string {
	switch c {
	case play:
		return "play"
	case next:
		return "next"
	case skip:
		return "skip"
	case stop:
		return "stop"
	case connect:
		return "connect"
	case disconnect:
		return "disconnect"
	case shuffle:
		return "shuffle"
	case loop:
		return "loop"
	}
	return ""
}

type command struct {
	Type      commandType
	guildID   string
	channelID string
	entry     *item.Song
	loop      bool
	logger    *zap.Logger
}

// Service all public methods are concurrent and
// most private methods are designed to be synchronous
type Service struct {
	voice VoiceClient
	audio MediaPlayer

	currentLock   sync.Mutex
	current       *item.Song
	isWaited      bool
	queue         Queue
	errs          chan error
	commands      chan *command
	errorHandlers chan ErrorHandler
}

func NewPlayer(ctx context.Context, voice VoiceClient, audio MediaPlayer) *Service {
	p := Service{
		voice: voice,
		audio: audio,
	}
	p.commands, p.errs = p.processCommands(ctx)
	p.errorHandlers = p.processErrors(p.errs)
	return &p
}

// Play next song and enqueue input
func (p *Service) Play(ctx context.Context, s *item.Song) {
	p.commands <- &command{
		Type:   play,
		entry:  s,
		logger: contexts.GetLogger(ctx),
	}
}

func (p *Service) Skip(ctx context.Context) {
	p.commands <- &command{
		Type:   skip,
		logger: contexts.GetLogger(ctx),
	}
}

func (p *Service) Stop(ctx context.Context) {
	p.commands <- &command{
		Type:   stop,
		logger: contexts.GetLogger(ctx),
	}
}

func (p *Service) LoopStatus() bool {
	return p.queue.LoopStatus()
}

func (p *Service) SetLoop(ctx context.Context, b bool) {
	p.commands <- &command{
		Type:   loop,
		loop:   b,
		logger: contexts.GetLogger(ctx),
	}
}

func (p *Service) Connect(ctx context.Context, guildID, channelID string) {
	p.commands <- &command{
		Type:      connect,
		guildID:   guildID,
		channelID: channelID,
		logger:    contexts.GetLogger(ctx),
	}
}

func (p *Service) IsConnected() bool {
	return p.voice.IsConnected()
}

func (p *Service) Disconnect(ctx context.Context) {
	p.commands <- &command{
		Type:   disconnect,
		logger: contexts.GetLogger(ctx),
	}
}

func (p *Service) NowPlaying() *item.Song {
	p.currentLock.Lock()
	defer p.currentLock.Unlock()
	return p.current
}

func (p *Service) setNowPlaying(s *item.Song) {
	p.currentLock.Lock()
	defer p.currentLock.Unlock()
	p.current = s
}

func (p *Service) SongStatus() item.SessionStats {
	s := p.audio.Stats()
	if s.Duration == 0 {
		now := p.NowPlaying()
		if now == nil {
			return item.SessionStats{}
		}
		s.Duration = now.Duration
	}
	return s
}

func (p *Service) SubscribeOnErrors(h ErrorHandler) {
	p.errorHandlers <- h
}
