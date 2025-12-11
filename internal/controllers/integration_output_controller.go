package controllers

import (
	"dmintegroff/internal/database"
	"dmintegroff/internal/models"
	"dmintegroff/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// IntegrationOutputsList - список выходов для интеграции
func IntegrationOutputsList(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	integrationIDStr := c.Param("id")
	integrationID, _ := strconv.ParseUint(integrationIDStr, 10, 32)

	// Проверяем доступ к интеграции
	var integration models.Integration
	query := database.DB
	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}
	if err := query.First(&integration, uint(integrationID)).Error; err != nil {
		c.HTML(http.StatusNotFound, "pages/404.html", gin.H{"title": "Интеграция не найдена"})
		return
	}

	// Получаем все выходы для этой интеграции
	var outputs []models.IntegrationOutput
	database.DB.Where("integration_id = ?", integrationID).Order("priority ASC, id ASC").Find(&outputs)

	c.HTML(http.StatusOK, "pages/integration_outputs.html", gin.H{
		"title":       "Выходы интеграции",
		"CurrentPage": "integrations",
		"integration": integration,
		"outputs":     outputs,
		"username":    session.Get("username"),
		"role":        role,
		"appURL":      getAppURL(c),
		"appPath":     getAppPath(),
	})
}

// IntegrationOutputCreate - форма создания выхода
func IntegrationOutputCreate(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	integrationIDStr := c.Param("id")
	integrationID, _ := strconv.ParseUint(integrationIDStr, 10, 32)

	// Проверяем доступ к интеграции
	var integration models.Integration
	query := database.DB
	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}
	if err := query.First(&integration, uint(integrationID)).Error; err != nil {
		c.HTML(http.StatusNotFound, "pages/404.html", gin.H{"title": "Интеграция не найдена"})
		return
	}

	c.HTML(http.StatusOK, "pages/integration_output_create.html", gin.H{
		"title":       "Создание выхода",
		"CurrentPage": "integrations",
		"integration": integration,
		"username":    session.Get("username"),
		"role":        role,
		"appURL":      getAppURL(c),
		"appPath":     getAppPath(),
	})
}

// IntegrationOutputStore - сохранение нового выхода
func IntegrationOutputStore(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	integrationIDStr := c.Param("id")
	integrationID, _ := strconv.ParseUint(integrationIDStr, 10, 32)

	// Проверяем доступ к интеграции
	var integration models.Integration
	query := database.DB
	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}
	if err := query.First(&integration, uint(integrationID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
		return
	}

	httpMethod := c.PostForm("http_method")
	if httpMethod == "" {
		httpMethod = "POST"
	}

	priorityStr := c.PostForm("priority")
	priority := 0
	if priorityStr != "" {
		priority, _ = strconv.Atoi(priorityStr)
	}

	output := models.IntegrationOutput{
		IntegrationID: uint(integrationID),
		Name:          c.PostForm("name"),
		Description:   c.PostForm("description"),
		TargetAPI:     c.PostForm("target_api"),
		HTTPMethod:    httpMethod,
		Priority:      priority,
		Enabled:       true,
		
		// Authentication
		AuthType:           c.PostForm("auth_type"),
		OAuth2TokenURL:     c.PostForm("oauth2_token_url"),
		OAuth2ClientID:     c.PostForm("oauth2_client_id"),
		OAuth2ClientSecret: c.PostForm("oauth2_client_secret"),
		OAuth2Scope:        c.PostForm("oauth2_scope"),
		OAuth2GrantType:    c.PostForm("oauth2_grant_type"),
		BearerToken:        c.PostForm("bearer_token"),
		BasicAuthUser:      c.PostForm("basic_auth_user"),
		BasicAuthPass:      c.PostForm("basic_auth_pass"),
		CustomHeaders:      c.PostForm("custom_headers"),
	}

	if err := database.DB.Create(&output).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/integrations/%d/outputs", integrationID))
}

