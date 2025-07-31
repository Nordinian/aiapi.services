package vertex

import (
	"one-api/common"
	"one-api/dto"
	"reflect"
)

type VertexAIClaudeRequest struct {
	AnthropicVersion string              `json:"anthropic_version"`
	Messages         []dto.ClaudeMessage `json:"messages"`
	System           any                 `json:"system,omitempty"`
	MaxTokens        uint                `json:"max_tokens,omitempty"`
	StopSequences    []string            `json:"stop_sequences,omitempty"`
	Stream           bool                `json:"stream,omitempty"`
	Temperature      *float64            `json:"temperature,omitempty"`
	TopP             float64             `json:"top_p,omitempty"`
	TopK             int                 `json:"top_k,omitempty"`
	Tools            any                 `json:"tools,omitempty"`
	ToolChoice       any                 `json:"tool_choice,omitempty"`
	Thinking         *dto.Thinking       `json:"thinking,omitempty"`
}

func copyRequest(req *dto.ClaudeRequest, version string) *VertexAIClaudeRequest {
	// Filter out unsupported tools for Vertex AI
	filteredTools := filterVertexAITools(req.Tools)
	
	// Handle tool_choice compatibility
	filteredToolChoice := filterVertexAIToolChoice(req.ToolChoice, filteredTools)
	
	// Handle Thinking compatibility - Vertex AI might not support all thinking features
	filteredThinking := filterVertexAIThinking(req.Thinking, version)
	
	// Filter messages for Vertex AI compatibility (remove unsupported features like cache_control)
	filteredMessages := filterVertexAIMessages(req.Messages, version)
	
	// Filter system message for compatibility
	filteredSystem := filterVertexAISystem(req.System, version)
	
	// Log compatibility adjustments
	logCompatibilityAdjustments(req, filteredTools, filteredToolChoice, filteredThinking)
	
	return &VertexAIClaudeRequest{
		AnthropicVersion: version,
		System:           filteredSystem,
		Messages:         filteredMessages,
		MaxTokens:        req.MaxTokens,
		Stream:           req.Stream,
		Temperature:      req.Temperature,
		TopP:             req.TopP,
		TopK:             req.TopK,
		StopSequences:    req.StopSequences,
		Tools:            filteredTools,
		ToolChoice:       filteredToolChoice,
		Thinking:         filteredThinking,
	}
}

// filterVertexAITools filters out tools not supported by Vertex AI
func filterVertexAITools(tools any) any {
	if tools == nil {
		return nil
	}

	// Handle slice of tools
	if toolsSlice, ok := tools.([]any); ok {
		var filteredTools []any
		for _, tool := range toolsSlice {
			if !isUnsupportedTool(tool) {
				filteredTools = append(filteredTools, tool)
			}
		}
		// Return nil if no tools remain after filtering
		if len(filteredTools) == 0 {
			return nil
		}
		return filteredTools
	}

	// Handle single tool
	if !isUnsupportedTool(tools) {
		return tools
	}

	return nil
}

// isUnsupportedTool checks if a tool is not supported by Vertex AI based on API version
func isUnsupportedTool(tool any) bool {
	if tool == nil {
		return false
	}

	// Since user confirmed Vertex AI supports WebSearch, disable all tool filtering
	// to allow WebSearch tools to be sent to Vertex AI
	return false
}

// isWebSearchTool checks if a tool is a WebSearch tool
func isWebSearchTool(tool any) bool {
	if tool == nil {
		return false
	}

	// Check for ClaudeWebSearchTool
	if webSearchTool, ok := tool.(*dto.ClaudeWebSearchTool); ok {
		return webSearchTool.Type == "web_search_20250305"
	}

	// Check for tool maps containing web search
	if toolMap, ok := tool.(map[string]any); ok {
		if toolType, exists := toolMap["type"]; exists {
			if typeStr, ok := toolType.(string); ok {
				return typeStr == "web_search_20250305"
			}
		}
		if toolName, exists := toolMap["name"]; exists {
			if nameStr, ok := toolName.(string); ok {
				return nameStr == "web_search"
			}
		}
	}

	// Use reflection to check struct fields
	v := reflect.ValueOf(tool)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() == reflect.Struct {
		typeField := v.FieldByName("Type")
		nameField := v.FieldByName("Name")
		
		if typeField.IsValid() && typeField.Kind() == reflect.String {
			if typeField.String() == "web_search_20250305" {
				return true
			}
		}
		
		if nameField.IsValid() && nameField.Kind() == reflect.String {
			if nameField.String() == "web_search" {
				return true
			}
		}
	}

	return false
}

