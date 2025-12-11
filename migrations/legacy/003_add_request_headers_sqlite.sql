-- Add request_headers column to request_logs table (SQLite version)
ALTER TABLE request_logs ADD COLUMN request_headers TEXT;
