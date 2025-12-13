package controllers

import (
	"dmintegroff/internal/config"
	"dmintegroff/internal/database"
	"sync"

	"github.com/gin-gonic/gin"
)

// Глобальный AI контроллер (singleton)
var (
	aiController *AIController
	aiMutex      sync.RWMutex
)

// getAIController возвращает AI контроллер с актуальной конфигурацией
func getAIController() *AIController {
	aiMutex.RLock()
	if aiController != nil {
		defer aiMutex.RUnlock()
		return aiController
	}
	aiMutex.RUnlock()

	aiMutex.Lock()
	defer aiMutex.Unlock()
	
	// Двойная проверка после получения блокировки записи
	if aiController == nil {
		aiConfig := config.LoadAIConfig(database.DB)
		aiController = NewAIController(aiConfig)
	}
	
	return aiController
}

// reloadAIController перезагружает AI контроллер с новой конфигурацией
func reloadAIController() {
	aiMutex.Lock()
	defer aiMutex.Unlock()
	
	// Принудительно сбрасываем контроллер
	aiController = nil
	
	// Загружаем новую конфигурацию и создаем новый контроллер
	aiConfig := config.LoadAIConfig(database.DB)
	aiController = NewAIController(aiConfig)
}

// AIChat обрабатывает запросы к AI чату
func AIChat(c *gin.Context) {
	getAIController().Chat(c)
}

// AIAnalyzeData анализирует структуру данных
func AIAnalyzeData(c *gin.Context) {
	getAIController().AnalyzeData(c)
}

// AIGenerateMapping генерирует маппинг для интеграции
func AIGenerateMapping(c *gin.Context) {
	getAIController().GenerateMapping(c)
}

// AIApplyMapping применяет сгенерированный маппинг к интеграции
func AIApplyMapping(c *gin.Context) {
	getAIController().ApplyMapping(c)
}

// AIGetStatus возвращает статус AI сервиса
func AIGetStatus(c *gin.Context) {
	// Для статуса всегда используем свежую конфигурацию
	aiMutex.Lock()
	aiConfig := config.LoadAIConfig(database.DB)
	tempController := NewAIController(aiConfig)
	aiMutex.Unlock()
	
	tempController.GetStatus(c)
}

// AIGetQuickSuggestions возвращает быстрые предложения
func AIGetQuickSuggestions(c *gin.Context) {
	getAIController().GetQuickSuggestions(c)
}

// AIGetModels возвращает список доступных AI моделей
func AIGetModels(c *gin.Context) {
	getAIController().GetModels(c)
}