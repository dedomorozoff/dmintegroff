@echo off
echo Обновление встроенных файлов...

REM Удаляем старые копии
if exist internal\assets\static rmdir /s /q internal\assets\static
if exist internal\assets\templates rmdir /s /q internal\assets\templates
if exist internal\assets\integration-templates rmdir /s /q internal\assets\integration-templates

REM Копируем актуальные файлы
xcopy static internal\assets\static\ /E /I /Q
xcopy templates internal\assets\templates\ /E /I /Q
xcopy integration-templates internal\assets\integration-templates\ /E /I /Q

echo Сборка проекта...
go build -o dmintegroff.exe cmd/server/main.go

if %ERRORLEVEL% EQU 0 (
    echo Сборка завершена успешно!
    echo Размер исполняемого файла:
    dir dmintegroff.exe | findstr dmintegroff.exe
) else (
    echo Ошибка при сборке!
    exit /b 1
)