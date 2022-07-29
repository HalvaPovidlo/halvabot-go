package player

import (
	"context"
	"github.com/HalvaPovidlo/halvabot-go/internal/pkg/item"
	"sync"
	"time"
)

type MockPlayer struct {
	statusMx    sync.Mutex
	loopStatus  bool
	radioStatus bool
}

func (m *MockPlayer) Play(ctx context.Context, query, userID, guildID, channelID string) (*item.Song, error) {
	song := &item.Song{
		Title:        query,
		URL:          "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		Service:      "youtube",
		ArtistName:   "Rick Astley",
		ArtistURL:    "https://www.youtube.com/channel/UCuAXFkgsw1L7xaCfnd5JJOw",
		ArtworkURL:   "https://s.namemc.com/3d/skin/body.png?id=d4347e67364ad441&model=slim&width=282&height=282",
		ThumbnailURL: "https://s.namemc.com/3d/skin/body.png?id=d4347e67364ad441&model=slim&width=282&height=282",
		Playbacks:    10,
		LastPlay:     time.Now(),
		ID: item.SongID{
			ID:      "dQw4w9WgXcQ",
			Service: "youtube",
		},
		Requester: nil,
		StreamURL: "",
		Duration:  212,
	}
	return song, nil
}

func (m *MockPlayer) Skip(ctx context.Context) {}

func (m *MockPlayer) SetLoop(ctx context.Context, b bool) {
	m.statusMx.Lock()
	m.loopStatus = b
	m.statusMx.Unlock()
}
func (m *MockPlayer) LoopStatus() bool {
	m.statusMx.Lock()
	b := m.loopStatus
	m.statusMx.Unlock()
	return b
}

func (m *MockPlayer) NowPlaying() *item.Song {
	return &item.Song{
		Title:        "Mock song",
		URL:          "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		Service:      "youtube",
		ArtistName:   "Rick Astley",
		ArtistURL:    "https://www.youtube.com/channel/UCuAXFkgsw1L7xaCfnd5JJOw",
		ArtworkURL:   "https://s.namemc.com/3d/skin/body.png?id=d4347e67364ad441&model=slim&width=282&height=282",
		ThumbnailURL: "https://s.namemc.com/3d/skin/body.png?id=d4347e67364ad441&model=slim&width=282&height=282",
		Playbacks:    10,
		LastPlay:     time.Now(),
		ID: item.SongID{
			ID:      "dQw4w9WgXcQ",
			Service: "youtube",
		},
		Requester: nil,
		StreamURL: "",
		Duration:  212,
	}
}

func (m *MockPlayer) SongStatus() item.SessionStats {
	return item.SessionStats{
		Pos:      111,
		Duration: 212,
	}
}

func (m *MockPlayer) SetRadio(ctx context.Context, b bool, guildID, channelID string) error {
	m.statusMx.Lock()
	m.radioStatus = b
	m.statusMx.Unlock()
	return nil
}

func (m *MockPlayer) RadioStatus() bool {
	m.statusMx.Lock()
	b := m.radioStatus
	m.statusMx.Unlock()
	return b
}

func (m *MockPlayer) Status() item.PlayerStatus {
	return item.PlayerStatus{
		Loop:  m.LoopStatus(),
		Radio: m.RadioStatus(),
		Song:  m.SongStatus(),
		Now:   m.NowPlaying(),
	}
}
