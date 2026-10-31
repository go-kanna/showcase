-- query: GetForUpdate
SELECT id, title, priority, completed_at, created_at, updated_at
FROM todos
WHERE id = :id
FOR UPDATE;

-- query: Complete
UPDATE todos
SET
    completed_at = :completedAt,
    updated_at = :completedAt
WHERE
    id = :id
    AND completed_at IS NULL;

-- query: Uncomplete
UPDATE todos
SET
    completed_at = NULL,
    updated_at = :updatedAt
WHERE
    id = :id
    AND completed_at IS NOT NULL;
