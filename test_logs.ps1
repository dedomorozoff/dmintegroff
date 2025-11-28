# PowerShell скрипт для тестирования логирования запросов

Write-Host "Отправка тестового запроса на /test endpoint..." -ForegroundColor Green

$body = @{
    test = "data"
    timestamp = (Get-Date -Format "o")
    message = "Тестовый запрос для проверки логирования"
} | ConvertTo-Json

$headers = @{
    "Content-Type" = "application/json"
    "X-Custom-Header" = "test-value"
    "User-Agent" = "TestScript/1.0"
}

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/test" `
        -Method Post `
        -Headers $headers `
        -Body $body

    Write-Host "`nОтвет сервера:" -ForegroundColor Yellow
    $response | ConvertTo-Json -Depth 10

    Write-Host "`nТестовый запрос отправлен успешно!" -ForegroundColor Green
    Write-Host "Откройте http://localhost:8080/logs для просмотра результата" -ForegroundColor Cyan
} catch {
    Write-Host "`nОшибка при отправке запроса:" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
}
