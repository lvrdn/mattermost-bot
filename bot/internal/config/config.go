package config

import (
	"net/url"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	MMServer   *urlDecoder `envconfig:"MM_SERVER"`
	MMTeam     string      `envconfig:"MM_TEAM"`
	MMToken    string      `envconfig:"MM_TOKEN"`
	MMChannels []string    `envconfig:"MM_CHANNEL"`
	MMBotname  string      `envconfig:"MM_BOTNAME"`
}

func GetConfig() (*Config, error) {
	cfg := &Config{}
	err := envconfig.Process("", cfg)

	if err != nil {
		return nil, err
	}

	return cfg, nil
}

type urlDecoder url.URL

func (ud *urlDecoder) Decode(value string) error {
	url, err := url.Parse(value)
	if err != nil {
		return err
	}

	*ud = urlDecoder(*url)

	return nil
}

func (ud *urlDecoder) String() string {

	url := url.URL(*ud)
	return url.String()
}
