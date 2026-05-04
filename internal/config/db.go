package config

import "github.com/rankguessr/api/pkg/renv"

type OsuAuth struct {
	OsuClientID     string `env:"OSU_CLIENT_ID,required"`
	OsuClientSecret string `env:"OSU_CLIENT_SECRET,required"`
}

func ReadOsuAuth() (*OsuAuth, error) {
	cfg := &OsuAuth{}
	return cfg, renv.Parse(cfg)
}
