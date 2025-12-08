-- Add request_headers column to request_logs table (MySQL version)
ALTER TABLE request_logs ADD COLUMN request_headers TEXT;
