-- +goose Up
CREATE TABLE todos
(
    id           UUID PRIMARY KEY DEFAULT UUIDV7(),
    title        TEXT NOT NULL,
    completed_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL,
    updated_at   TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE todos;
