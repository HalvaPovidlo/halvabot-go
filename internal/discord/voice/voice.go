package voice

import (
	"fmt"
	"io"
	"net/url"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"github.com/pkg/errors"

	"github.com/HalvaPovidlo/discordBotGo/cmd/config"
	"github.com/HalvaPovidlo/discordBotGo/pkg/zap"
)

type RepeatLevel int

const (
	RepeatNone RepeatLevel = iota
	RepeatPlaylist
	RepeatNowPlaying
)

var (
	// errVoiceJoinAlreadyInChannel = errors.New("voice: error joining channel, already in selected voice channel")
	// errVoiceJoinBusy             = errors.New("voice: error joining channel, busy in another channel")
	// errVoiceJoinChannel          = errors.New("voice: error joining channel")
	// errVoiceJoinChangeChannel    = errors.New("voice: error changing channel")
	// errVoiceLeaveChannel         = errors.New("voice: error leaving channel")
	errVoiceLeaveNotConnected = errors.New("voice: error leaving channel, not connected")
	errVoiceNotStreaming      = errors.New("voice: not streaming")
	// errVoicePausedAlready        = errors.New("voice: already paused")
	errVoicePlayAlreadyStreaming = errors.New("voice: error playing audio, already streaming")
	// errVoicePlayInvalidURL       = errors.New("voice: error playing audio, invalid URL")
	errVoicePlayMuted        = errors.New("voice: error playing audio, muted")
	errVoicePlayNotConnected = errors.New("voice: error playing audio, not connected")
	// errVoicePlayingAlready       = errors.New("voice: already playing")
	errVoiceSkippedManually = errors.New("voice: skipped audio manually")
	errVoiceStoppedManually = errors.New("voice: stopped audio manually")
)

// Voice contains data about the current voice session
type Voice struct {
	sync.Mutex `json:"-"` // This struct gets accessed very repeatedly throughout various goroutines so we need a mutex to prevent race conditions

	// Voice connections and audio sessions
	VoiceConnection  *discordgo.VoiceConnection `json:"voiceConnection"`
	DiscordSession   *discordgo.Session
	EncodingSession  *dca.EncodeSession    `json:"encodingSession"`
	StreamingSession *dca.StreamingSession `json:"streamingSession"`

	// Voice configurations
	EncodingOptions *dca.EncodeOptions `json:"encodingOptions"`
	RepeatLevel     RepeatLevel        `json:"repeatLevel"`
	Shuffle         bool               `json:"shuffle"`
	Muted           bool               `json:"muted"`    // Whether audio should be sent to Discord
	Deafened        bool               `json:"deafened"` // Whether audio should be received from Discord

	// Contains data about the current queue
	Entries    []*QueueEntry `json:"queueEntries"`
	NowPlaying *NowPlaying   `json:"nowPlaying"`

	// Miscellaneous
	TextChannelID string     `json:"textChannelID"` // TODO: The channel that was last used to interact with the voice session
	Started       bool       `json:"-"`             // If the playback session has started
	done          chan error // Used to signal when streaming is done or other actions are performed
}

// Connect connects to a given voice channel
func (v *Voice) Connect(guildID, vChannelID string) error {
	v.Lock()
	defer v.Unlock()

	if v.IsConnected() {
		err := v.VoiceConnection.ChangeChannel(vChannelID, v.Muted, v.Deafened)
		if err != nil {
			return err
		}

		return nil
	}

	fmt.Println("ChannelVoiceJoin")
	voiceConnection, err := v.DiscordSession.ChannelVoiceJoin(guildID, vChannelID, v.Muted, v.Deafened)
	if err != nil {
		return err
	}

	v.VoiceConnection = voiceConnection
	return nil
}

// Disconnect disconnects from the current voice channel
func (v *Voice) Disconnect() error {
	v.Lock()
	defer v.Unlock()

	if v.IsConnected() {
		err := v.VoiceConnection.Disconnect()
		v.VoiceConnection = nil
		v.Started = false
		return err
	}

	return errVoiceLeaveNotConnected
}

func (v *Voice) Play(logger *zap.Logger) error {
	if !v.IsConnected() {
		return errVoicePlayNotConnected
	}
	if v.IsStreaming() {
		return errVoicePlayAlreadyStreaming
	}
	nextQueueEntry, index := v.QueueGetNext()
	v.QueueRemove(index)
	v.PlaySong(nextQueueEntry, false, logger)
	return nil
}

// PlaySong plays a given queue entry in a connected voice channel
// - queueEntry: The queue entry to play/add to the queue
// - announceQueueAdded: Whether to announce a queue added message if something is already playing (used internally for mass playlist additions)
func (v *Voice) PlaySong(queueEntry *QueueEntry, announceQueueAdded bool, logger *zap.Logger) {
	go v.playSong(queueEntry, announceQueueAdded, logger)
}

