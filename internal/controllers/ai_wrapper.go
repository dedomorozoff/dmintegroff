package controllers

import (
	"dmintegroff/internal/config"
	"sync"

	"github.com/gin-gonic/gin"
)

// Глобальный AI контроллер (singleton)
var (
	aiController *AIController
	aiOnce       sync.Once
)

// getAIController возвращает singleton AI контроллера
func getAIController() *AIController {
	aiOnce.Do(func() {
		aiConfig := config.LoadAIConfig()
		aiController = NewAIController(aiConfig)
	})
	return aiController
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
	getAIController().GetStatus(c)
}

// AIGetQuickSuggestions возвращает быстрые предложения
func AIGetQuickSuggestions(c *gin.Context) {
	getAIController().GetQuickSuggestions(c)
}