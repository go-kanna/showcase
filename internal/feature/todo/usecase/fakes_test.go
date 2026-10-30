package usecase_test

import (
	"context"

	"github.com/go-kanna/showcase/internal/feature/todo/model"
	"github.com/go-kanna/showcase/internal/feature/todo/repository"
	"github.com/go-kanna/showcase/internal/infra/storage/tx"
)

// fakeTransactor runs fn at once, in no transaction, for a test with no
// database.
type fakeTransactor struct{}

func (fakeTransactor) WithTx(_ context.Context, fn func(tx.Tx) error) error { return fn(tx.Tx{}) }

// fakeTodos records what is upserted. Any other method panics, through the nil
// repository.Todo it embeds, as the use case under test is not to call it.
type fakeTodos struct {
	repository.Todo

	upserted []model.Todo
}

func (f *fakeTodos) Bind(tx.Tx) repository.Todo { return f }

func (f *fakeTodos) Upsert(_ context.Context, m model.Todo) error {
	f.upserted = append(f.upserted, m)
	return nil
}
