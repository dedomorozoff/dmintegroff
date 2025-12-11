#!/bin/bash

# Скрипт для тестирования функционала тестовых вебхуков

# Цвета для вывода
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

BASE_URL="${BASE_URL:-http://localhost:8080}"

echo -e "${YELLOW}=== Тестирование функционала тестовых вебхуков ===${NC}\n"

# Проверка, что сервер запущен
echo -e "${YELLOW}Проверка доступности сервера...${NC}"
if ! curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/login" | grep -q "200"; then
    echo -e "${RED}Ошибка: Сервер недоступен на $BASE_URL${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Сервер доступен${NC}\n"

# Получение токена (требуется авторизация)
echo -e "${YELLOW}Для тестирования необходимо создать тестовый вебхук через веб-интерфейс${NC}"
echo -e "${YELLOW}1. Откройте $BASE_URL/dashboard${NC}"
echo -e "${YELLOW}2. Нажмите 'Протестировать вебхук'${NC}"
echo -e "${YELLOW}3. Скопируйте токен из URL${NC}\n"

read -p "Введите токен тестового вебхука: " TOKEN

if [ -z "$TOKEN" ]; then
    echo -e "${RED}Ошибка: Токен не может быть пустым${NC}"
    exit 1
fi

WEBHOOK_URL="$BASE_URL/webhook/test/$TOKEN"

echo -e "\n${YELLOW}=== Отправка тестовых запросов ===${NC}\n"

# Тест 1: POST запрос с JSON
echo -e "${YELLOW}Тест 1: POST запрос с JSON${NC}"
RESPONSE=$(curl -s -X POST "$WEBHOOK_URL" \
  -H "Content-Type: application/json" \
  -H "X-Test-Header: Test-Value" \
  -d '{"test": true, "message": "Hello from test script", "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"}')

if echo "$RESPONSE" | grep -q "success"; then
    echo -e "${GREEN}✓ POST запрос успешно отправлен${NC}"
else
    echo -e "${RED}✗ Ошибка при отправке POST запроса${NC}"
    echo "$RESPONSE"
fi

sleep 1

# Тест 2: GET запрос с параметрами
echo -e "\n${YELLOW}Тест 2: GET запрос с параметрами${NC}"
RESPONSE=$(curl -s "$WEBHOOK_URL?param1=value1&param2=value2&test=true")

if echo "$RESPONSE" | grep -q "success"; then
    echo -e "${GREEN}✓ GET запрос успешно отправлен${NC}"
else
    echo -e "${RED}✗ Ошибка при отправке GET запроса${NC}"
    echo "$RESPONSE"
fi

sleep 1

# Тест 3: PUT запрос
echo -e "\n${YELLOW}Тест 3: PUT запрос${NC}"
RESPONSE=$(curl -s -X PUT "$WEBHOOK_URL" \
  -H "Content-Type: text/plain" \
  -d "This is a test PUT request")

if echo "$RESPONSE" | grep -q "success"; then
    echo -e "${GREEN}✓ PUT запрос успешно отправлен${NC}"
else
    echo -e "${RED}✗ Ошибка при отправке PUT запроса${NC}"
    echo "$RESPONSE"
fi

sleep 1

# Тест 4: DELETE запрос
echo -e "\n${YELLOW}Тест 4: DELETE запрос${NC}"
RESPONSE=$(curl -s -X DELETE "$WEBHOOK_URL")

if echo "$RESPONSE" | grep -q "success"; then
    echo -e "${GREEN}✓ DELETE запрос успешно отправлен${NC}"
else
    echo -e "${RED}✗ Ошибка при отправке DELETE запроса${NC}"
    echo "$RESPONSE"
fi

sleep 1

# Тест 5: POST с XML
echo -e "\n${YELLOW}Тест 5: POST запрос с XML${NC}"
RESPONSE=$(curl -s -X POST "$WEBHOOK_URL" \
  -H "Content-Type: application/xml" \
  -d '<?xml version="1.0"?><test><message>Hello from XML</message></test>')

if echo "$RESPONSE" | grep -q "success"; then
    echo -e "${GREEN}✓ POST запрос с XML успешно отправлен${NC}"
else
    echo -e "${RED}✗ Ошибка при отправке POST запроса с XML${NC}"
    echo "$RESPONSE"
fi

echo -e "\n${GREEN}=== Тестирование завершено ===${NC}"
echo -e "${YELLOW}Проверьте результаты на странице: $BASE_URL/webhook-test/$TOKEN${NC}\n"
