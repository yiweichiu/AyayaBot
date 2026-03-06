package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

// ChannelConfig holds information for a single Discord channel.
type ChannelConfig struct {
	Name string `yaml:"name"`
	ID   string `yaml:"id"`
}

// RedeemConfig holds configuration for the redeem functionality.
type RedeemConfig struct {
	Service   bool   `yaml:"service"`
	Channel   string `yaml:"channel"`
	HideEmbed bool   `yaml:"hide_embed"`
	API       struct {
		URL    string `yaml:"url"`
		APIKey string `yaml:"api_key"`
	} `yaml:"api"`
	Schedule []string `yaml:"schedule"` // e.g., ["0 8 * * *", "0 18 * * *"]
}

// NewsConfig holds configuration for the news fetching functionality.
type NewsConfig struct {
	Service     bool   `yaml:"service"`
	Channel     string `yaml:"channel"`
	SendContent bool   `yaml:"send_content"`
	HideEmbed   bool   `yaml:"hide_embed"`
	API         struct { // New nested struct
		URL string `yaml:"url"`
	} `yaml:"api"`
	Schedule []string `yaml:"schedule"` // e.g., ["0 8 * * *", "0 12 * * *"]
}

// Config is the main configuration structure for the application.
type Config struct {
	Discord struct {
		Token    string          `yaml:"token"`
		Channels []ChannelConfig `yaml:"channels"`
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

	// Create a map of channel names to IDs for easy lookup
	channelMap := make(map[string]string)
	for _, ch := range config.Discord.Channels {
		channelMap[ch.Name] = ch.ID
	}

	// Validation: If a service's channel is not found, disable the service
	if _, exists := channelMap[config.Redeem.Channel]; !exists {
		config.Redeem.Service = false
	}
	if _, exists := channelMap[config.News.Channel]; !exists {
		config.News.Service = false
	}

	return &config, nil
}
