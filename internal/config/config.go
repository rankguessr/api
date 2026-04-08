package config

import "github.com/rankguessr/api/pkg/renv"

type Config struct {
	AppURL          string `env:"APP_URL,required"`
	WebURL          string `env:"WEB_URL,required"`
	JWTSecret       string `env:"JWT_SECRET,required"`
	OsuClientID     string `env:"OSU_CLIENT_ID,required"`
	OsuClientSecret string `env:"OSU_CLIENT_SECRET,required"`
	DatabaseURL     string `env:"DATABASE_URL,required"`
}

func Read() (*Config, error) {
	cfg := &Config{}
	return cfg, renv.Parse(cfg)
}
