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
