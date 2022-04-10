package config

import (
	"encoding/json"
	"github.com/jonas747/dca"
	"io/ioutil"

	"github.com/pkg/errors"
)

const FilePath = "secret_config.json"

type Config struct {
	Discord DiscordConfig `json:"discord"`
	Sheets  SheetsConfig  `json:"sheets"`
	VK      VKConfig      `json:"vk"`
	Lichess LichessConfig `json:"lichess"`
	General GeneralConfig `json:"general"`
}

type DiscordConfig struct {
	Token  string      `json:"token"`
	Bot    string      `json:"bot"`
	ID     int64       `json:"id"`
	Prefix string      `json:"prefix"`
	Voice  VoiceConfig `json:"voice"`
}

type VoiceConfig struct {
	dca.EncodeOptions
}

type SheetsConfig struct {
	ID   string `json:"id"`
	Film string `json:"film"`
}

type VKConfig struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LichessConfig struct {
	Token string `json:"token"`
}

type GeneralConfig struct {
	Debug bool `json:"debug"`
}

func InitConfig() (*Config, error) {
	var config Config
	jsonFile, err := ioutil.ReadFile(FilePath)
	if err != nil {
		return nil, errors.Wrap(err, "read failed")
	}
	err = json.Unmarshal(jsonFile, &config)
	if err != nil {
		return nil, errors.Wrap(err, "Unmarshal failed")
	}
	//config.Discord.Voice = VoiceConfig{
	//	EncodeOptions: dca.EncodeOptions{
	//		Volume:        256,
	//		Channels:      1,
	//		FrameRate:     48000,
	//		FrameDuration: 40,
	//		Bitrate:       64,
	//		PacketLoss:    5,
	//		RawOutput:     false,
	//		Application:   "ffmpeg",
	//		//CoverFormat:      "",
	//		//CompressionLevel: 5,
	//		//BufferedFrames: 1000,
	//		VBR: true,
	//		//Threads:        0,
	//		//StartTime:      0,
	//		//AudioFilter:    "",
	//		//Comment:        "",
	//	},
	//}
	config.Discord.Voice = VoiceConfig{
		EncodeOptions: *dca.StdEncodeOptions,
	}
	return &config, nil
}
