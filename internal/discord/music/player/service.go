package player

import (
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"

	"github.com/HalvaPovidlo/discordBotGo/internal/discord/audio"
	"github.com/HalvaPovidlo/discordBotGo/internal/pkg"
	"github.com/HalvaPovidlo/discordBotGo/pkg/contexts"
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

type ErrorHandler interface {
	HandleError(err error)
}

var (
	ErrNotConnected   = errors.New("voice client not connected")
	ErrConnectFailed  = errors.New("connect to voice channel failed")
	ErrNotImplemented = errors.New("this command is not implemented")
)

type commandType int

const (
	play commandType = iota
	next
	skip
	stop
	radio
	connect
	disconnect
	enqueue
	// shuffle
	loop
)

type command struct {
	Type      commandType
	guildID   string
	channelID string
	entry     *pkg.SongRequest
	loop      bool
}

type Config struct {
}

// Player all public methods are concurrent and
// most private methods are designed to be synchronous
type Player struct {
	ctx    contexts.Context
	voice  VoiceClient
	audio  MediaPlayer
	config Config

	currentLock   sync.Mutex
	current       pkg.SongRequest
	isWaited      bool
	queue         Queue
	errs          chan error
	commands      chan *command
	errorHandlers chan ErrorHandler
}

func NewPlayer(ctx contexts.Context, voice VoiceClient, audio MediaPlayer, config Config) *Player {
	p := Player{
		ctx:    ctx,
		voice:  voice,
		audio:  audio,
		config: config,
	}
	p.commands, p.errs = p.processCommands()
	p.errorHandlers = p.processErrors(p.errs)
	return &p
}

// Play next song and enqueue input
func (p *Player) Play(s *pkg.SongRequest) {
	log := p.ctx.LoggerFromContext()
	log.Debug("sending command play")
	p.commands <- &command{
		Type:  play,
		entry: s,
	}
}

// Enqueue song to the player
func (p *Player) Enqueue(s *pkg.SongRequest) {
	p.commands <- &command{
		Type:  enqueue,
		entry: s,
	}
}

// Skip returns nil if there is nothing to play next
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

// Radio TODO: implement radio
func (p *Player) Radio() bool {
	p.commands <- &command{
		Type: radio,
	}
	return false
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

func (p *Player) NowPlaying() pkg.SongRequest {
	p.currentLock.Lock()
	defer p.currentLock.Unlock()
	return p.current
}

func (p *Player) setNowPlaying(s pkg.SongRequest) {
	p.currentLock.Lock()
	defer p.currentLock.Unlock()
	p.current = s
}

func (p *Player) Stats() audio.SessionStats {
	s := p.audio.Stats()
	if s.Duration == 0 {
		s.Duration = p.NowPlaying().Metadata.Duration
	}
	return s
}

// SubscribeOnErrors TODO: try to path functions not objects
func (p *Player) SubscribeOnErrors(h ErrorHandler) {
	p.errorHandlers <- h
}

func (p *Player) processCommands() (chan *command, chan error) {
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
				p.setNowPlaying(pkg.SongRequest{})
				if err == nil || err == audio.ErrManualStop {
					go func() {
						p.commands <- &command{Type: next}
					}()
				}
				if err != nil {
					p.ctx.LoggerFromContext().Error(errors.Wrap(err, "player"))
					out <- err
				}
			case <-p.ctx.Done():
				p.queue.Clear()
				p.audio.Stop()
				return
			}
		}
	}()

	return commands, out
}

func (p *Player) processCommand(c *command, out chan *audio.SongRequest) error {
	log := p.ctx.LoggerFromContext()
	log.Debugf("process command %d", c.Type)
	if c.Type != next {
		p.isWaited = false
	}

	switch c.Type {
	case play:
		return p.processPlay(c.entry, out)
	case next:
		return p.processNext(out)
	case enqueue:
		p.queue.Add(c.entry)
	case loop:
		p.queue.SetLoop(c.loop)
	case skip:
		p.audio.Stop()
	case radio:
		return ErrNotImplemented
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

func (p *Player) processPlay(entry *pkg.SongRequest, out chan *audio.SongRequest) error {
	log := p.ctx.LoggerFromContext()
	if !p.voice.IsConnected() {
		return ErrNotConnected
	}
	log.Debugf("adding to queue %s", entry.Metadata.Title)
	p.queue.Add(entry)
	if !p.audio.IsPlaying() {
		s := p.queue.Next()
		p.setNowPlaying(*s)
		log.Debugf("pushing song req")
		out <- requestFromEntry(s, p.voice.Connection())
	}
	return nil
}

func (p *Player) processNext(out chan *audio.SongRequest) error {
	if !p.voice.IsConnected() {
		return nil
	}
	if !p.audio.IsPlaying() {
		if !p.queue.IsEmpty() {
			s := p.queue.Next()
			p.setNowPlaying(*s)
			out <- requestFromEntry(s, p.voice.Connection())
		} else {
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
		}
	}
	return nil
}

func (p *Player) processConnect(gID, cID string) error {
	if p.voice.IsConnected() && p.voice.Connection().GuildID == gID && p.voice.Connection().ChannelID == cID {
		return nil
	}
	p.reset()
	if err := p.voice.Connect(gID, cID); err != nil {
		p.ctx.LoggerFromContext().Errorw("player command connect",
			"err", err,
			"guildID", gID,
			"channelID", cID)
		return ErrConnectFailed
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
					h.HandleError(err)
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
