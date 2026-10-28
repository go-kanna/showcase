package db

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const pingInterval = time.Second

type settings struct {
	pool        *pgxpool.Config
	startupWait time.Duration
}

type Option func(s *settings)

// WithMaxConns overrides the URL's pool_max_conns, or pgx's default of the
// larger of 4 and the CPU count.
func WithMaxConns(n int32) Option {
	return func(s *settings) {
		s.pool.MaxConns = n
	}
}

func WithConnectTimeout(d time.Duration) Option {
	return func(s *settings) {
		s.pool.ConnConfig.ConnectTimeout = d
	}
}

// WithMaxConnLifetimeJitter spreads out when connections opened together
// reach their lifetime.
func WithMaxConnLifetimeJitter(d time.Duration) Option {
	return func(s *settings) {
		s.pool.MaxConnLifetimeJitter = d
	}
}

// WithApplicationName names the connections in pg_stat_activity.
func WithApplicationName(name string) Option {
	return func(s *settings) {
		s.pool.ConnConfig.RuntimeParams["application_name"] = name
	}
}

// WithSessionTimeouts ends a statement that runs longer than statement, and a
// session that sits in a transaction doing nothing for idleInTransaction.
func WithSessionTimeouts(statement, idleInTransaction time.Duration) Option {
	return func(s *settings) {
		s.pool.ConnConfig.RuntimeParams["statement_timeout"] = strconv.FormatInt(statement.Milliseconds(), 10)
		s.pool.ConnConfig.RuntimeParams["idle_in_transaction_session_timeout"] =
			strconv.FormatInt(idleInTransaction.Milliseconds(), 10)
	}
}

// WithStartupWait keeps trying the first connection for up to d, so a process
// started alongside its database, as when the host reboots, waits for it.
func WithStartupWait(d time.Duration) Option {
	return func(s *settings) {
		s.startupWait = d
	}
}

func New(ctx context.Context, url string, opts ...Option) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("parse pool config: %w", err)
	}
	s := settings{pool: cfg}
	for _, opt := range opts {
		opt(&s)
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := ping(ctx, pool.Ping, s.startupWait, pingInterval); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}

func ping(ctx context.Context, f func(context.Context) error, wait, interval time.Duration) error {
	deadline := time.Now().Add(wait)
	for {
		err := f(ctx)
		if err == nil || !time.Now().Add(interval).Before(deadline) {
			return err
		}
		select {
		case <-ctx.Done():
			return err
		case <-time.After(interval):
		}
	}
}

type Writer struct {
	*pgxpool.Pool
}

type Reader struct {
	*pgxpool.Pool
}

func NewWriter(ctx context.Context, url string, opts ...Option) (Writer, error) {
	pool, err := New(ctx, url, opts...)
	if err != nil {
		return Writer{}, err
	}
	return Writer{Pool: pool}, nil
}

func NewReader(ctx context.Context, url string, opts ...Option) (Reader, error) {
	pool, err := New(ctx, url, opts...)
	if err != nil {
		return Reader{}, err
	}
	return Reader{Pool: pool}, nil
}
