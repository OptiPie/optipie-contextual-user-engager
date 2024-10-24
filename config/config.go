package config

import "github.com/caarlos0/env/v11"

type Config struct {
	Twitter struct {
		OAuthToken       string `env:"TWITTER_OAUTH_TOKEN"`
		OAuthTokenSecret string `env:"TWITTER_OAUTH_TOKEN_SECRET"`
	}
	Openai struct {
		SecretKey string `env:"OPENAI_SECRET_KEY"`
	}
}

func GetConfig() (*Config, error) {
	cfg := new(Config)
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
