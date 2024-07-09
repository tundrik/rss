package config


import (
	"github.com/ilyakaznacheev/cleanenv"
)


type Config struct {
	PgString	string `env:"DATABASE_URL" env-default:"postgres://postgres:postgres@postgres:5432/postgres"`
}

func GetConfig() (Config, error) {
	var cfg Config
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}