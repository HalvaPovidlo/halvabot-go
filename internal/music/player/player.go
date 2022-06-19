package player

import (
	"io"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"

	"github.com/HalvaPovidlo/discordBotGo/internal/audio"
	"github.com/HalvaPovidlo/discordBotGo/internal/pkg"
	"github.com/HalvaPovidlo/discordBotGo/pkg/contexts"
	"github.com/HalvaPovidlo/discordBotGo/pkg/zap"
)

type MediaPlayer interface {
	Process(requests <-chan *audio.SongRequest) <-chan error
	Stats() audio.SessionStats
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

	logger zap.Logger
}

func NewPlayer(ctx contexts.Context, voice VoiceClient, audio MediaPlayer, logger zap.Logger) *Player {
	p := Player{
		logger: logger,
		voice:  voice,
		audio:  audio,
	}
	p.commands, p.errs = p.processCommands(ctx)
	p.errorHandlers = p.processErrors(p.errs)
	return &p
}

// Play next song and enqueue input
func (p *Player) Play(s *pkg.Song) {
	p.commands <- &command{
		Type:  play,
		entry: s,
	}
}

func (p *Player) Skip() {
	p.commands <- &command{
		Type: skip,
	}
}

func (p *Player) Stop() {
	p.commands <- &command{
		Type: stop,
	}
}

func (p *Player) LoopStatus() bool {
	return p.queue.LoopStatus()
}

func (p *Player) SetLoop(b bool) {
	p.commands <- &command{
		Type: loop,
		loop: b,
	}
}

func (p *Player) Connect(guildID, channelID string) {
	p.commands <- &command{
		Type:      connect,
		guildID:   guildID,
		channelID: channelID,
	}
}

func (p *Player) Disconnect() {
	p.commands <- &command{
		Type: disconnect,
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

func (p *Player) SongStatus() audio.SessionStats {
	s := p.audio.Stats()
	if s.Duration == 0 {
		s.Duration = p.NowPlaying().Duration
	}
	return s
}

func (p *Player) SubscribeOnErrors(h ErrorHandler) {
	p.errorHandlers <- h
}

func (p *Player) processCommands(ctx contexts.Context) (chan *command, chan error) {
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
				if err == nil || err == audio.ErrManualStop || err == io.EOF {
					go func() {
						p.commands <- &command{Type: next}
					}()
				}
				if err != nil {
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

func (p *Player) processCommand(c *command, out chan *audio.SongRequest) error {
	p.logger.Infof("process command %s", c.Type)
	if c.Type != next {
		p.isWaited = false
	}
	switch c.Type {
	case play:
		return p.processPlay(c.entry, out)
	case next:
		return p.processNext(out)
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

func (p *Player) processPlay(entry *pkg.Song, out chan *audio.SongRequest) error {
	if !p.voice.IsConnected() {
		return ErrNotConnected.Wrap("voice client not connected")
	}
	p.logger.Debugf("adding to queue %s", entry.Title)
	p.queue.Add(entry)
	if !p.audio.IsPlaying() {
		s := p.queue.Next()
		p.setNowPlaying(s)
		p.logger.Debugf("pushing song req")
		out <- requestFromEntry(s, p.voice.Connection())
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
		return ErrConnectFailed.Wrap(errors.Wrapf(err, "connect on gid:%s cid:%s failed", gID, cID).Error())
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
					h(err)
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
