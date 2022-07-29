package player

import (
	"github.com/HalvaPovidlo/halvabot-go/internal/pkg/item"
	"sync"

	"github.com/bwmarrin/discordgo"

	"github.com/HalvaPovidlo/halvabot-go/internal/music/audio"
)

type Queue struct {
	entries []*item.Song
	current *item.Song

	loopLock sync.Mutex
	loop     bool
}

func (q *Queue) Next() *item.Song {
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

func (q *Queue) Add(e *item.Song) {
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

func (q *Queue) Front() *item.Song {
	if len(q.entries) == 0 {
		return nil
	}
	return q.entries[0]
}

func requestFromEntry(e *item.Song, connection *discordgo.VoiceConnection) *audio.SongRequest {
	return &audio.SongRequest{
		Voice: connection,
		URI:   e.StreamURL,
	}
}
