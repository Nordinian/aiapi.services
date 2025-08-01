package vertex

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"one-api/dto"
	"one-api/relay/channel"
	"one-api/relay/channel/claude"
	"one-api/relay/channel/gemini"
	"one-api/relay/channel/openai"
	relaycommon "one-api/relay/common"
	"one-api/relay/constant"
	"one-api/setting/model_setting"
	"one-api/types"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	RequestModeClaude = 1
	RequestModeGemini = 2
	RequestModeLlama  = 3
)

var claudeModelMap = map[string]string{
	"claude-3-sonnet-20240229":   "claude-3-sonnet@20240229",
	"claude-3-opus-20240229":     "claude-3-opus@20240229",
	"claude-3-haiku-20240307":    "claude-3-haiku@20240307",
	"claude-3-5-sonnet-20240620": "claude-3-5-sonnet@20240620",
	"claude-3-5-sonnet-20241022": "claude-3-5-sonnet-v2@20241022",
	"claude-3-7-sonnet-20250219": "claude-3-7-sonnet@20250219",
	"claude-sonnet-4-20250514":   "claude-sonnet-4@20250514",
	"claude-opus-4-20250514":     "claude-opus-4@20250514",
}

const anthropicVersion = "vertex-2023-10-16"

// RouteModel 处理模型路由逻辑
func (a *Adaptor) RouteModel(c *gin.Context, info *relaycommon.RelayInfo, messageContent string) string {
	if a.ModelRouter == nil {
		return info.UpstreamModelName
	}
	
	// 生成会话ID (基于请求上下文)
	sessionID := c.GetString("session_id")
	if sessionID == "" {
		sessionID = fmt.Sprintf("%s_%d", c.ClientIP(), c.Request.Header.Get("X-Request-Id"))
		c.Set("session_id", sessionID)
	}
	
	// 1. 检查消息中的/model命令
	if messageContent != "" {
		if modelFromCommand, remainingMessage, found := a.ModelRouter.ParseModelCommand(messageContent); found {
			// 设置会话模型
			a.ModelRouter.SetSessionModel(sessionID, modelFromCommand)
			// 更新消息内容（移除/model命令）
			c.Set("updated_message", remainingMessage)
			return modelFromCommand
		}
	}
	
	// 2. 检查环境变量/请求头
	headers := make(map[string]string)
	for key, values := range c.Request.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	
	if modelFromEnv := a.ModelRouter.ExtractModelFromEnvironment(headers); modelFromEnv != "" {
		a.ModelRouter.SetSessionModel(sessionID, modelFromEnv)
		return modelFromEnv
	}
	
	// 3. 检查会话中已设置的模型
	if sessionModel, found := a.ModelRouter.GetSessionModel(sessionID); found {
		return sessionModel
	}
	
	// 4. 返回默认模型或原始模型
	if info.UpstreamModelName == "" {
		return a.ModelRouter.GetDefaultModel()
	}
	
	return info.UpstreamModelName
}

type Adaptor struct {
	RequestMode        int
	AccountCredentials Credentials
	ModelRouter        *ModelRouter
}

