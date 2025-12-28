@echo off
echo ========================================
echo Запуск тестов AI промптов для создания интеграций
echo ========================================
echo.

cd /d "%~dp0\.."

echo Проверяем наличие Go...
go version >nul 2>&1
if errorlevel 1 (
    echo ОШИБКА: Go не установлен или не найден в PATH
    pause
    exit /b 1
)

echo Go найден. Запускаем тесты...
echo.

echo ----------------------------------------
echo 1. Основные тесты AI интеграций
echo ----------------------------------------
go test -v ./tests -run TestAIIntegrationPrompts -timeout 60s
if errorlevel 1 (
    echo ОШИБКА: Основные тесты не прошли
    goto :error
)

echo.
echo ----------------------------------------
echo 2. Тесты валидации промптов
echo ----------------------------------------
go test -v ./tests -run TestPromptValidation -timeout 60s
if errorlevel 1 (
    echo ОШИБКА: Тесты валидации не прошли
    goto :error
)

echo.
echo ----------------------------------------
echo 3. Тесты граничных случаев
echo ----------------------------------------
go test -v ./tests -run TestPromptEdgeCases -timeout 60s
if errorlevel 1 (
    echo ОШИБКА: Тесты граничных случаев не прошли
    goto :error
)

echo.
echo ----------------------------------------
echo 4. Тесты консистентности
echo ----------------------------------------
go test -v ./tests -run TestPromptConsistency -timeout 60s
if errorlevel 1 (
    echo ОШИБКА: Тесты консистентности не прошли
    goto :error
)

echo.
echo ----------------------------------------
echo 5. Тесты форматирования промптов
echo ----------------------------------------
go test -v ./tests -run TestPromptFormatting -timeout 60s
if errorlevel 1 (
    echo ОШИБКА: Тесты форматирования не прошли
    goto :error
)

echo.
echo ----------------------------------------
echo 6. Тесты валидации шаблонов
echo ----------------------------------------
go test -v ./tests -run TestPromptTemplateValidation -timeout 60s
if errorlevel 1 (
    echo ОШИБКА: Тесты валидации шаблонов не прошли
    goto :error
)

echo.
echo ----------------------------------------
echo 7. Тесты популярных API
echo ----------------------------------------
go test -v ./tests -run TestPopularAPIs -timeout 30s
if errorlevel 1 (
    echo ОШИБКА: Тесты популярных API не прошли
    goto :error
)

echo.
echo ----------------------------------------
echo 8. Тесты быстрых предложений
echo ----------------------------------------
go test -v ./tests -run TestQuickSuggestions -timeout 30s
if errorlevel 1 (
    echo ОШИБКА: Тесты быстрых предложений не прошли
    goto :error
)

echo.
echo ----------------------------------------
echo 9. Тесты системных сообщений
echo ----------------------------------------
go test -v ./tests -run TestPromptSystemMessages -timeout 30s
if errorlevel 1 (
    echo ОШИБКА: Тесты системных сообщений не прошли
    goto :error
)

echo.
echo ========================================
echo ✅ ВСЕ ТЕСТЫ ПРОШЛИ УСПЕШНО!
echo ========================================
echo.
echo Результаты тестирования:
echo - Основные функции AI создания интеграций: ✅
echo - Валидация промптов и результатов: ✅
echo - Обработка граничных случаев: ✅
echo - Консистентность результатов: ✅
echo - Форматирование промптов: ✅
echo - Валидация шаблонов: ✅
echo - Популярные API: ✅
echo - Быстрые предложения: ✅
echo - Системные сообщения: ✅
echo.
echo Все компоненты AI для создания интеграций работают корректно.
echo.
pause
exit /b 0

:error
echo.
echo ========================================
echo ❌ ТЕСТЫ НЕ ПРОШЛИ
echo ========================================
echo.
echo Проверьте логи выше для получения подробной информации об ошибках.
echo.
pause
exit /b 1