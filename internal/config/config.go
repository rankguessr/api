package config

import (
	"net/url"

	"github.com/rankguessr/api/pkg/renv"
)

type Config struct {
	TurnstileSecret string `env:"TURNSTILE_SECRET,required"`
	RedisURL        string `env:"REDIS_URL,required"`
	SentryDSN       string `env:"SENTRY_DSN,required"`
	PORT            string `env:"PORT,required"`
	AppURL          string `env:"APP_URL,required"`
	WebURL          string `env:"WEB_URL,required"`
	OsuClientID     string `env:"OSU_CLIENT_ID,required"`
	OsuClientSecret string `env:"OSU_CLIENT_SECRET,required"`
	EncryptionKey   string `env:"ENCRYPTION_KEY,required"`
	DatabaseURL     string `env:"DATABASE_URL,required"`
}

func (c *Config) WebDomain() string {
	parsedURL, _ := url.Parse(c.WebURL)
	return parsedURL.Hostname()
}

func Read() (*Config, error) {
	cfg := &Config{}
	return cfg, renv.Parse(cfg)
}
