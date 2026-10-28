-- name: GetTodo :one
SELECT
    t.id,
    t.title,
    t.completed_at,
    t.created_at,
    t.updated_at
FROM todos AS t
WHERE
    t.id = $1;

-- name: ListTodo :many
SELECT
    t.id,
    t.title,
    t.completed_at,
    t.created_at,
    t.updated_at
FROM todos AS t;

-- name: UpsertTodo :execrows
INSERT INTO todos (id, title, completed_at, created_at, updated_at)
SELECT
    sqlc.arg(id)::UUID,
    sqlc.arg(title)::TEXT,
    sqlc.narg(completed_at)::TIMESTAMPTZ,
    sqlc.arg(created_at)::TIMESTAMPTZ,
    sqlc.arg(updated_at)::TIMESTAMPTZ
ON CONFLICT (id) DO UPDATE
    SET
        title = EXCLUDED.title,
        completed_at = EXCLUDED.completed_at,
        updated_at = EXCLUDED.updated_at;

-- name: CompleteTodo :execrows
UPDATE todos
SET
    completed_at = sqlc.arg(completed_at)::TIMESTAMPTZ
WHERE
    id = sqlc.arg(id)
    AND completed_at IS NULL;

-- name: UncompleteTodo :execrows
UPDATE todos
SET
    completed_at = NULL
WHERE
    id = sqlc.arg(id)
    AND completed_at IS NOT NULL;

-- name: DeleteTodo :exec
DELETE
FROM todos
WHERE
    id = $1;