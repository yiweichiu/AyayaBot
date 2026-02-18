package config

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Discord struct {
		Token string `yaml:"token"`
		ChannelID string `yaml:"channel_id"` // Assuming a channel ID will be needed for sending messages
	} `yaml:"discord"`
	API struct {
		URL    string `yaml:"url"`
		APIKey string `yaml:"api_key"`
	} `yaml:"api"`
	Schedule []string `yaml:"schedule"` // e.g., ["0 8 * * *", "0 18 * * *"]
}

func LoadConfig(configPath string) (*Config, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

