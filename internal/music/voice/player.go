package voice

import (
	"context"
	"io"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/voice"
	"github.com/diamondburned/arikawa/v3/voice/udp"
	"github.com/diamondburned/oggreader"
	"github.com/pkg/errors"
)

const (
	frameDuration = 60 // ms
	timeIncrement = 2880
)

var (
	ErrManualStop = errors.New("stop")
	errDisconnect = errors.New("disconnect")
)

type SongRequest struct {
	ID   discord.ChannelID
	Path string
}

type filesCache interface {
	Remove(path string)
}

type Player struct {
	voice *voice.Session
	state *state.State

	files filesCache
	done  chan error

	mu        sync.Mutex
	isPlaying bool
	connected bool
}

func NewPlayer(files filesCache) *Player {
	return &Player{
		files: files,
		done:  make(chan error),
	}
}

func (p *Player) Process(ctx context.Context, requests <-chan *SongRequest) <-chan error {
	playContext, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	out := make(chan error)
	go func() {
		defer close(out)
		for {
			select {
			case r := <-requests:
				if p.IsPlaying() {
					continue
				}
				playContext, cancel = context.WithCancel(context.Background())
				wg.Add(1)
				go func() {
					defer wg.Done()
					err := errors.Wrap(p.play(playContext, r.Path, r.ID), "play audio")
					if playContext.Err() != nil {
						out <- ErrManualStop
					} else {
						out <- err
					}
				}()
			case err := <-p.done:
				cancel()
				wg.Wait()
				if errors.Is(err, errDisconnect) {
					_ = p.voice.Leave(ctx)
					p.setConnected(false)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return out
}

func (p *Player) Stop() {
	if p.IsPlaying() {
		p.done <- ErrManualStop
	}
}

func (p *Player) Disconnect() {
	p.done <- errDisconnect
}

func (p *Player) IsPlaying() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.isPlaying
}

func (p *Player) setPlaying(b bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.isPlaying = b
}

func (p *Player) IsConnected() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.connected
}

func (p *Player) setConnected(b bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.connected = b
}

func (p *Player) play(ctx context.Context, path string, channelID discord.ChannelID) error {
	p.setPlaying(true)
	defer p.setPlaying(false)
	defer p.files.Remove(path)

	p.voice.SetUDPDialer(udp.DialFuncWithFrequency(
		frameDuration*time.Millisecond, // correspond to -frame_duration
		timeIncrement,
	))

	ffmpeg, stdout, err := startFFMPEG(ctx, path)
	if err != nil {
		return errors.Wrap(err, "start ffmpeg")
	}

	if err := p.voice.JoinChannelAndSpeak(ctx, channelID, false, true); err != nil {
		return errors.Wrapf(err, "join channel %s", channelID.String())
	}
	p.setConnected(true)

	if err := oggreader.DecodeBuffered(p.voice, stdout); err != nil {
		return errors.Wrap(err, "failed to decode ogg")
	}

	if err := ffmpeg.Wait(); err != nil {
		return errors.Wrap(err, "failed to finish ffmpeg")
	}

	return nil
}

func startFFMPEG(ctx context.Context, path string) (*exec.Cmd, io.ReadCloser, error) {
	ffmpeg := exec.CommandContext(ctx,
		"ffmpeg", "-hide_banner", "-loglevel", "error",
		"-threads", "1",
		"-i", path,
		"-c:a", "libopus",
		"-b:a", "96k",
		"-frame_duration", strconv.Itoa(frameDuration),
		"-vbr", "off",
		"-f", "opus",
		"-",
	)
	ffmpeg.Stderr = os.Stderr
	stdout, err := ffmpeg.StdoutPipe()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get stdout pipe")
	}
	if err := ffmpeg.Start(); err != nil {
		return nil, nil, errors.Wrap(err, "failed to start ffmpeg")
	}
	return ffmpeg, stdout, nil
}

// TODO: is it necessary?
//func (p *Player) connected(id discord.GuildID) bool {
//	bot, err := p.state.Me()
//	if err != nil {
//		return false
//	}
//	vs, err := p.state.VoiceState(id, bot.ID)
//	if err != nil {
//		return false
//	}
//}