func (a *Adaptor) ConvertClaudeRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.ClaudeRequest) (any, error) {
	fmt.Printf("[DEBUG] ConvertClaudeRequest called with model: %s\n", info.UpstreamModelName)
	
	// 提取第一条用户消息用于模型路由
	var messageContent string
	if len(request.Messages) > 0 {
		for _, msg := range request.Messages {
			if msg.Role == "user" {
				messageContent = msg.GetStringContent()
				if messageContent != "" {
					break
				}
			}
		}
	}
	
	// 执行模型路由
	routedModel := a.RouteModel(c, info, messageContent)
	fmt.Printf("[DEBUG] ConvertClaudeRequest - 原始模型: %s, 路由后模型: %s\n", info.UpstreamModelName, routedModel)
	
	// 检查是否需要强制转换为Gemini格式（即使模型名相同）
	if strings.HasPrefix(routedModel, "gemini") {
		fmt.Printf("[DEBUG] 检测到Gemini模型，强制转换请求格式\n")
		a.RequestMode = RequestModeGemini
		// 当路由到Gemini时，需要转换为Gemini格式
		c.Set("force_gemini_mode", true)
		
		// 将Claude请求直接转换为Gemini格式（包含完整的工具转换）
		claudeRequestMap := map[string]interface{}{
			"model":      routedModel,
			"max_tokens": request.MaxTokens,
			"temperature": request.Temperature,
			"messages":   request.Messages,
			"tools":      request.Tools,
			"system":     request.System,
		}
		
		geminiRequest, err := ConvertClaudeRequestToGeminiWithTools(claudeRequestMap)
		if err != nil {
			return nil, fmt.Errorf("failed to convert Claude request to Gemini format: %w", err)
		}
		
		c.Set("request_model", routedModel)
		return geminiRequest, nil
	}
	
	if routedModel != info.UpstreamModelName {
		info.UpstreamModelName = routedModel
		request.Model = routedModel
		// 更新请求模式
		if strings.HasPrefix(routedModel, "claude") {
			a.RequestMode = RequestModeClaude
		} else if strings.HasPrefix(routedModel, "gemini") {
			a.RequestMode = RequestModeGemini
			// 当路由到Gemini时，需要转换为Gemini格式
			c.Set("force_gemini_mode", true)
			
			// 将Claude请求直接转换为Gemini格式（包含完整的工具转换）
			claudeRequestMap := map[string]interface{}{
				"model":      request.Model,
				"max_tokens": request.MaxTokens,
				"temperature": request.Temperature,
				"messages":   request.Messages,
				"tools":      request.Tools,
				"system":     request.System,
			}
			
			geminiRequest, err := ConvertClaudeRequestToGeminiWithTools(claudeRequestMap)
			if err != nil {
				return nil, fmt.Errorf("failed to convert Claude request to Gemini format: %w", err)
			}
			
			c.Set("request_model", request.Model)
			return geminiRequest, nil
		}
	}
	
	// 如果消息被更新（移除了/model命令），需要更新请求
	if updatedMessage := c.GetString("updated_message"); updatedMessage != "" {
		for i := range request.Messages {
			if request.Messages[i].Role == "user" {
				currentContent := request.Messages[i].GetStringContent()
				if currentContent == messageContent {
					request.Messages[i].SetStringContent(updatedMessage)
					break
				}
			}
		}
	}
	
	if v, ok := claudeModelMap[info.UpstreamModelName]; ok {
		c.Set("request_model", v)
	} else {
		c.Set("request_model", request.Model)
	}
	
	// Transform WebSearch tools to supported tools for Vertex AI
	if request.Tools != nil {
		transformedTools, err := transformWebSearchTools(request.Tools)
		if err != nil {
			return nil, fmt.Errorf("failed to transform tools: %w", err)
		}
		request.Tools = transformedTools
	}
	
	vertexClaudeReq := copyRequest(request, anthropicVersion)
	return vertexClaudeReq, nil
}

func (a *Adaptor) ConvertAudioRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.AudioRequest) (io.Reader, error) {
	//TODO implement me
	return nil, errors.New("not implemented")
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.ImageRequest) (any, error) {
	//TODO implement me
	return nil, errors.New("not implemented")
}

func (a *Adaptor) Init(info *relaycommon.RelayInfo) {
	// Initialize ModelRouter if not already done
	if a.ModelRouter == nil {
		a.ModelRouter = NewModelRouter()
	}
	
	if strings.HasPrefix(info.UpstreamModelName, "claude") {
		a.RequestMode = RequestModeClaude
	} else if strings.HasPrefix(info.UpstreamModelName, "gemini") {
		a.RequestMode = RequestModeGemini
	} else if strings.Contains(info.UpstreamModelName, "llama") {
		a.RequestMode = RequestModeLlama
	}
}

