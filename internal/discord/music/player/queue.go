package player

import (
	"sync"

	"github.com/bwmarrin/discordgo"

	"github.com/HalvaPovidlo/discordBotGo/internal/discord/audio"
	"github.com/HalvaPovidlo/discordBotGo/internal/discord/pkg"
)

type Queue struct {
	entries []*pkg.SongRequest

	loopLock sync.Mutex
	loop     bool
}

func (q *Queue) Next() *pkg.SongRequest {
	e := q.entries[0]
	if q.LoopStatus() {
		return e
	}
	q.entries = q.entries[1:]
	return e
}

func (q *Queue) Add(e *pkg.SongRequest) {
	q.entries = append(q.entries, e)
}

func (q *Queue) Clear() {
	q.entries = nil
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

func (q *Queue) Front() *pkg.SongRequest {
	if len(q.entries) == 0 {
		return nil
	}
	return q.entries[0]
}

func requestFromEntry(e *pkg.SongRequest, connection *discordgo.VoiceConnection) *audio.SongRequest {
	return &audio.SongRequest{
		Voice: connection,
		URI:   e.Metadata.StreamURL,
	}
}
