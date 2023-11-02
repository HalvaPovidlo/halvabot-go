package player

import (
	"context"
	"io"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/HalvaPovidlo/halvabot-go/internal/music/voice"
	"github.com/HalvaPovidlo/halvabot-go/internal/pkg/item"
	"github.com/HalvaPovidlo/halvabot-go/pkg/contexts"
	"github.com/HalvaPovidlo/halvabot-go/pkg/log"
)

func (p *Service) processCommands(ctx context.Context) (chan *command, chan error) {
	requests := make(chan *voice.SongRequest)
	playerErrors := p.audio.Process(ctx, requests)
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
				if err == nil || errors.Is(err, voice.ErrManualStop) || errors.Is(err, io.EOF) {
					go func() {
						p.commands <- &command{Type: next}
					}()
				}
				if err != nil {
					if logError(err) {
						contexts.GetLogger(ctx).Error("voice player", zap.Error(err))
					}
					out <- err
				}
			case <-ctx.Done():
				p.queue.Clear()
				p.audio.Disconnect()
				return
			}
		}
	}()

	return commands, out
}

func (p *Service) processCommand(c *command, requests chan *voice.SongRequest) error {
	if c.logger == nil {
		c.logger = log.NewLogger(false)
	}
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
	case disconnect:
		p.reset()
		p.audio.Disconnect()
	}
	return nil
}

func (p *Service) processPlay(entry *item.Song, requests chan *voice.SongRequest, logger *zap.Logger) error {
	if !p.audio.IsConnected() {
		return ErrNotConnected
	}
	logger.Debug("adding to queue", zap.String("title", entry.Title))
	p.queue.Add(entry)
	if !p.audio.IsPlaying() {
		s := p.queue.Next()
		p.setNowPlaying(s)
		logger.Debug("pushing song req")
		requests <- buildRequest(s, p.voice.Connection())
	}
	return nil
}

func (p *Service) processNext(out chan *voice.SongRequest) error {
	if !p.audio.IsConnected() {
		p.setNowPlaying(nil)
		return nil
	}
	if p.audio.IsPlaying() {
		return nil
	}
	if s := p.queue.Next(); s != nil {
		p.setNowPlaying(s)
		out <- buildRequest(s, p.voice.Connection())
		return nil
	}
	p.setNowPlaying(nil)
	if p.isWaited {
		p.isWaited = false
		err := p.audio.Disconnect()
		if err != nil {
			return errors.Wrap(err, "player: disconnecting because there is nothing to play next")
		}
	} else {
		p.isWaited = true
		p.tryNextAfterTimeout(time.Minute)
	}

	return ErrQueueEmpty
}

func (p *Service) reset() {
	p.queue.Clear()
	p.audio.Stop()
}

func (p *Service) processErrors(errs <-chan error) chan ErrorHandler {
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

func (p *Service) tryNextAfterTimeout(d time.Duration) {
	go func() {
		time.Sleep(d)
		p.commands <- &command{
			Type: next,
		}
	}()
}

func logError(err error) bool {
	return !errors.Is(err, io.EOF) && !errors.Is(err, voice.ErrManualStop) && !errors.Is(err, ErrQueueEmpty)
}
