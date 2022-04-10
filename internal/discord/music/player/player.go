package player

import (
	"fmt"
	"github.com/HalvaPovidlo/discordBotGo/cmd/config"
	"github.com/HalvaPovidlo/discordBotGo/internal/discord/search"
	"github.com/HalvaPovidlo/discordBotGo/internal/discord/voice"
	"github.com/HalvaPovidlo/discordBotGo/pkg/zap"
	"github.com/pkg/errors"
)

type Player struct {
	logger  *zap.Logger
	config  config.DiscordConfig
	voice   *voice.Voice
	youtube *search.YouTube
}

func NewPlayer(youtube *search.YouTube, voiceClient *voice.Voice, config config.DiscordConfig, logger *zap.Logger) *Player {
	p := Player{
		youtube: youtube,
		voice:   voiceClient,
		logger:  logger,
		config:  config,
	}
	return &p
}

func (p *Player) PlayYoutube(query string) (*voice.QueueEntry, error) {
	song, err := p.youtube.FindSong(query)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Println("Found song", song.Metadata.Title)
	// TODO: Announce according to debug settings
	p.voice.PlaySong(song, true, p.logger)
	return song, err
}

func (p *Player) Skip() (*voice.QueueEntry, error) {
	qe, err := p.voice.Skip()
	if err != nil {
		return nil, errors.Wrap(err, "skip failed")
	}
	return qe, nil
}

func (p *Player) Stop() error {
	return p.voice.Stop()
}

func (p *Player) Connect(guildID, channelID string) error {
	return p.voice.Connect(guildID, channelID)
}
