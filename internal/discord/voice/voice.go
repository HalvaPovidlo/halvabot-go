package voice

import (
	"fmt"
	"github.com/HalvaPovidlo/discordBotGo/cmd/config"
	"github.com/pkg/errors"
	"hash"
	"hash/fnv"
	"io"
	"net/url"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
)

type RepeatLevel int

const (
	RepeatNone RepeatLevel = iota
	RepeatPlaylist
	RepeatNowPlaying
)

var (
	fnvHash hash.Hash32 = fnv.New32a()

	errVoiceJoinAlreadyInChannel = errors.New("voice: error joining channel, already in selected voice channel")
	errVoiceJoinBusy             = errors.New("voice: error joining channel, busy in another channel")
	errVoiceJoinChannel          = errors.New("voice: error joining channel")
	errVoiceJoinChangeChannel    = errors.New("voice: error changing channel")
	errVoiceLeaveChannel         = errors.New("voice: error leaving channel")
	errVoiceLeaveNotConnected    = errors.New("voice: error leaving channel, not connected")
	errVoiceNotStreaming         = errors.New("voice: not streaming")
	errVoicePausedAlready        = errors.New("voice: already paused")
	errVoicePlayAlreadyStreaming = errors.New("voice: error playing audio, already streaming")
	errVoicePlayInvalidURL       = errors.New("voice: error playing audio, invalid URL")
	errVoicePlayMuted            = errors.New("voice: error playing audio, muted")
	errVoicePlayNotConnected     = errors.New("voice: error playing audio, not connected")
	errVoicePlayingAlready       = errors.New("voice: already playing")
	errVoiceSkippedManually      = errors.New("voice: skipped audio manually")
	errVoiceStoppedManually      = errors.New("voice: stopped audio manually")
)

//Voice contains data about the current voice session
type Voice struct {
	sync.Mutex `json:"-"` //This struct gets accessed very repeatedly throughout various goroutines so we need a mutex to prevent race conditions

	//Voice connections and audio sessions
	VoiceConnection  *discordgo.VoiceConnection `json:"voiceConnection"` //The current Discord voice connection
	DiscordSession   *discordgo.Session
	EncodingSession  *dca.EncodeSession    `json:"encodingSession"`  //The encoding session for encoding the audio stream to Opus
	StreamingSession *dca.StreamingSession `json:"streamingSession"` //The streaming session for sending the Opus audio to Discord

	//Voice configurations
	EncodingOptions *dca.EncodeOptions `json:"encodingOptions"` //The settings that will be used for encoding the audio stream to Opus
	RepeatLevel     RepeatLevel        `json:"repeatLevel"`     //0 = No Repeat, 1 = Repeat Playlist, 2 = Repeat Now Playing
	Shuffle         bool               `json:"shuffle"`         //If enabled, entries will be pulled from the queue at random instead of in order
	Muted           bool               `json:"muted"`           //Whether or not audio should be sent to Discord
	Deafened        bool               `json:"deafened"`        //Whether or not audio should be received from Discord

	//Contains data about the current queue
	Entries    []*QueueEntry    `json:"queueEntries"` //Holds a list of queue entries
	NowPlaying *VoiceNowPlaying `json:"nowPlaying"`   //Holds the queue entry currently in the now playing slot

	//Miscellaneous
	TextChannelID string     `json:"textChannelID"` //The channel that was last used to interact with the voice session
	Started       bool       `json:"-"`             //If the playback session has started
	done          chan error //Used to signal when streaming is done or other actions are performed
}

// Connect connects to a given voice channel
func (voice *Voice) Connect(guildID, vChannelID string) error {
	voice.Lock()
	defer voice.Unlock()

	if voice.IsConnected() {
		err := voice.VoiceConnection.ChangeChannel(vChannelID, voice.Muted, voice.Deafened)
		if err != nil {
			return err
		}

		return nil
	}

	fmt.Println("ChannelVoiceJoin")
	voiceConnection, err := voice.DiscordSession.ChannelVoiceJoin(guildID, vChannelID, voice.Muted, voice.Deafened)
	if err != nil {
		return err
	}

	voice.VoiceConnection = voiceConnection
	return nil
}

