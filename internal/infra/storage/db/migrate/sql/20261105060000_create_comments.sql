-- +goose Up
CREATE TABLE comments
(
    id         UUID PRIMARY KEY DEFAULT UUIDV7(),
    todo_id    UUID NOT NULL REFERENCES todos (id) ON DELETE CASCADE,
    body       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

-- The latest comments of each to-do, which ListTodo reads.
CREATE INDEX comments_todo_id_created_at_idx ON comments (todo_id, created_at DESC, id DESC);

-- +goose Down
DROP TABLE comments;
