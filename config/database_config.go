package config

//go:generate go tool kanna-validate
//kanna:validate
type Database struct {
	AdminURL     string `env:"DATABASE_URL"`
	WriterURL    string `env:"DATABASE_WRITER_URL"     validate:"required"`
	ReaderURL    string `env:"DATABASE_READER_URL"     validate:"required"`
	PoolMaxConns int32  `env:"DATABASE_POOL_MAX_CONNS" envDefault:"10"     validate:"min=1"`
}

func ParseDatabase() Database {
	return parse[Database]()
}
