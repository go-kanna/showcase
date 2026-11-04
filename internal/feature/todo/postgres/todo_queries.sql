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

-- query: ListWithCommentCounts
SELECT
    t.id,
    t.title,
    t.priority,
    t.completed_at,
    t.created_at,
    t.updated_at,
    COUNT(c.id) AS comment_count
FROM todos AS t
LEFT JOIN comments AS c ON t.id = c.todo_id
GROUP BY t.id;
