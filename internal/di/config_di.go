package di

import "github.com/go-kanna/showcase/config"

//kanna:container returns=Config
type Config struct {
	API      config.API
	Database config.Database
}
