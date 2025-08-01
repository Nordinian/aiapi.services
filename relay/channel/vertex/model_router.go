package vertex

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

// ModelSession 存储会话的模型状态
type ModelSession struct {
	SessionID string
	Model     string
	UpdatedAt time.Time
}

// ModelRouter 处理模型路由和会话管理
type ModelRouter struct {
	sessions map[string]*ModelSession
	mutex    sync.RWMutex
}

// NewModelRouter 创建新的模型路由器
func NewModelRouter() *ModelRouter {
	return &ModelRouter{
		sessions: make(map[string]*ModelSession),
	}
}

// SupportedModels 定义支持的模型别名到实际模型ID的映射
var SupportedModels = map[string]string{
	// Gemini模型
	"gemini-2.0-flash":      "gemini-2.0-flash",
	"gemini-2.5-pro":        "gemini-2.5-pro",
	"gemini-2.5-flash":      "gemini-2.5-flash",
	"gemini-2.0-flash-lite": "gemini-2.0-flash-lite",
	
	// Claude模型
	"claude-3-7-sonnet": "claude-3-7-sonnet",
	"claude-opus-4":     "claude-opus-4",
	"claude-sonnet-4":   "claude-sonnet-4",
	
	// 简短别名
	"gemini": "gemini-2.5-pro",
	"claude": "claude-sonnet-4",
	"sonnet": "claude-sonnet-4",
	"opus":   "claude-opus-4",
	"flash":  "gemini-2.5-flash",
	"pro":    "gemini-2.5-pro",
	"lite":   "gemini-2.0-flash-lite",
}

// ModelCommandRegex 匹配/model命令的正则表达式
var ModelCommandRegex = regexp.MustCompile(`(?i)^/model\s+([a-zA-Z0-9\-\.]+)(?:\s|$)`)

// ParseModelCommand 从消息中解析/model命令
func (mr *ModelRouter) ParseModelCommand(message string) (string, string, bool) {
	matches := ModelCommandRegex.FindStringSubmatch(strings.TrimSpace(message))
	if len(matches) < 2 {
		return "", message, false
	}
	
	modelName := strings.ToLower(strings.TrimSpace(matches[1]))
	normalizedModel := mr.normalizeModelName(modelName)
	
	if normalizedModel == "" {
		return "", message, false
	}
	
	// 移除/model命令，返回剩余消息
	remainingMessage := strings.TrimSpace(ModelCommandRegex.ReplaceAllString(message, ""))
	
	return normalizedModel, remainingMessage, true
}

// normalizeModelName 将用户输入的模型名称标准化为实际模型ID
func (mr *ModelRouter) normalizeModelName(modelName string) string {
	modelName = strings.ToLower(strings.TrimSpace(modelName))
	
	// 检查是否为别名或完整模型名
	if actualModel, exists := SupportedModels[modelName]; exists {
		return actualModel
	}
	
	// 模糊匹配
	for alias, actualModel := range SupportedModels {
		if strings.Contains(modelName, alias) {
			return actualModel
		}
	}
	
	return ""
}

// SetSessionModel 设置会话的模型
func (mr *ModelRouter) SetSessionModel(sessionID, model string) {
	mr.mutex.Lock()
	defer mr.mutex.Unlock()
	
	mr.sessions[sessionID] = &ModelSession{
		SessionID: sessionID,
		Model:     model,
		UpdatedAt: time.Now(),
	}
}

// GetSessionModel 获取会话的模型
func (mr *ModelRouter) GetSessionModel(sessionID string) (string, bool) {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()
	
	session, exists := mr.sessions[sessionID]
	if !exists {
		return "", false
	}
	
	return session.Model, true
}

// ExtractModelFromEnvironment 从环境变量/请求头中提取模型配置
func (mr *ModelRouter) ExtractModelFromEnvironment(headers map[string]string) string {
	// 首先检查系统环境变量
	envVars := []string{
		"CLAUDE_CUSTOM_MODEL",
		"CLAUDE_MODEL",
		"ANTHROPIC_MODEL",
		"AI_MODEL",
	}
	
	for _, envVar := range envVars {
		if value := os.Getenv(envVar); value != "" {
			normalized := mr.normalizeModelName(value)
			fmt.Printf("[DEBUG] 环境变量 %s=%s, 标准化后: %s\n", envVar, value, normalized)
			if normalized != "" {
				return normalized
			}
		}
	}
	
	// 然后检查HTTP请求头
	envHeaders := []string{
		"X-Claude-Custom-Model",
		"X-Claude-Model",
		"X-Model",
		"Claude-Custom-Model",
		"Claude-Model", 
		"Model",
		"X-Anthropic-Model",
	}
	
	for _, header := range envHeaders {
		if value, exists := headers[header]; exists && value != "" {
			normalized := mr.normalizeModelName(value)
			if normalized != "" {
				return normalized
			}
		}
	}
	
	return ""
}

// IsGeminiModel 检查是否为Gemini模型
func (mr *ModelRouter) IsGeminiModel(model string) bool {
	return strings.Contains(model, "gemini")
}

// IsClaudeModel 检查是否为Claude模型
func (mr *ModelRouter) IsClaudeModel(model string) bool {
	return strings.Contains(model, "claude")
}

// GetDefaultModel 获取默认模型
func (mr *ModelRouter) GetDefaultModel() string {
	return "claude-sonnet-4"
}

// ValidateModel 验证模型是否支持
func (mr *ModelRouter) ValidateModel(model string) bool {
	// 检查是否在支持的模型列表中
	for _, actualModel := range SupportedModels {
		if actualModel == model {
			return true
		}
	}
	return false
}

// GetSupportedModelsList 获取支持的模型列表
func (mr *ModelRouter) GetSupportedModelsList() []string {
	models := make([]string, 0, len(SupportedModels))
	for alias := range SupportedModels {
		models = append(models, alias)
	}
	return models
}

// CleanupOldSessions 清理超过24小时的旧会话
func (mr *ModelRouter) CleanupOldSessions() {
	mr.mutex.Lock()
	defer mr.mutex.Unlock()
	
	cutoff := time.Now().Add(-24 * time.Hour)
	for sessionID, session := range mr.sessions {
		if session.UpdatedAt.Before(cutoff) {
			delete(mr.sessions, sessionID)
		}
	}
}

// GetModelUsageStats 获取模型使用统计
func (mr *ModelRouter) GetModelUsageStats() map[string]int {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()
	
	stats := make(map[string]int)
	for _, session := range mr.sessions {
		stats[session.Model]++
	}
	return stats
}

// FormatModelHelpMessage 格式化模型帮助信息
func (mr *ModelRouter) FormatModelHelpMessage() string {
	return fmt.Sprintf(`Available Models:

Gemini Models:
- gemini-2.0-flash, gemini-2.5-pro, gemini-2.5-flash, gemini-2.0-flash-lite

Claude Models:
- claude-3-7-sonnet, claude-opus-4, claude-sonnet-4

Short Aliases:
- gemini → gemini-2.5-pro
- flash → gemini-2.5-flash
- pro → gemini-2.5-pro
- lite → gemini-2.0-flash-lite
- claude, sonnet → claude-sonnet-4
- opus → claude-opus-4

Usage:
- /model gemini-2.0-flash
- /model claude-sonnet-4
- /model flash (uses gemini-2.5-flash)

Environment Variable:
export CLAUDE_CUSTOM_MODEL="gemini-2.0-flash"
claude "your prompt here"
`)
}