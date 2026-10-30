-- query: Complete
UPDATE todos
SET
    completed_at = :completedAt
WHERE
    id = :id
    AND completed_at IS NULL;

-- query: Uncomplete
UPDATE todos
SET
    completed_at = NULL
WHERE
    id = :id
    AND completed_at IS NOT NULL;
