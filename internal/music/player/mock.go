package player

import (
	"sync"
	"time"

	"github.com/HalvaPovidlo/discordBotGo/internal/music/audio"
	"github.com/HalvaPovidlo/discordBotGo/internal/pkg"
	"github.com/HalvaPovidlo/discordBotGo/pkg/contexts"
)

type MockPlayer struct {
	statusMx    sync.Mutex
	loopStatus  bool
	radioStatus bool
}

func (m *MockPlayer) Play(ctx contexts.Context, query, userID, guildID, channelID string) (*pkg.Song, int, error) {
	song := &pkg.Song{
		Title:        query,
		URL:          "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		Service:      "youtube",
		ArtistName:   "Rick Astley",
		ArtistURL:    "https://www.youtube.com/channel/UCuAXFkgsw1L7xaCfnd5JJOw",
		ArtworkURL:   "https://s.namemc.com/3d/skin/body.png?id=d4347e67364ad441&model=slim&width=282&height=282",
		ThumbnailURL: "https://s.namemc.com/3d/skin/body.png?id=d4347e67364ad441&model=slim&width=282&height=282",
		Playbacks:    10,
		LastPlay:     pkg.PlayDate{Time: time.Now().Add(-35 * time.Hour)},
		ID: pkg.SongID{
			ID:      "dQw4w9WgXcQ",
			Service: "youtube",
		},
		Requester: nil,
		StreamURL: "",
		Duration:  212,
	}
	return song, 11, nil
}

func (m *MockPlayer) Skip() {

}

func (m *MockPlayer) SetLoop(b bool) {
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

func (m *MockPlayer) NowPlaying() *pkg.Song {
	return &pkg.Song{
		Title:        "Mock song",
		URL:          "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		Service:      "youtube",
		ArtistName:   "Rick Astley",
		ArtistURL:    "https://www.youtube.com/channel/UCuAXFkgsw1L7xaCfnd5JJOw",
		ArtworkURL:   "https://s.namemc.com/3d/skin/body.png?id=d4347e67364ad441&model=slim&width=282&height=282",
		ThumbnailURL: "https://s.namemc.com/3d/skin/body.png?id=d4347e67364ad441&model=slim&width=282&height=282",
		Playbacks:    10,
		LastPlay:     pkg.PlayDate{Time: time.Now().Add(-35 * time.Hour)},
		ID: pkg.SongID{
			ID:      "dQw4w9WgXcQ",
			Service: "youtube",
		},
		Requester: nil,
		StreamURL: "",
		Duration:  212,
	}
}

func (m *MockPlayer) SongStatus() audio.SessionStats {
	return audio.SessionStats{
		Pos:      111,
		Duration: 212,
	}
}

func (m *MockPlayer) SetRadio(ctx contexts.Context, b bool, guildID, channelID string) error {
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

func (m *MockPlayer) Status() pkg.PlayerStatus {
	return pkg.PlayerStatus{
		Loop:  m.LoopStatus(),
		Radio: m.RadioStatus(),
		Now:   m.NowPlaying(),
	}
}