func (a *Adaptor) GetRequestURL(info *relaycommon.RelayInfo) (string, error) {
	adc := &Credentials{}
	if err := json.Unmarshal([]byte(info.ApiKey), adc); err != nil {
		return "", fmt.Errorf("failed to decode credentials file: %w", err)
	}
	region := GetModelRegion(info.ApiVersion, info.OriginModelName)
	a.AccountCredentials = *adc
	suffix := ""
	if a.RequestMode == RequestModeGemini {
		if model_setting.GetGeminiSettings().ThinkingAdapterEnabled {
			// 新增逻辑：处理 -thinking-<budget> 格式
			if strings.Contains(info.UpstreamModelName, "-thinking-") {
				parts := strings.Split(info.UpstreamModelName, "-thinking-")
				info.UpstreamModelName = parts[0]
			} else if strings.HasSuffix(info.UpstreamModelName, "-thinking") { // 旧的适配
				info.UpstreamModelName = strings.TrimSuffix(info.UpstreamModelName, "-thinking")
			} else if strings.HasSuffix(info.UpstreamModelName, "-nothinking") {
				info.UpstreamModelName = strings.TrimSuffix(info.UpstreamModelName, "-nothinking")
			}
		}

		if info.IsStream {
			suffix = "streamGenerateContent?alt=sse"
		} else {
			suffix = "generateContent"
		}
		if region == "global" {
			return fmt.Sprintf(
				"https://aiplatform.googleapis.com/v1/projects/%s/locations/global/publishers/google/models/%s:%s",
				adc.ProjectID,
				info.UpstreamModelName,
				suffix,
			), nil
		} else {
			return fmt.Sprintf(
				"https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:%s",
				region,
				adc.ProjectID,
				region,
				info.UpstreamModelName,
				suffix,
			), nil
		}
	} else if a.RequestMode == RequestModeClaude {
		if info.IsStream {
			suffix = "streamRawPredict?alt=sse"
		} else {
			suffix = "rawPredict"
		}
		model := info.UpstreamModelName
		if v, ok := claudeModelMap[info.UpstreamModelName]; ok {
			model = v
		}
		if region == "global" {
			return fmt.Sprintf(
				"https://aiplatform.googleapis.com/v1/projects/%s/locations/global/publishers/anthropic/models/%s:%s",
				adc.ProjectID,
				model,
				suffix,
			), nil
		} else {
			return fmt.Sprintf(
				"https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/anthropic/models/%s:%s",
				region,
				adc.ProjectID,
				region,
				model,
				suffix,
			), nil
		}
	} else if a.RequestMode == RequestModeLlama {
		return fmt.Sprintf(
			"https://%s-aiplatform.googleapis.com/v1beta1/projects/%s/locations/%s/endpoints/openapi/chat/completions",
			region,
			adc.ProjectID,
			region,
		), nil
	}
	return "", errors.New("unsupported request mode")
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Header, info *relaycommon.RelayInfo) error {
	channel.SetupApiRequestHeader(info, c, req)
	accessToken, err := getAccessToken(a, info)
	if err != nil {
		return err
	}
	req.Set("Authorization", "Bearer "+accessToken)
	return nil
}

