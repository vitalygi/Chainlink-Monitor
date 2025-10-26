package config

import (
	"os"

	"github.com/goccy/go-yaml"
)

type Config struct {
	NodeUrl string
	Feeds []FeedConfig `yaml:"feeds"`
}

type FeedConfig struct {
	CurrencyPair string `yaml:"currency_pair"`
	HexAddress   string `yaml:"hex_address"`
	Decimals     int32  `yaml:"decimals"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	cfg.NodeUrl = os.Getenv("node_url")
	return &cfg, nil
}