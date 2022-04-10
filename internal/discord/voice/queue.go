package voice

import (
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"
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
type QueueEntry struct {
	Metadata     *Metadata       `json:"metadata"`
	ServiceName  string          `json:"service_name"`  // Name of service used for this queue entry
	ServiceColor int             `json:"service_color"` // Color of service used for this queue entry
	Requester    *discordgo.User `json:"requester"`
}

// NowPlaying contains data about the now playing queue entry
type NowPlaying struct {
	Entry    *QueueEntry
	Position time.Duration // The current position in the audio stream
}

// Metadata stores the metadata of a queue entry
// TODO: Do not pass to http stream_url
type Metadata struct {
	Artists      []MetadataArtist `json:"artists,omitempty"`
	Title        string           `json:"title,omitempty"`
	DisplayURL   string           `json:"display_url,omitempty"`
	StreamURL    string           `json:"stream_url,omitempty"`
	Duration     float64          `json:"duration,omitempty"`
	ArtworkURL   string           `json:"artwork_url,omitempty"`
	ThumbnailURL string           `json:"thumbnail_url,omitempty"`
}

// MetadataArtist stores the data about an artist
type MetadataArtist struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}
