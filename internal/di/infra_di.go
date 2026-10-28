package di

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kanna/showcase/config"
	"github.com/go-kanna/showcase/internal/infra/storage/db"
	"github.com/go-kanna/showcase/internal/infra/storage/tx"
)

const (
	dbConnectTimeout           = 5 * time.Second
	dbConnLifetimeJitter       = 10 * time.Minute
	dbStatementTimeout         = 30 * time.Second
	dbIdleInTransactionTimeout = time.Minute
	dbStartupWait              = time.Minute
)

//kanna:container
type Infra struct {
	_              context.Context `di:"arg"` //nolint:containedctx // required by kanna-di
	cfg            Config
	Writer         db.Writer         `di:"with=provideWriter"`
	Reader         db.Reader         `di:"with=provideReader"`
	Transactor     tx.Transactor     `di:"with=provideTransactor"`
	ReadTransactor tx.ReadTransactor `di:"with=provideReadTransactor"`
}

func provideWriter(ctx context.Context, cfg config.Database) (db.Writer, func(), error) {
	writer, err := db.NewWriter(ctx, cfg.WriterURL, poolOptions(cfg)...)
	if err != nil {
		return db.Writer{}, nil, fmt.Errorf("new writer: %w", err)
	}
	return writer, writer.Close, nil
}

func provideReader(ctx context.Context, cfg config.Database, writer db.Writer) (db.Reader, func(), error) {
	reader, err := db.NewReader(ctx, cfg.ReaderURL, poolOptions(cfg)...)
	if err != nil {
		writer.Close()
		return db.Reader{}, nil, fmt.Errorf("new reader: %w", err)
	}
	return reader, reader.Close, nil
}

func poolOptions(cfg config.Database) []db.Option {
	return []db.Option{
		db.WithMaxConns(cfg.PoolMaxConns),
		db.WithConnectTimeout(dbConnectTimeout),
		db.WithMaxConnLifetimeJitter(dbConnLifetimeJitter),
		db.WithSessionTimeouts(dbStatementTimeout, dbIdleInTransactionTimeout),
		db.WithStartupWait(dbStartupWait),
	}
}

func provideTransactor(writer db.Writer) tx.Transactor {
	return tx.NewTransactor(writer)
}

func provideReadTransactor(reader db.Reader) tx.ReadTransactor {
	return tx.NewReadTransactor(reader)
}
