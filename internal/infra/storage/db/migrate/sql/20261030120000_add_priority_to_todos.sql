-- +goose Up
-- The default fills the rows there are; the application gives every new one a
-- priority of its own.
ALTER TABLE todos ADD COLUMN priority TEXT NOT NULL DEFAULT 'normal';
ALTER TABLE todos ALTER COLUMN priority DROP DEFAULT;

-- +goose Down
ALTER TABLE todos DROP COLUMN priority;
