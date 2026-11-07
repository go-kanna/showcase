package di

import (
	"github.com/go-kanna/config/source"

	"github.com/go-kanna/showcase/config"
)

//kanna:container returns=Config
type Config struct {
	_ source.Source `di:"arg=src"`

	API      config.API
	Database config.Database
}
