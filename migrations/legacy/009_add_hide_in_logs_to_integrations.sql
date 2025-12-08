-- +goose Up
ALTER TABLE integrations ADD COLUMN hide_in_logs BOOLEAN DEFAULT FALSE;

-- +goose Down
ALTER TABLE integrations DROP COLUMN hide_in_logs;
