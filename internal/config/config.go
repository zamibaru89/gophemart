package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

type ServerConfig struct {
	Address string `env:"ADDRESS"`
	DSN     string `env:"DATABASE_URI"`
}

func LoadServerConfig() (conf ServerConfig, err error) {
	flag.StringVar(&conf.Address, "a", ":8080", "")
	flag.StringVar(&conf.DSN, "d", "", "")
	flag.Parse()

	err = env.Parse(&conf)
	if err != nil {
		return conf, err
	}
	return
}
