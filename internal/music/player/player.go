package player

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/HalvaPovidlo/discordBotGo/internal/music/audio"
	"github.com/HalvaPovidlo/discordBotGo/internal/pkg"
	"github.com/HalvaPovidlo/discordBotGo/pkg/contexts"
)

var ErrNotConnected = errors.New("player not connected")
var ErrQueueEmpty = errors.New("queue is empty")

type MediaPlayer interface {
	Process(requests <-chan *audio.SongRequest) <-chan error
	Stats() pkg.SessionStats
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
	entry     *pkg.Song
	loop      bool
	logger    *zap.Logger
}

// Player all public methods are concurrent and
// most private methods are designed to be synchronous
type Player struct {
	voice VoiceClient
	audio MediaPlayer

	currentLock   sync.Mutex
	current       *pkg.Song
	isWaited      bool
	queue         Queue
	errs          chan error
	commands      chan *command
	errorHandlers chan ErrorHandler
}

func NewPlayer(ctx context.Context, voice VoiceClient, audio MediaPlayer) *Player {
	p := Player{
		voice: voice,
		audio: audio,
	}
	p.commands, p.errs = p.processCommands(ctx)
	p.errorHandlers = p.processErrors(p.errs)
	return &p
}

// Play next song and enqueue input
func (p *Player) Play(ctx context.Context, s *pkg.Song) {
	p.commands <- &command{
		Type:   play,
		entry:  s,
		logger: contexts.GetLogger(ctx),
	}
}

func (p *Player) Skip(ctx context.Context) {
	p.commands <- &command{
		Type:   skip,
		logger: contexts.GetLogger(ctx),
	}
}

func (p *Player) Stop(ctx context.Context) {
	p.commands <- &command{
		Type:   stop,
		logger: contexts.GetLogger(ctx),
	}
}

func (p *Player) LoopStatus() bool {
	return p.queue.LoopStatus()
}

func (p *Player) SetLoop(ctx context.Context, b bool) {
	p.commands <- &command{
		Type:   loop,
		loop:   b,
		logger: contexts.GetLogger(ctx),
	}
}

func (p *Player) Connect(ctx context.Context, guildID, channelID string) {
	p.commands <- &command{
		Type:      connect,
		guildID:   guildID,
		channelID: channelID,
		logger:    contexts.GetLogger(ctx),
	}
}

func (p *Player) Disconnect(ctx context.Context) {
	p.commands <- &command{
		Type:   disconnect,
		logger: contexts.GetLogger(ctx),
	}
}

func (p *Player) NowPlaying() *pkg.Song {
	p.currentLock.Lock()
	defer p.currentLock.Unlock()
	return p.current
}

func (p *Player) setNowPlaying(s *pkg.Song) {
	p.currentLock.Lock()
	defer p.currentLock.Unlock()
	p.current = s
}

func (p *Player) SongStatus() pkg.SessionStats {
	s := p.audio.Stats()
	if s.Duration == 0 {
		now := p.NowPlaying()
		if now == nil {
			return pkg.SessionStats{}
		}
		s.Duration = now.Duration
	}
	return s
}

func (p *Player) SubscribeOnErrors(h ErrorHandler) {
	p.errorHandlers <- h
}

func (p *Player) processCommands(ctx context.Context) (chan *command, chan error) {
	requests := make(chan *audio.SongRequest)
	playerErrors := p.audio.Process(requests)
	commands := make(chan *command)
	out := make(chan error)
	go func() {
		defer func() {
			close(requests)
			close(out)
			close(commands)
		}()

		for {
			select {
			case c := <-commands:
				if err := p.processCommand(c, requests); err != nil {
					out <- err
				}
			case err := <-playerErrors:
				if err == nil || errors.Is(err, audio.ErrManualStop) || errors.Is(err, io.EOF) {
					go func() {
						p.commands <- &command{Type: next}
					}()
				}
				if err != nil {
					if logError(err) {
						contexts.GetLogger(ctx).Error("audio player", zap.Error(err))
					}
					out <- err
				}
			case <-ctx.Done():
				p.queue.Clear()
				p.audio.Stop()
				return
			}
		}
	}()

	return commands, out
}

func (p *Player) processCommand(c *command, requests chan *audio.SongRequest) error {
	c.logger.Info("process command", zap.String("type", c.Type.String()))
	if c.Type != next {
		p.isWaited = false
	}
	switch c.Type {
	case play:
		return p.processPlay(c.entry, requests, c.logger)
	case next:
		return p.processNext(requests)
	case loop:
		p.queue.SetLoop(c.loop)
	case skip:
		p.audio.Stop()
	case stop:
		p.reset()
	case disconnect:
		p.reset()
		if err := p.voice.Disconnect(); err != nil {
			return err
		}
	case connect:
		return p.processConnect(c.guildID, c.channelID)
	}
	return nil
}

func (p *Player) processPlay(entry *pkg.Song, requests chan *audio.SongRequest, logger *zap.Logger) error {
	if !p.voice.IsConnected() {
		return ErrNotConnected
	}
	logger.Debug("adding to queue", zap.String("title", entry.Title))
	p.queue.Add(entry)
	if !p.audio.IsPlaying() {
		s := p.queue.Next()
		p.setNowPlaying(s)
		logger.Debug("pushing song req")
		requests <- requestFromEntry(s, p.voice.Connection())
	}
	return nil
}

func (p *Player) processNext(out chan *audio.SongRequest) error {
	if !p.voice.IsConnected() {
		p.setNowPlaying(nil)
		return nil
	}
	if p.audio.IsPlaying() {
		return nil
	}
	if s := p.queue.Next(); s != nil {
		p.setNowPlaying(s)
		out <- requestFromEntry(s, p.voice.Connection())
		return nil
	}
	p.setNowPlaying(nil)
	if p.isWaited {
		p.isWaited = false
		err := p.voice.Disconnect()
		if err != nil {
			return errors.Wrap(err, "player: disconnecting because there is nothing to play next")
		}
	} else {
		p.isWaited = true
		p.tryNextAfterTimeout(time.Minute)
	}

	return ErrQueueEmpty
}

func (p *Player) processConnect(gID, cID string) error {
	if p.voice.IsConnected() && p.voice.Connection().GuildID == gID && p.voice.Connection().ChannelID == cID {
		return nil
	}
	p.reset()
	if err := p.voice.Connect(gID, cID); err != nil {
		return errors.Wrapf(err, "connect on gid:%s cid:%s", gID, cID)
	}
	return nil
}

func (p *Player) reset() {
	p.queue.Clear()
	p.audio.Stop()
}

func (p *Player) processErrors(errs <-chan error) chan ErrorHandler {
	handlers := make([]ErrorHandler, 0)
	newHandlers := make(chan ErrorHandler)
	go func() {
		defer close(newHandlers)
		for {
			select {
			case err, ok := <-errs:
				if !ok {
					return
				}
				for _, h := range handlers {
					go h(err)
				}
			case h := <-newHandlers:
				handlers = append(handlers, h)
			}
		}
	}()
	return newHandlers
}

func (p *Player) tryNextAfterTimeout(d time.Duration) {
	go func() {
		time.Sleep(d)
		p.commands <- &command{
			Type: next,
		}
	}()
}

func logError(err error) bool {
	return !errors.Is(err, io.EOF) && !errors.Is(err, audio.ErrManualStop) && !errors.Is(err, ErrQueueEmpty)
}
