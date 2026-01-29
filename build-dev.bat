@echo off
echo Сборка для разработки (без встроенных файлов)...

REM Временно переименовываем пакет assets чтобы не использовать embed
if exist internal\assets\embed.go ren internal\assets\embed.go embed.go.bak

REM Возвращаем старые роуты для разработки
echo Восстанавливаем роуты для разработки...

REM Создаем временный файл routes для разработки
copy internal\routes\routes.go internal\routes\routes.go.bak

REM Собираем без embed
go build -o dmintegroff-dev.exe cmd/server/main.go

if %ERRORLEVEL% EQU 0 (
    echo Сборка для разработки завершена успешно!
    echo Размер исполняемого файла:
    dir dmintegroff-dev.exe | findstr dmintegroff-dev.exe
    echo.
    echo ВНИМАНИЕ: Для работы требуются папки static/, templates/, integration-templates/
) else (
    echo Ошибка при сборке!
    exit /b 1
)

REM Восстанавливаем embed.go
if exist internal\assets\embed.go.bak ren internal\assets\embed.go.bak embed.go