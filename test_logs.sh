#!/bin/bash

# Скрипт для тестирования логирования запросов

echo "Отправка тестового запроса на /test endpoint..."
curl -X POST http://localhost:8080/test \
  -H "Content-Type: application/json" \
  -H "X-Custom-Header: test-value" \
  -H "User-Agent: TestScript/1.0" \
  -d '{
    "test": "data",
    "timestamp": "'$(date -Iseconds)'",
    "message": "Тестовый запрос для проверки логирования"
  }'

echo -e "\n\nТестовый запрос отправлен!"
echo "Откройте http://localhost:8080/logs для просмотра результата"
