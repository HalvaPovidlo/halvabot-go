package voice

import (
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"
)

func (voice *Voice) QueueAdd(entry *QueueEntry) {
	//Add the new queue entry
	voice.Entries = append(voice.Entries, entry)
}
func (voice *Voice) QueueRemove(entry int) {
	//Remove the queue entry
	voice.Entries = append(voice.Entries[:entry], voice.Entries[entry+1:]...)
}
func (voice *Voice) QueueRemoveRange(start, end int) {
	if len(voice.Entries) == 0 {
		return
	}

	if start < 0 {
		start = 0
	}
	if end > len(voice.Entries) {
		end = len(voice.Entries)
	}

	for entry := end; entry < start; entry-- {
		voice.QueueRemove(entry)
	}
}
func (voice *Voice) QueueClear() {
	voice.Entries = nil
}
func (voice *Voice) QueueGet(entry int) *QueueEntry {
	if len(voice.Entries) < entry {
		return nil
	}

	return voice.Entries[entry]
}
func (voice *Voice) QueueGetNext() (entry *QueueEntry, index int) {
	if len(voice.Entries) == 0 {
		return nil, -1
	}
	if voice.Shuffle {
		index = rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(voice.Entries))
	}
	entry = voice.Entries[index]
	return voice.Entries[index], index
}

// QueueEntry stores the data about a queue entry
type QueueEntry struct {
	Metadata     *Metadata       `json:"metadata"`      //Queue entry metadata
	ServiceName  string          `json:"service_name"`  //Name of service used for this queue entry
	ServiceColor int             `json:"service_color"` //Color of service used for this queue entry
	Requester    *discordgo.User `json:"requester"`
}

//VoiceNowPlaying contains data about the now playing queue entry
type VoiceNowPlaying struct {
	Entry    *QueueEntry   //The underlying queue entry
	Position time.Duration //The current position in the audio stream
}

// Metadata stores the metadata of a queue entry
type Metadata struct {
	Artists      []MetadataArtist `json:"artists,omitempty"`       //List of artists for this queue entry
	Title        string           `json:"title,omitempty"`         //Entry title
	DisplayURL   string           `json:"display_url,omitempty"`   //Entry page URL to display to users
	StreamURL    string           `json:"stream_url,omitempty"`    //Entry URL for streaming
	Duration     float64          `json:"duration,omitempty"`      //Entry duration
	ArtworkURL   string           `json:"artwork_url,omitempty"`   //Entry artwork URL
	ThumbnailURL string           `json:"thumbnail_url,omitempty"` //Entry artwork thumbnail URL
}

// MetadataArtist stores the data about an artist
type MetadataArtist struct {
	Name string `json:"name,omitempty"` //Artist name
	URL  string `json:"url,omitempty"`  //Artist page URL
}
