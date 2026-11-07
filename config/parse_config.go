package config

import "github.com/caarlos0/env/v11"

type validatable interface {
	Validate() error
}

func parse[T validatable]() T {
	var cfg T
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}
	if err := cfg.Validate(); err != nil {
		panic(err)
	}
	return cfg
}
