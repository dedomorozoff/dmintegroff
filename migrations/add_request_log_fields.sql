-- Миграция для добавления новых полей в таблицу request_logs
-- Дата: 2024-12-12
-- Описание: Добавляем поля для детальной истории запросов

-- Добавляем новые поля в таблицу request_logs
ALTER TABLE request_logs ADD COLUMN response_headers TEXT DEFAULT '';
ALTER TABLE request_logs ADD COLUMN response_time INTEGER DEFAULT 0;
ALTER TABLE request_logs ADD COLUMN request_size INTEGER DEFAULT 0;
ALTER TABLE request_logs ADD COLUMN response_size INTEGER DEFAULT 0;

-- Создаем индексы для быстрого поиска
CREATE INDEX IF NOT EXISTS idx_request_logs_integration_created ON request_logs(integration_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_request_logs_status_created ON request_logs(status_code, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_request_logs_method_created ON request_logs(method, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_request_logs_log_type_created ON request_logs(log_type, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_request_logs_output_name ON request_logs(output_name);

-- Обновляем существующие записи, устанавливая размеры на основе существующих данных
UPDATE request_logs 
SET 
    request_size = LENGTH(COALESCE(request_body, '')),
    response_size = LENGTH(COALESCE(response_body, ''))
WHERE request_size = 0 OR response_size = 0;