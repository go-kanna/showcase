package config

//kanna:validate
type API struct {
	Port int `env:"PORT" envDefault:"8080" validate:"gte=1,lte=65535"`
}

func ParseAPI() API {
	return parse[API]()
}
