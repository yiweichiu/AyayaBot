package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

// RedeemConfig holds configuration for the redeem functionality.
type RedeemConfig struct {
	Service bool `yaml:"service"`
	API     struct {
		URL    string `yaml:"url"`
		APIKey string `yaml:"api_key"`
	} `yaml:"api"`
	Schedule []string `yaml:"schedule"` // e.g., ["0 8 * * *", "0 18 * * *"]
}

// NewsConfig holds configuration for the news fetching functionality.
type NewsConfig struct {
	Service bool `yaml:"service"`
	API     struct { // New nested struct
		URL string `yaml:"url"`
	} `yaml:"api"`
	Schedule []string `yaml:"schedule"` // e.g., ["0 8 * * *", "0 12 * * *"]
}

// Config is the main configuration structure for the application.
type Config struct {
	Discord struct {
		Token     string `yaml:"token"`
		ChannelID string `yaml:"channel_id"`
	} `yaml:"discord"`
	Redeem RedeemConfig `yaml:"redeem"`
	News   NewsConfig   `yaml:"news"`
}

func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Set default values for backward compatibility
	config := Config{}
	config.Redeem.Service = true
	config.News.Service = true

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