func (v *Voice) playSong(queueEntry *QueueEntry, announceQueueAdded bool, logger *zap.Logger) {
	if !v.IsConnected() {
		logger.Error(errVoicePlayNotConnected)
		return
	}

	if v.IsStreaming() {
		v.QueueAdd(queueEntry)
		if announceQueueAdded {
			// TODO: send message play
			logger.Info("Added to queue")
		}
		return
	}

	if v.Muted {
		logger.Error(errVoicePlayMuted)
		return
	}

	v.Lock()

	v.Started = true
	v.NowPlaying = &NowPlaying{Entry: queueEntry}
	updateListeningStatus(v.DiscordSession, v.NowPlaying.Entry.Metadata.Artists[0].Name, v.NowPlaying.Entry.Metadata.Title)

	v.done = make(chan error)
	v.Unlock()
	logger.Infow("Start playing",
		"streamURL", v.NowPlaying.Entry.Metadata.StreamURL)
	msg, err := v.playRaw(v.NowPlaying.Entry.Metadata.StreamURL)

	if msg != nil {
		if msg == errVoiceStoppedManually {
			v.Started = false
			logger.Info(errVoiceStoppedManually)
			return
		}
	}

	if err != nil {
		v.Started = false
		switch err {
		case io.ErrUnexpectedEOF:
			if msg != errVoiceSkippedManually {
				logger.Error(err)
				return
			}
		default:
			logger.Error(err)
			return
		}
	}

	var nextQueueEntry *QueueEntry
	index := 0

	switch v.RepeatLevel {
	case RepeatNone:
		v.NowPlaying = nil
		if len(v.Entries) == 0 {
			err := v.Disconnect()
			if err != nil {
				logger.Error(err)
			}
			return
		}
		nextQueueEntry, index = v.QueueGetNext()
		v.QueueRemove(index)
	case RepeatPlaylist:
		v.QueueAdd(v.NowPlaying.Entry)
		nextQueueEntry, index = v.QueueGetNext()
		v.QueueRemove(index)
	case RepeatNowPlaying:
		nextQueueEntry = v.NowPlaying.Entry
	}

	v.NowPlaying = nil
	v.playSong(nextQueueEntry, announceQueueAdded, logger)
}

func updateListeningStatus(session *discordgo.Session, name string, title string) {
	err := session.UpdateListeningStatus(name + " - " + title)
	if err != nil {
		return
	}
}

// playRaw plays a given media URL in a connected voice channel
func (v *Voice) playRaw(mediaURL string) (error, error) {
	if !v.IsConnected() {
		return nil, nil
	}
	if v.IsStreaming() {
		return nil, nil
	}

	v.Lock()

	if v.Muted {
		return nil, nil
	}

	_, err := url.ParseRequestURI(mediaURL)
	if err != nil {
		return nil, nil
	}

	fmt.Println("dca.EncodeFile")
	v.EncodingSession, err = dca.EncodeFile(mediaURL, v.EncodingOptions)
	if err != nil {
		return nil, err
	}
	v.speaking()
	fmt.Println("dca.NewStream")
	v.StreamingSession = dca.NewStream(v.EncodingSession, v.VoiceConnection, v.done)
	v.Unlock()

	go v.updatePosition()
	msg := <-v.done

	v.Lock()

	v.silent()
	_, err = v.StreamingSession.Finished()
	v.StreamingSession = nil

	v.EncodingSession.Stop()
	v.EncodingSession.Cleanup()
	v.EncodingSession = nil

	v.Unlock()

	return msg, err
}

// updatePosition updates the current position of a playing media
func (v *Voice) updatePosition() {
	for {
		v.Lock()

		if v.StreamingSession == nil || v.NowPlaying == nil {
			v.Unlock()
			return
		}
		v.NowPlaying.Position = v.StreamingSession.PlaybackPosition()

		v.Unlock()
	}
}

// Stop stops the playback of a media
func (v *Voice) Stop() error {
	if !v.IsStreaming() {
		return errVoiceNotStreaming
	}

	v.done <- errVoiceStoppedManually

	if err := v.EncodingSession.Stop(); err != nil {
		return err
	}
	v.EncodingSession.Cleanup()

	return nil
}

// Skip stops the encoding session of a playing media, allowing the play wrapper to continue to the next media in a queue
func (v *Voice) Skip() (*QueueEntry, error) {
	if !v.IsStreaming() {
		return nil, errVoiceNotStreaming
	}

	nextQueueEntry, _ := v.QueueGetNext()
	v.done <- errVoiceSkippedManually
	if err := v.EncodingSession.Stop(); err != nil {
		return nil, err
	}
	v.EncodingSession.Cleanup()

	return nextQueueEntry, nil
}

// speaking allows the sending of audio to Discord
func (v *Voice) speaking() {
	if v.IsConnected() {
		v.VoiceConnection.Speaking(true)
	}
}

// silent prevents the sending of audio to Discord
func (v *Voice) silent() {
	if v.IsConnected() {
		v.VoiceConnection.Speaking(false)
	}
}

// IsConnected returns whether a voice connection exists
func (v *Voice) IsConnected() bool {
	if v == nil {
		return false
	}

	return v.VoiceConnection != nil
}

// IsStreaming returns whether a media is playing
func (v *Voice) IsStreaming() bool {
	v.Lock()
	defer v.Unlock()

	return !(!v.IsConnected() || !v.Started || v.StreamingSession == nil || v.EncodingSession == nil)
}

// SetTextChannel sets the text channel to send messages to
func (v *Voice) SetTextChannel(tChannelID string) {
	v.TextChannelID = tChannelID
}

// NewVoice VoiceInit initializes a voice object for the given guild
func NewVoice(session *discordgo.Session, config config.VoiceConfig) *Voice {
	return &Voice{
		DiscordSession:  session,
		EncodingOptions: &config.EncodeOptions,
	}
}
