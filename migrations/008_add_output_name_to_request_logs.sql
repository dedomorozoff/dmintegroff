-- +goose Up
ALTER TABLE request_logs ADD COLUMN output_name VARCHAR(255) DEFAULT '';

-- +goose Down
ALTER TABLE request_logs DROP COLUMN output_name;
