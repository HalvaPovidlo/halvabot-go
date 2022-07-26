package player

import (
	"sync"

	"github.com/bwmarrin/discordgo"

	"github.com/HalvaPovidlo/halvabot-go/internal/music/audio"
	"github.com/HalvaPovidlo/halvabot-go/internal/pkg"
)

type Queue struct {
	entries []*pkg.Song
	current *pkg.Song

	loopLock sync.Mutex
	loop     bool
}

func (q *Queue) Next() *pkg.Song {
	if q.LoopStatus() {
		return q.current
	}
	if len(q.entries) == 0 {
		return nil
	}
	q.current = q.entries[0]
	q.entries = q.entries[1:]
	return q.current
}

func (q *Queue) Add(e *pkg.Song) {
	q.entries = append(q.entries, e)
}

func (q *Queue) Clear() {
	q.entries = nil
	q.SetLoop(false)
}

func (q *Queue) IsEmpty() bool {
	return len(q.entries) == 0
}

func (q *Queue) SetLoop(b bool) {
	q.loopLock.Lock()
	defer q.loopLock.Unlock()
	q.loop = b
}

func (q *Queue) LoopStatus() bool {
	q.loopLock.Lock()
	defer q.loopLock.Unlock()
	return q.loop
}

func (q *Queue) Front() *pkg.Song {
	if len(q.entries) == 0 {
		return nil
	}
	return q.entries[0]
}

func requestFromEntry(e *pkg.Song, connection *discordgo.VoiceConnection) *audio.SongRequest {
	return &audio.SongRequest{
		Voice: connection,
		URI:   e.StreamURL,
	}
}