// IntegrationOutputConfigure - настройка маппинга для выхода
func IntegrationOutputConfigure(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	integrationIDStr := c.Param("id")
	outputIDStr := c.Param("output_id")
	integrationID, _ := strconv.ParseUint(integrationIDStr, 10, 32)
	outputID, _ := strconv.ParseUint(outputIDStr, 10, 32)

	// Проверяем доступ к интеграции
	var integration models.Integration
	query := database.DB
	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}
	if err := query.First(&integration, uint(integrationID)).Error; err != nil {
		c.HTML(http.StatusNotFound, "pages/404.html", gin.H{"title": "Интеграция не найдена"})
		return
	}

	// Получаем выход
	var output models.IntegrationOutput
	if err := database.DB.Where("integration_id = ?", integrationID).First(&output, uint(outputID)).Error; err != nil {
		c.HTML(http.StatusNotFound, "pages/404.html", gin.H{"title": "Выход не найден"})
		return
	}

	// Используем SamplePayload из интеграции для настройки маппинга
	var fields []utils.FieldInfo
	var sampleData map[string]interface{}

	if integration.SamplePayload != "" {
		var err error
		fields, err = utils.ParseJSONString(integration.SamplePayload)
		if err != nil {
			fields = []utils.FieldInfo{}
		}
		json.Unmarshal([]byte(integration.SamplePayload), &sampleData)
	}

	// Подготавливаем payload для JavaScript
	var payloadJSON string
	if integration.SamplePayload != "" {
		payloadBytes, _ := json.Marshal(integration.SamplePayload)
		payloadJSON = string(payloadBytes)
	} else {
		payloadJSON = "null"
	}

	// Загружаем текущий маппинг
	var currentMapping map[string]string
	if output.MappingConfig != "" {
		json.Unmarshal([]byte(output.MappingConfig), &currentMapping)
	}

	c.HTML(http.StatusOK, "pages/integration_output_configure.html", gin.H{
		"title":          "Настройка маппинга выхода",
		"CurrentPage":    "integrations",
		"integration":    integration,
		"output":         output,
		"sampleData":     sampleData,
		"fields":         fields,
		"payloadJSON":    payloadJSON,
		"currentMapping": currentMapping,
		"username":       session.Get("username"),
		"role":           role,
		"appURL":         getAppURL(c),
		"appPath":        getAppPath(),
	})
}

// IntegrationOutputSaveMapping - сохранение маппинга для выхода
func IntegrationOutputSaveMapping(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	integrationIDStr := c.Param("id")
	outputIDStr := c.Param("output_id")
	integrationID, _ := strconv.ParseUint(integrationIDStr, 10, 32)
	outputID, _ := strconv.ParseUint(outputIDStr, 10, 32)

	// Проверяем доступ к интеграции
	var integration models.Integration
	query := database.DB
	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}
	if err := query.First(&integration, uint(integrationID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
		return
	}

	// Получаем выход
	var output models.IntegrationOutput
	if err := database.DB.Where("integration_id = ?", integrationID).First(&output, uint(outputID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Output not found"})
		return
	}

	// Получаем output_template или mapping_config
	outputTemplate := c.PostForm("output_template")
	mappingConfig := c.PostForm("mapping_config")
	templateType := c.PostForm("template_type")

	// Валидируем output_template, если он задан
	if outputTemplate != "" {
		if templateType == "json" || templateType == "" {
			processor := utils.NewTemplateProcessor()
			if err := processor.ValidateTemplate(outputTemplate); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid output template: " + err.Error()})
				return
			}
		}
		output.OutputTemplate = outputTemplate
		output.TemplateType = templateType
		if output.TemplateType == "" {
			output.TemplateType = "json"
		}
		output.MappingConfig = ""
	} else if mappingConfig != "" {
		output.MappingConfig = mappingConfig
		output.OutputTemplate = ""
		output.TemplateType = "json"
	}

	database.DB.Save(&output)

	c.Redirect(http.StatusFound, fmt.Sprintf("/integrations/%d/outputs", integrationID))
}

