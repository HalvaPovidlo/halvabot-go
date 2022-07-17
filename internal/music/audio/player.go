package audio

import (
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/khodand/dca"
	"github.com/pkg/errors"

	"github.com/HalvaPovidlo/discordBotGo/internal/pkg"
)

var (
	ErrManualStop = errors.New("stop")
)

type SongRequest struct {
	Voice *discordgo.VoiceConnection
	URI   string
}

type Player struct {
	Options *dca.EncodeOptions `json:"encodingOptions"`
	done    chan error

	isPlayingLock sync.Mutex
	isPlaying     bool

	statsLock sync.Mutex
	stats     pkg.SessionStats
}

func NewPlayer(options *dca.EncodeOptions) *Player {
	return &Player{
		Options: options,
		done:    make(chan error),
	}
}

func (p *Player) Process(requests <-chan *SongRequest) <-chan error {
	out := make(chan error)
	go func() {
		defer close(out)
		for req := range requests {
			err := p.play(req.Voice, req.URI)
			out <- err
		}
	}()
	return out
}

func (p *Player) Stop() {
	if p.IsPlaying() {
		p.done <- ErrManualStop
	}
}

func (p *Player) Stats() pkg.SessionStats {
	p.statsLock.Lock()
	defer p.statsLock.Unlock()
	return p.stats
}

func (p *Player) setStatsDuration(d time.Duration) {
	p.statsLock.Lock()
	defer p.statsLock.Unlock()
	p.stats.Duration = d.Seconds()
}

func (p *Player) setStatsPos(d time.Duration) {
	p.statsLock.Lock()
	defer p.statsLock.Unlock()
	p.stats.Pos = d.Seconds()
}

func (p *Player) IsPlaying() bool {
	p.isPlayingLock.Lock()
	defer p.isPlayingLock.Unlock()
	return p.isPlaying
}

func (p *Player) setPlaying(b bool) {
	p.isPlayingLock.Lock()
	defer p.isPlayingLock.Unlock()
	p.isPlaying = b
}

func (p *Player) play(v *discordgo.VoiceConnection, uri string) error {
	if v == nil {
		return errors.New("voice connection doesn't exists")
	}
	err := v.Speaking(true)
	if err != nil {
		return errors.Wrap(err, "set speaking true")
	}
	p.setPlaying(true)

	encodeSession, err := dca.EncodeFile(uri, p.Options)
	if err != nil {
		return errors.Wrapf(err, "encode %s", uri)
	}
	defer encodeSession.Cleanup()

	p.setStatsDuration(encodeSession.Stats().Duration)

	stream := dca.NewStream(encodeSession, v, p.done)
	err = p.updatePosition(stream)
	p.setPlaying(false)
	_ = v.Speaking(false)
	return err
}

func (p *Player) updatePosition(stream *dca.StreamingSession) error {
	ticker := time.NewTicker(500 * time.Millisecond)
	for {
		select {
		case err := <-p.done:
			stream.SetPaused(true)
			return err
		case <-ticker.C:
			p.setStatsPos(stream.PlaybackPosition())
		}
	}
}