// filterVertexAIToolChoice handles tool_choice compatibility for Vertex AI
func filterVertexAIToolChoice(toolChoice any, filteredTools any) any {
	if toolChoice == nil {
		return nil
	}

	// If no tools remain after filtering, remove tool_choice
	if filteredTools == nil {
		return nil
	}

	// Check if toolChoice references a specific tool that was filtered out
	if toolChoiceMap, ok := toolChoice.(map[string]any); ok {
		if toolName, exists := toolChoiceMap["name"]; exists {
			if nameStr, ok := toolName.(string); ok {
				// If tool_choice references web_search which was filtered, remove it
				if nameStr == "web_search" {
					return nil
				}
			}
		}
	}

	// For other tool choice types, pass through
	return toolChoice
}

// filterVertexAIThinking handles Thinking feature compatibility for Vertex AI
func filterVertexAIThinking(thinking *dto.Thinking, version string) *dto.Thinking {
	if thinking == nil {
		return nil
	}

	// Check if Vertex AI version supports thinking
	// vertex-2023-10-16 might not support all thinking features
	if version == "vertex-2023-10-16" {
		// For older versions, we might need to limit thinking features
		// For now, pass through but log a warning
		common.SysLog("Vertex AI Thinking feature compatibility: using version " + version)
	}

	return thinking
}

// logCompatibilityAdjustments logs any compatibility adjustments made for debugging
func logCompatibilityAdjustments(originalReq *dto.ClaudeRequest, filteredTools, filteredToolChoice any, filteredThinking *dto.Thinking) {
	if common.DebugEnabled {
		adjustments := []string{}

		// Check if tools were filtered
		if originalReq.Tools != nil && filteredTools == nil {
			adjustments = append(adjustments, "removed all tools (WebSearch not supported)")
		} else if originalReq.Tools != nil && filteredTools != nil {
			// Check if some tools were filtered
			originalCount := countTools(originalReq.Tools)
			filteredCount := countTools(filteredTools)
			if originalCount > filteredCount {
				adjustments = append(adjustments, "filtered out unsupported tools")
			}
		}

		// Check if tool_choice was filtered
		if originalReq.ToolChoice != nil && filteredToolChoice == nil {
			adjustments = append(adjustments, "removed tool_choice (referenced filtered tool)")
		}

		// Check if thinking was modified
		if originalReq.Thinking != nil && filteredThinking != originalReq.Thinking {
			adjustments = append(adjustments, "modified thinking parameters for Vertex AI compatibility")
		}

		if len(adjustments) > 0 {
			common.SysLog("Vertex AI compatibility adjustments: " + joinStrings(adjustments, ", "))
		}
	}
}

// countTools counts the number of tools in a tools interface
func countTools(tools any) int {
	if tools == nil {
		return 0
	}

	if toolsSlice, ok := tools.([]any); ok {
		return len(toolsSlice)
	}

	// Single tool
	return 1
}

// joinStrings joins string slice with separator (simple implementation)
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// filterVertexAIMessages filters message content for Vertex AI compatibility
func filterVertexAIMessages(messages []dto.ClaudeMessage, version string) []dto.ClaudeMessage {
	if len(messages) == 0 {
		return messages
	}

	filteredMessages := make([]dto.ClaudeMessage, len(messages))
	for i, msg := range messages {
		filteredMessages[i] = dto.ClaudeMessage{
			Role:    msg.Role,
			Content: filterMessageContent(msg.Content, version),
		}
	}

	return filteredMessages
}

// filterVertexAISystem filters system content for Vertex AI compatibility
func filterVertexAISystem(system any, version string) any {
	if system == nil {
		return nil
	}

	// For older Vertex AI versions, system might need special handling
	return filterMessageContent(system, version)
}

// filterMessageContent removes unsupported features from message content
func filterMessageContent(content any, version string) any {
	if content == nil {
		return nil
	}

	// Handle string content - pass through as is
	if _, ok := content.(string); ok {
		return content
	}

	// Handle array content - filter each item
	if contentArray, ok := content.([]any); ok {
		filteredArray := make([]any, 0, len(contentArray))
		
		for _, item := range contentArray {
			if filteredItem := filterContentItem(item, version); filteredItem != nil {
				filteredArray = append(filteredArray, filteredItem)
			}
		}
		
		return filteredArray
	}

	// Handle single content item
	return filterContentItem(content, version)
}

// filterContentItem filters individual content items
func filterContentItem(item any, version string) any {
	if item == nil {
		return nil
	}

	// Handle map content items
	if itemMap, ok := item.(map[string]any); ok {
		filteredItem := make(map[string]any)
		
		for key, value := range itemMap {
			// Skip unsupported fields for older Vertex AI versions
			if version == "vertex-2023-10-16" && key == "cache_control" {
				// Cache control not supported in older versions
				continue
			}
			
			filteredItem[key] = value
		}
		
		return filteredItem
	}

	// For other types, pass through
	return item
}
