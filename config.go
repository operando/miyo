package main

import "github.com/BurntSushi/toml"

type Config struct {
	Crowi      Crowi  `toml:"crowi"`
	SlackToken string `toml:"slack_token"`
}

type Crowi struct {
	ApiUrl string `toml:"api_url"`
	Token  string `toml:"token"`
}

func LoadConfig(configPath string, config *Config) (*Config, error) {
	_, err := toml.DecodeFile(configPath, config)
	if err != nil {
		return config, err
	}
	return config, nil
}
