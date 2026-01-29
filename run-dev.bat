@echo off
echo Запуск в режиме разработки (без встроенных файлов)...

REM Устанавливаем переменную окружения для использования файловой системы
set USE_EMBEDDED_FILES=false

REM Запускаем сервер
go run cmd/server/main.go