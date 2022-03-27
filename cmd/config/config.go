package config

import (
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
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
	Token  string `json:"token"`
	Bot    string `json:"bot"`
	Id     int64  `json:"id"`
	Prefix string `json:"prefix"`
}

type SheetsConfig struct {
	Id   string `json:"id"`
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
	return &config, nil
}
