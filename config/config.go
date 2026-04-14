package config

import "github.com/caarlos0/env/v11"

type Config struct {
	Twitter struct {
		OAuthToken       string `env:"TWITTER_OAUTH_TOKEN"`
		OAuthTokenSecret string `env:"TWITTER_OAUTH_TOKEN_SECRET"`
		User             string `env:"TWITTER_USER"`
		Password         string `env:"TWITTER_PASSWORD"`
	}
	Openai struct {
		SecretKey string `env:"OPENAI_SECRET_KEY"`
	}
	Browser struct {
		UserDataDir string `env:"BROWSER_USER_DATA_DIR"`
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
