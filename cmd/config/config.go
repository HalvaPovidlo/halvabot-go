package config

import (
	"encoding/json"
	"github.com/HalvaPovidlo/discordBotGo/internal/music/search"
	"io/ioutil"

	"github.com/khodand/dca"
	"github.com/pkg/errors"

	"github.com/HalvaPovidlo/discordBotGo/internal/music/api/discord"
)

const FilePath = "secret_config.json"

type Config struct {
	General GeneralConfig        `json:"general"`
	Host    HostConfig           `json:"host"`
	Discord DiscordConfig        `json:"discord"`
	Youtube search.YouTubeConfig `json:"youtube"`
	// Sheets  SheetsConfig  `json:"sheets"`
	// VK      VKConfig      `json:"vk"`
	// Lichess LichessConfig `json:"lichess"`
}

type DiscordConfig struct {
	Token  string            `json:"token"`
	Bot    string            `json:"bot"`
	ID     int64             `json:"id"`
	Prefix string            `json:"prefix"`
	API    discord.APIConfig `json:"api"`
	Voice  VoiceConfig       `json:"voice"`
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

type HostConfig struct {
	IP   string `json:"ip"`
	Bot  string `json:"bot"`
	Mock string `json:"mock"`
	Web  string `json:"web"`
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

	config.Discord.Voice = VoiceConfig{
		EncodeOptions: *dca.StdEncodeOptions,
	}
	return &config, nil
}
