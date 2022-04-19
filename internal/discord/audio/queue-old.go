package audio

import (
	"math/rand"
	"time"

	"github.com/HalvaPovidlo/discordBotGo/internal/pkg"
)

func (v *Voice) QueueAdd(entry *QueueEntry) {
	v.Entries = append(v.Entries, entry)
}

func (v *Voice) QueueRemove(entry int) {
	v.Entries = append(v.Entries[:entry], v.Entries[entry+1:]...)
}

func (v *Voice) QueueRemoveRange(start, end int) {
	if len(v.Entries) == 0 {
		return
	}

	if start < 0 {
		start = 0
	}
	if end > len(v.Entries) {
		end = len(v.Entries)
	}

	for entry := end; entry < start; entry-- {
		v.QueueRemove(entry)
	}
}

func (v *Voice) QueueClear() {
	v.Entries = nil
}

func (v *Voice) QueueGet(entry int) *QueueEntry {
	if len(v.Entries) < entry {
		return nil
	}

	return v.Entries[entry]
}

func (v *Voice) QueueGetNext() (*QueueEntry, int) {
	if len(v.Entries) == 0 {
		return nil, -1
	}
	index := 0
	if v.Shuffle {
		index = rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(v.Entries))
	}
	return v.Entries[index], index
}

// QueueEntry stores the data about a queue entry
// TODO: Move this structs to some sort of pkg
type QueueEntry pkg.SongRequest

// NowPlaying contains data about the now playing queue entry
type NowPlaying struct {
	Entry    *QueueEntry
	Position time.Duration // The current position in the audio stream
}
