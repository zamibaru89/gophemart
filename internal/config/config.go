package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

type ServerConfig struct {
	Address        string `env:"ADDRESS"`
	DSN            string `env:"DATABASE_URI"`
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func LoadServerConfig() (conf ServerConfig, err error) {
	flag.StringVar(&conf.Address, "a", ":80", "")
	flag.StringVar(&conf.AccrualAddress, "r", ":8080", "")
	flag.StringVar(&conf.DSN, "d", "", "")
	flag.Parse()

	err = env.Parse(&conf)
	if err != nil {
		return conf, err
	}
	return
}
