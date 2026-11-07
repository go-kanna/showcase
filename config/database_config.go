package config

//kanna:config
type Database struct {
	AdminURL     string
	WriterURL    string `validate:"required"`
	ReaderURL    string `validate:"required"`
	PoolMaxConns int32  `default:"10"        validate:"gte=1"`
}
