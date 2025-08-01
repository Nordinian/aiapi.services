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
	"reflect"
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
	"claude-3-5-haiku-20241022":  "claude-3-5-haiku@20241022",
	"claude-3-7-sonnet-20250219": "claude-3-7-sonnet@20250219",
	"claude-sonnet-4-20250514":   "claude-sonnet-4@20250514",
	"claude-opus-4-20250514":     "claude-opus-4@20250514",
}

const anthropicVersion = "vertex-2023-10-16"

type Adaptor struct {
	RequestMode        int
	AccountCredentials Credentials
}

func (a *Adaptor) ConvertClaudeRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.ClaudeRequest) (any, error) {
	if v, ok := claudeModelMap[info.UpstreamModelName]; ok {
		c.Set("request_model", v)
	} else {
		c.Set("request_model", request.Model)
	}
	
	// Normalize tools to ensure compatibility with Vertex AI
	if request.Tools != nil {
		normalizedTools, err := normalizeToolsForVertexAI(request.Tools)
		if err != nil {
			return nil, fmt.Errorf("failed to normalize tools: %w", err)
		}
		request.Tools = normalizedTools
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
	if a.RequestMode == RequestModeClaude {
		claudeReq, err := claude.RequestOpenAI2ClaudeMessage(*request)
		if err != nil {
			return nil, err
		}
		
		// Normalize tools to ensure compatibility with Vertex AI
		if claudeReq.Tools != nil {
			fmt.Printf("[DEBUG] ConvertOpenAIRequest - Before normalization: %+v\n", claudeReq.Tools)
			normalizedTools, err := normalizeToolsForVertexAI(claudeReq.Tools)
			if err != nil {
				return nil, fmt.Errorf("failed to normalize tools: %w", err)
			}
			claudeReq.Tools = normalizedTools
			fmt.Printf("[DEBUG] ConvertOpenAIRequest - After normalization: %+v\n", claudeReq.Tools)
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
	if info.IsStream {
		switch a.RequestMode {
		case RequestModeClaude:
			err, usage = claude.ClaudeStreamHandler(c, resp, info, claude.RequestModeMessage)
		case RequestModeGemini:
			if info.RelayMode == constant.RelayModeGemini {
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
			if info.RelayMode == constant.RelayModeGemini {
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

// normalizeToolsForVertexAI converts Anthropic-specific tools to function tools compatible with Vertex AI
func normalizeToolsForVertexAI(tools any) (any, error) {
	if tools == nil {
		return nil, nil
	}

	toolsList, ok := tools.([]any)
	if !ok {
		return tools, nil
	}

	normalizedTools := make([]any, 0, len(toolsList))
	hasWebTools := false
	hasBashTool := false
	
	fmt.Printf("[DEBUG] normalizeToolsForVertexAI: Processing %d tools\n", len(toolsList))
	
	for i, tool := range toolsList {
		fmt.Printf("[DEBUG] Tool %d: type=%s\n", i, reflect.TypeOf(tool))
		
		// Handle map-based tools (e.g., from /v1/messages direct requests)
		if toolMap, ok := tool.(map[string]any); ok {
			toolType, hasType := toolMap["type"].(string)
			toolName, hasName := toolMap["name"].(string)
			
			fmt.Printf("[DEBUG] Tool %d (map): type=%s, name=%s, hasType=%v, hasName=%v\n", i, toolType, toolName, hasType, hasName)
			
			// Check for existing bash tool
			if hasType && toolType == "bash_20250124" {
				fmt.Printf("[DEBUG] Found existing bash tool\n")
				hasBashTool = true
				normalizedTools = append(normalizedTools, tool)
				continue
			}
			
			// Skip unsupported web_search_20250305 tool - will add bash instead
			if hasType && toolType == "web_search_20250305" {
				fmt.Printf("[DEBUG] Skipping unsupported web_search_20250305, will add bash tool\n")
				hasWebTools = true
				continue // Skip this tool
			}
			
			// Skip unsupported web_fetch_20250305 tool - will add bash instead
			if hasType && toolType == "web_fetch_20250305" {
				fmt.Printf("[DEBUG] Skipping unsupported web_fetch_20250305, will add bash tool\n")
				hasWebTools = true
				continue // Skip this tool
			}
			
			// Handle function tools (OpenAI format) - pass through as-is
			if hasType && toolType == "function" {
				fmt.Printf("[DEBUG] Keeping function tool: %s\n", getToolFunctionName(toolMap))
				
				// Check if this is a web-related function
				functionName := getToolFunctionName(toolMap)
				if functionName == "web_search" || functionName == "web_fetch" {
					hasWebTools = true
				}
				
				normalizedTools = append(normalizedTools, tool)
				continue
			}
			
			// Handle Claude Code native tools (without type field)
			if !hasType && hasName {
				if toolName == "WebFetch" || toolName == "WebSearch" {
					fmt.Printf("[DEBUG] Skipping Claude Code web tool: %s, will add bash tool\n", toolName)
					hasWebTools = true
					continue // Skip web tools, use bash instead
				}
			}
			
			// Handle any other tool formats - pass through with logging
			fmt.Printf("[DEBUG] Passing through unknown tool format: type=%s, name=%s\n", toolType, toolName)
			normalizedTools = append(normalizedTools, tool)
			continue
		}
		
		// Handle pointer to dto.Tool (from RequestOpenAI2ClaudeMessage)
		if toolPtr, ok := tool.(*dto.Tool); ok {
			fmt.Printf("[DEBUG] Tool %d (dto.Tool pointer): name=%s\n", i, toolPtr.Name)
			
			// Check if this is a web-related tool
			if toolPtr.Name == "web_search" || toolPtr.Name == "web_fetch" {
				fmt.Printf("[DEBUG] Skipping dto.Tool %s, will add bash tool\n", toolPtr.Name)
				hasWebTools = true
				continue // Skip web tools, use bash instead
			}
			
			// For non-web tools, convert to function tool format
			convertedTool := map[string]any{
				"type": "function",
				"function": map[string]any{
					"name":        toolPtr.Name,
					"description": toolPtr.Description,
					"parameters":  toolPtr.InputSchema,
				},
			}
			normalizedTools = append(normalizedTools, convertedTool)
			continue
		}
		
		// Handle pointer to dto.ClaudeWebSearchTool (from RequestOpenAI2ClaudeMessage WebSearchOptions)
		if webSearchTool, ok := tool.(*dto.ClaudeWebSearchTool); ok {
			fmt.Printf("[DEBUG] Tool %d (ClaudeWebSearchTool pointer): type=%s, name=%s\n", i, webSearchTool.Type, webSearchTool.Name)
			
			// Skip web search tool, will add bash instead
			fmt.Printf("[DEBUG] Skipping ClaudeWebSearchTool, will add bash tool\n")
			hasWebTools = true
			continue // Skip this tool
		}
		
		// Keep any other tool types unchanged
		fmt.Printf("[DEBUG] Keeping unknown tool type: %s\n", reflect.TypeOf(tool))
		normalizedTools = append(normalizedTools, tool)
	}
	
	// Add bash tool if we have web tools but no bash tool
	if hasWebTools && !hasBashTool {
		fmt.Printf("[DEBUG] Adding bash tool for web function support\n")
		bashTool := map[string]any{
			"type": "bash_20250124",
			"name": "bash",
		}
		normalizedTools = append(normalizedTools, bashTool)
	}
	
	fmt.Printf("[DEBUG] normalizeToolsForVertexAI: Original: %d, Final: %d tools\n", len(toolsList), len(normalizedTools))
	return normalizedTools, nil
}

// Helper function to extract function name from function tool
func getToolFunctionName(toolMap map[string]any) string {
	if function, ok := toolMap["function"].(map[string]any); ok {
		if name, ok := function["name"].(string); ok {
			return name
		}
	}
	return ""
}