// IntegrationOutputEdit - редактирование выхода
func IntegrationOutputEdit(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	integrationIDStr := c.Param("id")
	outputIDStr := c.Param("output_id")
	integrationID, _ := strconv.ParseUint(integrationIDStr, 10, 32)
	outputID, _ := strconv.ParseUint(outputIDStr, 10, 32)

	// Проверяем доступ к интеграции
	var integration models.Integration
	query := database.DB
	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}
	if err := query.First(&integration, uint(integrationID)).Error; err != nil {
		c.HTML(http.StatusNotFound, "pages/404.html", gin.H{"title": "Интеграция не найдена"})
		return
	}

	// Получаем выход
	var output models.IntegrationOutput
	if err := database.DB.Where("integration_id = ?", integrationID).First(&output, uint(outputID)).Error; err != nil {
		c.HTML(http.StatusNotFound, "pages/404.html", gin.H{"title": "Выход не найден"})
		return
	}

	c.HTML(http.StatusOK, "pages/integration_output_edit.html", gin.H{
		"title":       "Редактирование выхода",
		"CurrentPage": "integrations",
		"integration": integration,
		"output":      output,
		"username":    session.Get("username"),
		"role":        role,
		"appURL":      getAppURL(c),
		"appPath":     getAppPath(),
	})
}

