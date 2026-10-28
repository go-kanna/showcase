package config

import "github.com/caarlos0/env/v11"

type validatable interface {
	Validate() error
}

func parse[T any]() T {
	var cfg T
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}
	if v, ok := any(cfg).(validatable); ok {
		if err := v.Validate(); err != nil {
			panic(err)
		}
	}
	return cfg
}
