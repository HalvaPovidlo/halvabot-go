package player

import (
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"

	"github.com/HalvaPovidlo/discordBotGo/internal/discord/audio"
	"github.com/HalvaPovidlo/discordBotGo/internal/discord/pkg"
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

type YouTube interface {
	FindSong(query string) (*pkg.SongRequest, error)
}

var (
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
)

type command struct {
	Type      commandType
	guildID   string
	channelID string
	entry     *pkg.SongRequest
}

type Config struct {
}

type Player struct {
	ctx     contexts.Context
	youtube YouTube
	voice   VoiceClient
	audio   MediaPlayer
	config  Config

	currentLock sync.Mutex
	current     pkg.SongRequest
	queue       Queue
	errs        chan error
	commands    chan *command
}

func NewPlayer(ctx contexts.Context, youtube YouTube, voice VoiceClient, audio MediaPlayer, config Config) *Player {
	p := Player{
		ctx:     ctx,
		youtube: youtube,
		voice:   voice,
		audio:   audio,
		config:  config,
	}
	p.commands, p.errs = p.processCommands()
	return &p
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
					// out <- err
				}
			case <-p.ctx.Done():
				p.queue.Clear()
				p.audio.Stop()
			}
		}
	}()

	return commands, out
}

func (p *Player) processCommand(c *command, out chan *audio.SongRequest) error {
	log := p.ctx.LoggerFromContext()
	log.Debugf("process command %d", c.Type)
	switch c.Type {
	case play:
		log.Debugf("adding to queue")
		p.queue.Add(c.entry)
		if !p.audio.IsPlaying() {
			s := p.queue.Next()
			p.setNowPlaying(*s)
			log.Debugf("pushing song req")
			out <- requestFromEntry(s, p.voice.Connection())
		}
	case next:
		if !p.audio.IsPlaying() {
			if !p.queue.IsEmpty() {
				s := p.queue.Next()
				p.setNowPlaying(*s)
				out <- requestFromEntry(s, p.voice.Connection())
			} else {
				err := p.voice.Disconnect()
				if err != nil {
					log.Error(errors.Wrap(err, "player: disconnecting because there is nothing to play next"))
				}
			}
		}
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
		if p.voice.IsConnected() && p.voice.Connection().GuildID == c.guildID && p.voice.Connection().ChannelID == c.channelID {
			return nil
		}
		p.reset()
		if err := p.voice.Connect(c.guildID, c.channelID); err != nil {
			p.ctx.LoggerFromContext().Errorw("player command connect",
				"err", err,
				"guildID", c.guildID,
				"channelID", c.channelID)
			return ErrConnectFailed
		}
	}
	return nil
}

func (p *Player) reset() {
	p.queue.Clear()
	p.audio.Stop()
}

func (p *Player) PlayYoutube(query string) (*pkg.SongRequest, error) {
	log := p.ctx.LoggerFromContext()
	log.Debug("Finding song")
	song, err := p.youtube.FindSong(query)
	if err != nil {
		return nil, errors.Wrap(err, "player: youtube song not found")
	}
	log.Debug("sending command play")
	p.commands <- &command{
		Type:  play,
		entry: song,
	}
	return song, nil
}

// Skip returns nil if there is nothing to play next
func (p *Player) Skip() *pkg.SongRequest {
	next := p.queue.Front()
	p.commands <- &command{
		Type: skip,
	}
	return next
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
	p.queue.SetLoop(b)
}

func (p *Player) Stats() audio.SessionStats {
	return p.audio.Stats()
}

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

// Errors TODO: make smth like subscribe on errors function to prevent deadlocks and implement possibility of multiple error readers
func (p *Player) Errors() <-chan error {
	return p.errs
}
