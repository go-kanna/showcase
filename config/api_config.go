package config

//kanna:config
type API struct {
	Port int `default:"8080" validate:"gte=1,lte=65535"`
}
