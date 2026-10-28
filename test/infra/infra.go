package infra

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/peterldowns/pgtestdb"
	"github.com/stretchr/testify/require"

	"github.com/go-kanna/showcase/config"
	"github.com/go-kanna/showcase/internal/di"
	"github.com/go-kanna/showcase/internal/infra/storage/db"
	"github.com/go-kanna/showcase/internal/infra/storage/db/migrate"
	"github.com/go-kanna/showcase/internal/infra/storage/tx"
)

const (
	testPoolMaxConns = 2
)

type Infra struct {
	*di.Infra
}

func New(t *testing.T) *Infra {
	t.Helper()

	requireDatabaseEnv(t)
	dbCfg := config.ParseDatabase()

	cfg := pgtestdb.Custom(t, configFromURL(t, dbCfg.AdminURL), migrator{})

	writer, err := db.NewWriter(t.Context(), testPoolURL(t, dbCfg.WriterURL, cfg.Database))
	require.NoError(t, err)
	t.Cleanup(writer.Close)

	reader, err := db.NewReader(t.Context(), testPoolURL(t, dbCfg.ReaderURL, cfg.Database))
	require.NoError(t, err)
	t.Cleanup(reader.Close)

	return &Infra{
		Infra: &di.Infra{
			Writer:         writer,
			Reader:         reader,
			Transactor:     tx.NewTransactor(writer),
			ReadTransactor: tx.NewReadTransactor(reader),
		},
	}
}

func (i *Infra) InTx(t *testing.T, fn func(tx tx.Tx)) {
	t.Helper()

	require.NoError(t, i.Transactor.WithTx(t.Context(), func(tx tx.Tx) error {
		fn(tx)
		return nil
	}))
}

func requireDatabaseEnv(t *testing.T) {
	t.Helper()

	var missing []string
	for _, key := range []string{"DATABASE_URL", "DATABASE_WRITER_URL", "DATABASE_READER_URL"} {
		if os.Getenv(key) == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		t.Fatalf(
			"missing env vars %v: create .env via `envsubst < .env.example > .env` and run tests through make",
			missing,
		)
	}
}

type migrator struct{}

var _ pgtestdb.Migrator = migrator{}

func (migrator) Hash() (string, error) {
	hash, err := migrate.Hash()
	if err != nil {
		return "", fmt.Errorf("hash migrations: %w", err)
	}
	return hash, nil
}

func (migrator) Migrate(ctx context.Context, sqlDB *sql.DB, _ pgtestdb.Config) error {
	if err := migrate.UpDB(ctx, sqlDB); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}

func configFromURL(t *testing.T, raw string) pgtestdb.Config {
	t.Helper()

	u, err := url.Parse(raw)
	require.NoError(t, err)
	password, _ := u.User.Password()
	return pgtestdb.Config{
		DriverName: "pgx",
		Host:       u.Hostname(),
		Port:       u.Port(),
		User:       u.User.Username(),
		Password:   password,
		Database:   strings.TrimPrefix(u.Path, "/"),
		Options:    u.RawQuery,
	}
}

func testPoolURL(t *testing.T, raw, name string) string {
	t.Helper()

	u, err := url.Parse(raw)
	require.NoError(t, err)
	u.Path = "/" + name
	q := u.Query()
	q.Set("pool_max_conns", strconv.Itoa(testPoolMaxConns))
	u.RawQuery = q.Encode()
	return u.String()
}
