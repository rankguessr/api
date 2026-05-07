package config

import (
	"net/url"

	"github.com/rankguessr/api/pkg/renv"
)

type Config struct {
	RedisURL        string `env:"REDIS_URL,required"`
	PORT            string `env:"PORT,required"`
	AppURL          string `env:"APP_URL,required"`
	WebURL          string `env:"WEB_URL,required"`
	TurnstileSecret string `env:"TURNSTILE_SECRET,required"`
	OsuClientID     string `env:"OSU_CLIENT_ID,required"`
	OsuClientSecret string `env:"OSU_CLIENT_SECRET,required"`
	EncryptionKey   string `env:"ENCRYPTION_KEY,required"`
	DatabaseURL     string `env:"DATABASE_URL,required"`

	S3Endpoint   string `env:"S3_ENDPOINT,required"`
	S3Region     string `env:"S3_REGION,required"`
	S3BucketName string `env:"S3_BUCKET_NAME,required"`
	S3PublicURL  string `env:"S3_PUBLIC_URL,required"`
	S3SecretKey  string `env:"S3_SECRET_KEY,required"`
	S3AccessKey  string `env:"S3_ACCESS_KEY,required"`
}

func (c *Config) WebDomain() string {
	parsedURL, _ := url.Parse(c.WebURL)
	return parsedURL.Hostname()
}

func Read() (*Config, error) {
	cfg := &Config{}
	return cfg, renv.Parse(cfg)
}