func (a *Adaptor) ConvertOpenAIRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	
	// 提取第一条用户消息用于模型路由
	var messageContent string
	if len(request.Messages) > 0 {
		for _, msg := range request.Messages {
			if msg.Role == "user" {
				if content, ok := msg.Content.(string); ok {
					messageContent = content
					break
				}
			}
		}
	}
	
	// 执行模型路由
	routedModel := a.RouteModel(c, info, messageContent)
	if routedModel != info.UpstreamModelName {
		info.UpstreamModelName = routedModel
		// 更新请求模式
		if strings.HasPrefix(routedModel, "claude") {
			a.RequestMode = RequestModeClaude
		} else if strings.HasPrefix(routedModel, "gemini") {
			a.RequestMode = RequestModeGemini
		}
	}
	
	// 如果消息被更新（移除了/model命令），需要更新请求
	if updatedMessage := c.GetString("updated_message"); updatedMessage != "" {
		for i, msg := range request.Messages {
			if msg.Role == "user" {
				if content, ok := msg.Content.(string); ok && content == messageContent {
					request.Messages[i].Content = updatedMessage
					break
				}
			}
		}
	}
	
	if a.RequestMode == RequestModeClaude {
		claudeReq, err := claude.RequestOpenAI2ClaudeMessage(*request)
		if err != nil {
			return nil, err
		}
		
		// Transform WebSearch tools to supported tools for Vertex AI
		if claudeReq.Tools != nil {
			transformedTools, err := transformWebSearchTools(claudeReq.Tools)
			if err != nil {
				return nil, fmt.Errorf("failed to transform tools: %w", err)
			}
			claudeReq.Tools = transformedTools
		}
		
		vertexClaudeReq := copyRequest(claudeReq, anthropicVersion)
		c.Set("request_model", claudeReq.Model)
		info.UpstreamModelName = claudeReq.Model
		return vertexClaudeReq, nil
	} else if a.RequestMode == RequestModeGemini {
		geminiRequest, err := gemini.CovertGemini2OpenAI(*request, info)
		if err != nil {
			return nil, err
		}
		c.Set("request_model", request.Model)
		return geminiRequest, nil
	} else if a.RequestMode == RequestModeLlama {
		return request, nil
	}
	return nil, errors.New("unsupported request mode")
}

func (a *Adaptor) ConvertRerankRequest(c *gin.Context, relayMode int, request dto.RerankRequest) (any, error) {
	return nil, nil
}

func (a *Adaptor) ConvertEmbeddingRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.EmbeddingRequest) (any, error) {
	//TODO implement me
	return nil, errors.New("not implemented")
}

func (a *Adaptor) ConvertOpenAIResponsesRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.OpenAIResponsesRequest) (any, error) {
	// TODO implement me
	return nil, errors.New("not implemented")
}

func (a *Adaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (any, error) {
	return channel.DoApiRequest(a, c, info, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage any, err *types.NewAPIError) {
	// 检查是否需要将Gemini响应转换为Claude格式
	forceGeminiMode := c.GetBool("force_gemini_mode")
	
	if info.IsStream {
		switch a.RequestMode {
		case RequestModeClaude:
			err, usage = claude.ClaudeStreamHandler(c, resp, info, claude.RequestModeMessage)
		case RequestModeGemini:
			if forceGeminiMode {
				// 当使用Claude Code工具时，需要特殊处理Gemini流式响应
				usage, err = handleGeminiStreamWithClaudeFormat(c, info, resp)
			} else if info.RelayMode == constant.RelayModeGemini {
				usage, err = gemini.GeminiTextGenerationStreamHandler(c, info, resp)
			} else {
				usage, err = gemini.GeminiChatStreamHandler(c, info, resp)
			}
		case RequestModeLlama:
			usage, err = openai.OaiStreamHandler(c, info, resp)
		}
	} else {
		switch a.RequestMode {
		case RequestModeClaude:
			err, usage = claude.ClaudeHandler(c, resp, claude.RequestModeMessage, info)
		case RequestModeGemini:
			if forceGeminiMode {
				// 当使用Claude Code工具时，需要将Gemini响应转换为Claude格式
				usage, err = handleGeminiResponseWithClaudeFormat(c, info, resp)
			} else if info.RelayMode == constant.RelayModeGemini {
				usage, err = gemini.GeminiTextGenerationHandler(c, info, resp)
			} else {
				usage, err = gemini.GeminiChatHandler(c, info, resp)
			}
		case RequestModeLlama:
			usage, err = openai.OpenaiHandler(c, info, resp)
		}
	}
	return
}

func (a *Adaptor) GetModelList() []string {
	var modelList []string
	for i, s := range ModelList {
		modelList = append(modelList, s)
		ModelList[i] = s
	}
	for i, s := range claude.ModelList {
		modelList = append(modelList, s)
		claude.ModelList[i] = s
	}
	for i, s := range gemini.ModelList {
		modelList = append(modelList, s)
		gemini.ModelList[i] = s
	}
	return modelList
}

func (a *Adaptor) GetChannelName() string {
	return ChannelName
}

// transformWebSearchTools converts WebSearch tools to supported Vertex AI tools
func transformWebSearchTools(tools any) (any, error) {
	if tools == nil {
		return nil, nil
	}

	toolsList, ok := tools.([]any)
	if !ok {
		return tools, nil
	}

	transformedTools := make([]any, 0, len(toolsList))
	
	for _, tool := range toolsList {
		if toolMap, ok := tool.(map[string]any); ok {
			toolType, hasType := toolMap["type"].(string)
			
			// Check if this is a WebSearch tool that needs transformation
			if hasType && (toolType == "web_search_20250305" || toolType == "WebSearch" || toolType == "websearch") {
				// Add bash tool for web requests (name must be "bash" for bash_20250124 type)
				bashTool := map[string]any{
					"type": "bash_20250124",
					"name": "bash",
				}
				transformedTools = append(transformedTools, bashTool)
				
				// Add custom tool to provide web search interface
				customTool := map[string]any{
					"type": "custom",
					"name": "web_search",
					"description": "Search the web for information. Use the bash tool to execute curl commands for web searching. Example: curl -s 'https://www.google.com/search?q=your+query' or curl -s 'https://duckduckgo.com/?q=your+query&format=json'",
					"input_schema": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"query": map[string]any{
								"type": "string",
								"description": "Search query to find information about",
							},
							"max_results": map[string]any{
								"type": "integer",
								"description": "Maximum number of results to return",
								"default": 5,
							},
						},
						"required": []string{"query"},
					},
				}
				transformedTools = append(transformedTools, customTool)
			} else {
				// Keep other tools unchanged
				transformedTools = append(transformedTools, tool)
			}
		} else {
			// Keep non-map tools unchanged
			transformedTools = append(transformedTools, tool)
		}
	}
	
	return transformedTools, nil
}