// IntegrationOutputUpdate - обновление выхода
func IntegrationOutputUpdate(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	integrationIDStr := c.Param("id")
	outputIDStr := c.Param("output_id")
	integrationID, _ := strconv.ParseUint(integrationIDStr, 10, 32)
	outputID, _ := strconv.ParseUint(outputIDStr, 10, 32)

	// Проверяем доступ к интеграции
	var integration models.Integration
	query := database.DB
	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}
	if err := query.First(&integration, uint(integrationID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
		return
	}

	// Получаем выход
	var output models.IntegrationOutput
	if err := database.DB.Where("integration_id = ?", integrationID).First(&output, uint(outputID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Output not found"})
		return
	}

	httpMethod := c.PostForm("http_method")
	if httpMethod == "" {
		httpMethod = "POST"
	}

	priorityStr := c.PostForm("priority")
	priority := 0
	if priorityStr != "" {
		priority, _ = strconv.Atoi(priorityStr)
	}

	output.Name = c.PostForm("name")
	output.Description = c.PostForm("description")
	output.TargetAPI = c.PostForm("target_api")
	output.HTTPMethod = httpMethod
	output.Priority = priority
	output.Condition = c.PostForm("condition")
	
	// Update authentication
	output.AuthType = c.PostForm("auth_type")
	output.OAuth2TokenURL = c.PostForm("oauth2_token_url")
	output.OAuth2ClientID = c.PostForm("oauth2_client_id")
	
	if newSecret := c.PostForm("oauth2_client_secret"); newSecret != "" {
		output.OAuth2ClientSecret = newSecret
	}
	
	output.OAuth2Scope = c.PostForm("oauth2_scope")
	output.OAuth2GrantType = c.PostForm("oauth2_grant_type")
	
	if newToken := c.PostForm("bearer_token"); newToken != "" {
		output.BearerToken = newToken
	}
	
	output.BasicAuthUser = c.PostForm("basic_auth_user")
	
	if newPass := c.PostForm("basic_auth_pass"); newPass != "" {
		output.BasicAuthPass = newPass
	}
	
	output.CustomHeaders = c.PostForm("custom_headers")

	database.DB.Save(&output)

	c.Redirect(http.StatusFound, fmt.Sprintf("/integrations/%d/outputs", integrationID))
}

// IntegrationOutputDelete - удаление выхода
func IntegrationOutputDelete(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	integrationIDStr := c.Param("id")
	outputIDStr := c.Param("output_id")
	integrationID, _ := strconv.ParseUint(integrationIDStr, 10, 32)
	outputID, _ := strconv.ParseUint(outputIDStr, 10, 32)

	// Проверяем доступ к интеграции
	var integration models.Integration
	query := database.DB
	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}
	if err := query.First(&integration, uint(integrationID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
		return
	}

	// Удаляем выход
	database.DB.Where("integration_id = ? AND id = ?", integrationID, outputID).Delete(&models.IntegrationOutput{})

	c.Redirect(http.StatusFound, fmt.Sprintf("/integrations/%d/outputs", integrationID))
}

// IntegrationOutputToggle - включение/выключение выхода
func IntegrationOutputToggle(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	integrationIDStr := c.Param("id")
	outputIDStr := c.Param("output_id")
	integrationID, _ := strconv.ParseUint(integrationIDStr, 10, 32)
	outputID, _ := strconv.ParseUint(outputIDStr, 10, 32)

	// Проверяем доступ к интеграции
	var integration models.Integration
	query := database.DB
	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}
	if err := query.First(&integration, uint(integrationID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
		return
	}

	// Получаем выход
	var output models.IntegrationOutput
	if err := database.DB.Where("integration_id = ?", integrationID).First(&output, uint(outputID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Output not found"})
		return
	}

	// Переключаем состояние
	output.Enabled = !output.Enabled
	database.DB.Save(&output)

	c.Redirect(http.StatusFound, fmt.Sprintf("/integrations/%d/outputs", integrationID))
}

// IntegrationOutputTest - тестирование выхода
func IntegrationOutputTest(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	role := session.Get("role")
	integrationIDStr := c.Param("id")
	outputIDStr := c.Param("output_id")
	integrationID, _ := strconv.ParseUint(integrationIDStr, 10, 32)
	outputID, _ := strconv.ParseUint(outputIDStr, 10, 32)

	// Проверяем доступ к интеграции
	var integration models.Integration
	query := database.DB
	if role != "admin" {
		query = query.Where("created_by_id = ?", userID)
	}
	if err := query.First(&integration, uint(integrationID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
		return
	}

	// Получаем выход
	var output models.IntegrationOutput
	if err := database.DB.Where("integration_id = ?", integrationID).First(&output, uint(outputID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Output not found"})
		return
	}

	// Проверяем наличие SamplePayload
	if integration.SamplePayload == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No sample payload available"})
		return
	}

	// Парсим sample payload
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(integration.SamplePayload), &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sample payload JSON"})
		return
	}

	// Обрабатываем трансформацию
	var transformed map[string]interface{}
	var transformedString string
	var isStringTemplate bool

	if output.OutputTemplate != "" {
		processor := utils.NewTemplateProcessor()
		templateType := output.TemplateType
		if templateType == "" {
			templateType = "json"
		}

		if templateType == "json" {
			var err error
			transformed, err = processor.ProcessTemplate(output.OutputTemplate, payload)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Template processing failed: " + err.Error(),
				})
				return
			}
		} else {
			var err error
			transformedString, err = processor.ProcessTemplateString(output.OutputTemplate, payload)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Template processing failed: " + err.Error(),
				})
				return
			}
			isStringTemplate = true
		}
	} else if output.MappingConfig != "" {
		var mapping map[string]string
		if err := json.Unmarshal([]byte(output.MappingConfig), &mapping); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid mapping config"})
			return
		}

		transformed = make(map[string]interface{})
		for targetField, sourceField := range mapping {
			if v, err := utils.GetValueByPath(payload, sourceField); err == nil {
				transformed[targetField] = v
			}
		}
	} else {
		transformed = payload
	}

	// Подготавливаем ответ
	var transformedOutput string
	if isStringTemplate {
		transformedOutput = transformedString
	} else {
		transformedJSON, _ := json.MarshalIndent(transformed, "", "  ")
		transformedOutput = string(transformedJSON)
	}

	inputJSON, _ := json.MarshalIndent(payload, "", "  ")

	c.JSON(http.StatusOK, gin.H{
		"status":        "success",
		"input":         string(inputJSON),
		"output":        transformedOutput,
		"target_api":    output.TargetAPI,
		"http_method":   output.HTTPMethod,
		"template_type": output.TemplateType,
	})
}
