package config

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)


type Config struct {
	PgString    string        `env:"DATABASE_URL" env-default:"postgres://postgres:postgres@postgres:5432/postgres"`
	KeeperDelay time.Duration `env:"KEEPER_DELAY" env-default:"60s"`
	ConnLimit   int           `env:"CONN_LIMIT" env-default:"256"`
	ReqTimeout  time.Duration `env:"REQ_TIMEOUT" env-default:"10s"`
	CumLimit    int           `env:"CUM_LIMIT" env-default:"300"`
	CumDeadline time.Duration `env:"CUM_DEADLINE" env-default:"200ms"`
}

func GetConfig() (Config, error) {
	var cfg Config
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}