// Disconnect disconnects from the current voice channel
func (voice *Voice) Disconnect() error {
	voice.Lock()
	defer voice.Unlock()

	//If a voice connection is already established...
	if voice.IsConnected() {
		err := voice.VoiceConnection.Disconnect()
		voice.VoiceConnection = nil
		voice.Started = false
		return err
	}

	//We're not in a voice channel right now
	return nil // eror
}

// Play plays a given queue entry in a connected voice channel
// - queueEntry: The queue entry to play/add to the queue
// - announceQueueAdded: Whether or not to announce a queue added message if something is already playing (used internally for mass playlist additions)
func (voice *Voice) Play(queueEntry *QueueEntry, announceQueueAdded bool) error {
	//Make sure we're conected first
	if !voice.IsConnected() {
		return nil // eror
	}

	//Make sure we're not streaming already
	if voice.IsStreaming() {
		//If we are streaming, add to the queue instead
		voice.QueueAdd(queueEntry)
		if announceQueueAdded {
			// send message play
		}
		return nil
	}

	//Make sure we're allowed to speak
	if voice.Muted {
		return nil
	}

	voice.Lock()

	//Let others know we're beginning to play something
	voice.Started = true
	voice.NowPlaying = &VoiceNowPlaying{Entry: queueEntry}
	updateListeningStatus(voice.DiscordSession, voice.NowPlaying.Entry.Metadata.Artists[0].Name, voice.NowPlaying.Entry.Metadata.Title)

	voice.done = make(chan error)
	voice.Unlock()
	fmt.Println("PlayRaw", voice.NowPlaying.Entry.Metadata.StreamURL)
	msg, err := voice.playRaw(voice.NowPlaying.Entry.Metadata.StreamURL)

	if msg != nil {
		if msg == nil { //errVoiceStoppedManually {
			voice.Started = false
			return nil
		}
	}

	if err != nil {
		voice.Started = false
		switch err {
		case io.ErrUnexpectedEOF:
			if msg != nil { // errVoiceSkippedManually {
				return err
			}
		default:
			return err
		}
	}

	nextQueueEntry := &QueueEntry{}
	index := 0

	switch voice.RepeatLevel {
	case RepeatNone:
		voice.NowPlaying = nil
		if len(voice.Entries) <= 0 {
			voice.Disconnect()
			return nil
		}
		nextQueueEntry, index = voice.QueueGetNext()
		voice.QueueRemove(index)
	case RepeatPlaylist:
		voice.QueueAdd(voice.NowPlaying.Entry)
		nextQueueEntry, index = voice.QueueGetNext()
		voice.QueueRemove(index)
	case RepeatNowPlaying:
		nextQueueEntry = voice.NowPlaying.Entry
	}

	voice.NowPlaying = nil
	return voice.Play(nextQueueEntry, announceQueueAdded)
}

func updateListeningStatus(session *discordgo.Session, name string, title string) {
	err := session.UpdateListeningStatus(name + " - " + title)
	if err != nil {
		return
	}
}

// playRaw plays a given media URL in a connected voice channel
func (voice *Voice) playRaw(mediaURL string) (error, error) {
	if !voice.IsConnected() {
		return nil, nil
	}
	if voice.IsStreaming() {
		return nil, nil
	}

	voice.Lock()

	if voice.Muted {
		return nil, nil
	}

	_, err := url.ParseRequestURI(mediaURL)
	if err != nil {
		return nil, nil
	}

	fmt.Println("dca.EncodeFile")
	voice.EncodingSession, err = dca.EncodeFile(mediaURL, voice.EncodingOptions)
	if err != nil {
		return nil, err
	}
	voice.speaking()
	fmt.Println("dca.NewStream")
	voice.StreamingSession = dca.NewStream(voice.EncodingSession, voice.VoiceConnection, voice.done)
	voice.Unlock()

	//Start a goroutine to update the current streaming position
	go voice.updatePosition()

	//Wait for the streaming session to finish
	msg := <-voice.done

	voice.Lock()

	voice.silent()
	_, err = voice.StreamingSession.Finished()
	voice.StreamingSession = nil

	//Clean up the encoding session
	voice.EncodingSession.Stop()
	voice.EncodingSession.Cleanup()
	voice.EncodingSession = nil

	voice.Unlock()

	//Return any streaming errors, if any
	return msg, err
}