// handleGeminiResponseWithClaudeFormat 处理Gemini非流式响应并转换为Claude格式
func handleGeminiResponseWithClaudeFormat(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Response) (any, *types.NewAPIError) {
	// 读取原始响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, types.NewError(err, types.ErrorCodeReadResponseBodyFailed)
	}
	defer resp.Body.Close()

	// 解析Gemini响应
	var geminiResponse map[string]interface{}
	if err := json.Unmarshal(body, &geminiResponse); err != nil {
		return nil, types.NewError(err, types.ErrorCodeJsonMarshalFailed)
	}

	// 转换为Claude格式
	claudeResponse, err := ConvertGeminiResponseToClaudeFormat(geminiResponse)
	if err != nil {
		return nil, types.NewError(err, types.ErrorCodeConvertRequestFailed)
	}

	// 设置正确的响应头
	c.Header("Content-Type", "application/json")
	
	// 写入转换后的响应
	if err := json.NewEncoder(c.Writer).Encode(claudeResponse); err != nil {
		return nil, types.NewError(err, types.ErrorCodeJsonMarshalFailed)
	}

	// 提取并转换usage信息为dto.Usage格式
	var usage *dto.Usage
	if usageData, ok := claudeResponse["usage"].(map[string]interface{}); ok {
		usage = &dto.Usage{}
		
		if inputTokens, ok := usageData["input_tokens"].(float64); ok {
			usage.PromptTokens = int(inputTokens)
			usage.InputTokens = int(inputTokens)
		}
		
		if outputTokens, ok := usageData["output_tokens"].(float64); ok {
			usage.CompletionTokens = int(outputTokens)
			usage.OutputTokens = int(outputTokens)
		}
		
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}

	return usage, nil
}

// handleGeminiStreamWithClaudeFormat 处理Gemini流式响应并转换为Claude格式
func handleGeminiStreamWithClaudeFormat(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Response) (any, *types.NewAPIError) {
	// 对于流式响应，我们暂时使用非流式处理方式
	// 在实际应用中，可能需要实现真正的流式转换
	return handleGeminiResponseWithClaudeFormat(c, info, resp)
}
