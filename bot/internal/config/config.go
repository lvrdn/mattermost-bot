package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	MMTeam     string   `envconfig:"MM_TEAM"`
	MMToken    string   `envconfig:"MM_TOKEN"`
	MMChannels []string `envconfig:"MM_CHANNEL"`
	MMBotname  string   `envconfig:"MM_BOTNAME"`
}

func GetConfig() (*Config, error) {
	cfg := &Config{}
	err := envconfig.Process("", cfg)

	if err != nil {
		return nil, err
	}

	return cfg, nil
}