// updatePosition updates the current position of a playing media
func (voice *Voice) updatePosition() {
	for {
		voice.Lock()

		if voice.StreamingSession == nil || voice.NowPlaying == nil {
			voice.Unlock()
			return
		}
		voice.NowPlaying.Position = voice.StreamingSession.PlaybackPosition()

		voice.Unlock()
	}
}

// Stop stops the playback of a media
func (voice *Voice) Stop() error {
	if !voice.IsStreaming() {
		return errVoiceNotStreaming
	}

	voice.done <- errVoiceStoppedManually

	if err := voice.EncodingSession.Stop(); err != nil {
		return err
	}
	voice.EncodingSession.Cleanup()

	return nil
}

// Skip stops the encoding session of a playing media, allowing the play wrapper to continue to the next media in a queue
func (voice *Voice) Skip() error {
	if !voice.IsStreaming() {
		return errVoiceNotStreaming
	}

	//Stop the current now playing
	voice.done <- errVoiceSkippedManually

	//Stop the encoding session
	if err := voice.EncodingSession.Stop(); err != nil {
		return err
	}

	//Clean up the encoding session
	voice.EncodingSession.Cleanup()

	return nil
}

// Pause pauses the playback of a media
func (voice *Voice) Pause() (bool, error) {
	if !voice.IsStreaming() {
		return false, errVoiceNotStreaming
	}

	if isPaused := voice.StreamingSession.Paused(); isPaused {
		return true, errVoicePausedAlready
	}

	voice.StreamingSession.SetPaused(true)
	return true, nil
}

// Resume resumes the playback of a media
func (voice *Voice) Resume() (bool, error) {
	if !voice.IsStreaming() {
		return false, errVoiceNotStreaming
	}

	if isPaused := voice.StreamingSession.Paused(); !isPaused {
		return true, errVoicePlayingAlready
	}

	voice.StreamingSession.SetPaused(false)
	return true, nil
}

// ToggleShuffle toggles the current shuffle setting and manages the queue accordingly
func (voice *Voice) ToggleShuffle() error {
	return nil
}

// speaking allows the sending of audio to Discord
func (voice *Voice) speaking() {
	if voice.IsConnected() {
		voice.VoiceConnection.Speaking(true)
	}
}

// silent prevents the sending of audio to Discord
func (voice *Voice) silent() {
	if voice.IsConnected() {
		voice.VoiceConnection.Speaking(false)
	}
}

// IsConnected returns whether or not a voice connection exists
func (voice *Voice) IsConnected() bool {
	if voice == nil {
		return false
	}

	return voice.VoiceConnection != nil
}

// IsStreaming returns whether a media is playing
func (voice *Voice) IsStreaming() bool {
	voice.Lock()
	defer voice.Unlock()

	//Return false if a voice connection does not exist
	if !voice.IsConnected() {
		return false
	}

	//Return false if the playback session hasn't started
	if !voice.Started {
		return false
	}

	//Return false if a streaming session does not exist
	if voice.StreamingSession == nil {
		return false
	}

	//Return false if an encoding session does not exist
	if voice.EncodingSession == nil {
		return false
	}

	//Otherwise return true
	return true
}

// SetTextChannel sets the text channel to send messages to
func (voice *Voice) SetTextChannel(tChannelID string) {
	//Set voice message output to current text channel
	voice.TextChannelID = tChannelID
}

// VoiceInit initializes a voice object for the given guild
func NewVoice(session *discordgo.Session, config config.VoiceConfig) *Voice {
	return &Voice{
		DiscordSession:  session,
		EncodingOptions: &config.EncodeOptions,
	}
}
