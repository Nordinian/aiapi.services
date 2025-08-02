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
	"time"

	"github.com/gin-gonic/gin"
)

const (
	RequestModeClaude    = 1
	RequestModeGemini    = 2
	RequestModeLlama     = 3
	RequestModeVeo       = 4  // 视频生成
	RequestModeImagen    = 5  // 图像生成
	RequestModeDeepSeek  = 6  // DeepSeek推理模型
	RequestModeLyria     = 7  // 音频生成
	RequestModeEmbedding = 8  // 文本嵌入
	RequestModeTTS       = 9  // 语音合成
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

// Veo 模型映射
var veoModelMap = map[string]string{
	"veo-2-generate-001":      "veo-2-generate-001",
	"veo-3-generate-001":      "veo-3-generate-001",
	"veo-3-fast-generate-001": "veo-3-fast-generate-001",
}

// Imagen 模型映射
var imagenModelMap = map[string]string{
	// Imagen 4 系列 (Preview 版本)
	"imagen-4.0-generate-preview-06-06":      "imagen-4.0-generate-preview-06-06",      // Imagen 4 标准版
	"imagen-4.0-fast-generate-preview-06-06": "imagen-4.0-fast-generate-preview-06-06", // Imagen 4 快速版
	"imagen-4.0-ultra-generate-preview-06-06": "imagen-4.0-ultra-generate-preview-06-06", // Imagen 4 超高质量版
	
	// Imagen 3 系列 (已验证可用)
	"imagen-3.0-generate-002":      "imagen-3.0-generate-002",
	"imagen-3.0-generate-001":      "imagen-3.0-generate-001",
	"imagen-3.0-fast-generate-001": "imagen-3.0-fast-generate-001",
	"imagen-3.0-capability-001":    "imagen-3.0-capability-001",
}

// DeepSeek 推理模型映射 - 支持思维链推理(Chain-of-Thought)
// 注意：仅在 us-central1 和 us-east1 地区可用
var deepseekModelMap = map[string]string{
	"deepseek-ai/deepseek-r1-0528-maas": "deepseek-ai/deepseek-r1-0528-maas", // DeepSeek-R1推理模型
}

// Lyria 音频生成模型映射
var lyriaModelMap = map[string]string{
	"lyria-music-generate-001":    "lyria-music-generate-001",    // 音乐生成
	"lyria-audio-generate-001":    "lyria-audio-generate-001",    // 音频生成
	"lyria-voice-clone-001":       "lyria-voice-clone-001",       // 语音克隆
	"lyria-sound-effects-001":     "lyria-sound-effects-001",     // 音效生成
}

// Embedding 文本嵌入模型映射
var embeddingModelMap = map[string]string{
	"text-embedding-004":               "text-embedding-004",               // 最新嵌入模型
	"text-multilingual-embedding-002":  "text-multilingual-embedding-002",  // 多语言嵌入
	"textembedding-gecko":              "textembedding-gecko@001",           // Gecko嵌入
	"textembedding-gecko-multilingual": "textembedding-gecko-multilingual@001", // Gecko多语言
	"text-embedding-preview-0815":      "text-embedding-preview-0815",      // 预览版本
}

// Text-to-Speech 语音合成模型映射
var ttsModelMap = map[string]string{
	"text-to-speech-001":           "text-to-speech-001",           // 标准TTS
	"text-to-speech-multilingual":  "text-to-speech-multilingual",  // 多语言TTS
	"text-to-speech-neural":        "text-to-speech-neural",        // 神经网络TTS
	"text-to-speech-standard":      "text-to-speech-standard",      // 标准质量TTS
}

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
	var convertedRequest any
	var err error
	
	switch a.RequestMode {
	case RequestModeLyria:
		convertedRequest, err = a.ConvertLyriaRequest(c, info, request)
	case RequestModeTTS:
		convertedRequest, err = a.ConvertTTSRequest(c, info, request)
	default:
		return nil, errors.New("unsupported audio request mode")
	}
	
	if err != nil {
		return nil, err
	}
	
	// 将转换后的请求序列化为 JSON
	jsonBytes, err := json.Marshal(convertedRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal audio request: %w", err)
	}
	
	return strings.NewReader(string(jsonBytes)), nil
}

// ConvertVeoRequest 将 OpenAI 格式请求转换为 Veo 格式
func (a *Adaptor) ConvertVeoRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.GeneralOpenAIRequest) (*VeoRequest, error) {
	if a.RequestMode != RequestModeVeo {
		return nil, errors.New("not veo mode")
	}
	
	veoRequest := &VeoRequest{
		Contents: []VeoContent{},
		GenerationConfig: VeoGenerationConfig{
			Temperature:     0.8,
			MaxOutputTokens: 1024,
		},
	}

	// 处理消息内容
	if len(request.Messages) > 0 {
		lastMessage := request.Messages[len(request.Messages)-1]
		content := VeoContent{Parts: []VeoPart{}}

		// 处理文本内容
		if textContent := extractTextContentFromMessage(lastMessage); textContent != "" {
			content.Parts = append(content.Parts, VeoPart{Text: textContent})
		}

		// 处理图片内容（用于图片到视频）
		if imageData := extractImageContentFromMessage(lastMessage); imageData != nil {
			content.Parts = append(content.Parts, VeoPart{
				InlineData: &VeoInlineData{
					MimeType: imageData.MimeType,
					Data:     imageData.Data,
				},
			})
		}

		veoRequest.Contents = append(veoRequest.Contents, content)
	}

	return veoRequest, nil
}

// 提取消息中的文本内容
func extractTextContentFromMessage(message dto.Message) string {
	if content, ok := message.Content.(string); ok {
		return content
	}
	return ""
}

// 提取消息中的图片内容
func extractImageContentFromMessage(message dto.Message) *VeoInlineData {
	// 这里需要根据实际的消息结构来实现
	// 目前返回 nil，后续可以扩展
	return nil
}

// Imagen API 相关结构体
type ImagenRequest struct {
	Instances  []ImagenInstance  `json:"instances"`
	Parameters *ImagenParameters `json:"parameters,omitempty"`
}

type ImagenInstance struct {
	Prompt string `json:"prompt"`
}

type ImagenParameters struct {
	SampleCount       int     `json:"sampleCount,omitempty"`        // 生成数量
	AspectRatio       string  `json:"aspectRatio,omitempty"`        // 宽高比
	SafetyFilterLevel string  `json:"safetyFilterLevel,omitempty"`  // 安全过滤
	PersonGeneration  string  `json:"personGeneration,omitempty"`   // 人物生成
	NegativePrompt    string  `json:"negativePrompt,omitempty"`     // 负面提示
	Seed             int64   `json:"seed,omitempty"`               // 随机种子
	GuidanceScale    float64 `json:"guidanceScale,omitempty"`      // 引导强度
}

type ImagenResponse struct {
	Predictions []ImagenPrediction `json:"predictions"`
}

type ImagenPrediction struct {
	BytesBase64Encoded string                 `json:"bytesBase64Encoded"`
	MimeType          string                 `json:"mimeType"`
	SafetyAttributes  *ImagenSafetyAttributes `json:"safetyAttributes,omitempty"`
}

type ImagenSafetyAttributes struct {
	Blocked bool `json:"blocked"`
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.ImageRequest) (any, error) {
	if a.RequestMode != RequestModeImagen {
		return nil, errors.New("not imagen mode")
	}
	
	instances := []ImagenInstance{
		{Prompt: request.Prompt},
	}

	parameters := &ImagenParameters{
		SampleCount:       1,
		AspectRatio:       "1:1",
		SafetyFilterLevel: "block_some",
		PersonGeneration:  "allow_adult",
		GuidanceScale:     7.5,
	}

	// 处理生成数量
	if request.N > 0 {
		parameters.SampleCount = request.N
	}

	// 处理尺寸转换
	if request.Size != "" {
		parameters.AspectRatio = convertSizeToAspectRatio(request.Size)
	}

	return &ImagenRequest{
		Instances:  instances,
		Parameters: parameters,
	}, nil
}

// ConvertLyriaRequest 将 OpenAI 音频请求转换为 Lyria 格式
func (a *Adaptor) ConvertLyriaRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.AudioRequest) (*LyriaRequest, error) {
	if a.RequestMode != RequestModeLyria {
		return nil, errors.New("not lyria mode")
	}
	
	instance := LyriaInstance{
		Prompt:   request.Input,
		Duration: 30, // 默认30秒
	}

	// 根据模型类型设置参数
	switch {
	case strings.Contains(info.UpstreamModelName, "music"):
		instance.Style = "pop"
		instance.Tempo = "medium"
	case strings.Contains(info.UpstreamModelName, "voice"):
		if request.Voice != "" {
			instance.VoiceClone = request.Voice
		}
	case strings.Contains(info.UpstreamModelName, "sound"):
		instance.Duration = 10 // 音效较短
	}

	parameters := &LyriaParameters{
		Temperature:      0.7,
		MaxDurationSec:  60,
		AudioFormat:     "mp3",
		SampleRate:      44100,
		ReturnAudioData: true,
	}

	return &LyriaRequest{
		Instances:  []LyriaInstance{instance},
		Parameters: parameters,
	}, nil
}

// ConvertTTSRequest 将 OpenAI 音频请求转换为 TTS 格式
func (a *Adaptor) ConvertTTSRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.AudioRequest) (*TTSRequest, error) {
	if a.RequestMode != RequestModeTTS {
		return nil, errors.New("not tts mode")
	}
	
	// 保存原始文本到上下文中，用于后续计费
	c.Set("tts_original_text", request.Input)
	
	voice := TTSVoiceConfig{
		LanguageCode: "en-US", // 默认英语
		SsmlGender:   "NEUTRAL",
	}

	// 处理语音设置
	if request.Voice != "" {
		voice.Name = a.mapOpenAIVoiceToVertex(request.Voice)
		// 根据语音ID推断语言
		if lang := inferLanguageFromVoice(request.Voice); lang != "" {
			voice.LanguageCode = lang
		}
	}

	audioConfig := TTSAudioConfig{
		AudioEncoding:   "MP3",
		SampleRateHertz: 24000,
		SpeakingRate:    1.0,
		Pitch:          0.0,
		VolumeGainDb:    0.0,
	}

	// 处理音频格式
	if request.ResponseFormat != "" {
		audioConfig.AudioEncoding = strings.ToUpper(request.ResponseFormat)
	}

	// 处理语速设置
	if request.Speed > 0 {
		audioConfig.SpeakingRate = request.Speed
	}

	return &TTSRequest{
		Input:       request.Input,
		Voice:       voice,
		AudioConfig: audioConfig,
	}, nil
}

// inferLanguageFromVoice 根据语音ID推断语言
func inferLanguageFromVoice(voice string) string {
	// 根据语音ID推断语言
	voiceLanguageMap := map[string]string{
		"alloy":    "en-US",
		"echo":     "en-US", 
		"fable":    "en-US",
		"onyx":     "en-US",
		"nova":     "en-US",
		"shimmer":  "en-US",
	}
	
	if lang, exists := voiceLanguageMap[voice]; exists {
		return lang
	}
	
	// 从语音ID中提取语言前缀
	if len(voice) >= 5 && voice[2] == '-' {
		return voice[:5] // 例如 "en-US" from "en-US-Neural2-A"
	}
	
	return "en-US" // 默认
}

// mapOpenAIVoiceToVertex 将 OpenAI 语音名称映射到 Vertex AI 格式
func (a *Adaptor) mapOpenAIVoiceToVertex(openaiVoice string) string {
	voiceMap := map[string]string{
		"alloy":    "en-US-Neural2-A",
		"echo":     "en-US-Neural2-B", 
		"fable":    "en-US-Neural2-C",
		"onyx":     "en-US-Neural2-D",
		"nova":     "en-US-Neural2-E",
		"shimmer":  "en-US-Neural2-F",
	}
	
	if vertexVoice, exists := voiceMap[openaiVoice]; exists {
		return vertexVoice
	}
	return openaiVoice // 直接返回，假设是Vertex AI格式
}

func convertSizeToAspectRatio(size string) string {
	sizeMap := map[string]string{
		"256x256":   "1:1",
		"512x512":   "1:1",
		"1024x1024": "1:1",
		"1024x768":  "4:3",
		"768x1024":  "3:4",
		"1536x1024": "3:2",
		"1024x1536": "2:3",
		"1792x1024": "16:9",
		"1024x1792": "9:16",
	}
	
	if ratio, exists := sizeMap[size]; exists {
		return ratio
	}
	return "1:1"
}

// Veo API 相关结构体
type VeoRequest struct {
	Contents []VeoContent `json:"contents"`
	GenerationConfig VeoGenerationConfig `json:"generationConfig,omitempty"`
}

type VeoContent struct {
	Parts []VeoPart `json:"parts"`
}

type VeoPart struct {
	Text       string          `json:"text,omitempty"`
	InlineData *VeoInlineData  `json:"inlineData,omitempty"`
}

type VeoInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type VeoGenerationConfig struct {
	Temperature      float64 `json:"temperature,omitempty"`
	MaxOutputTokens  int     `json:"maxOutputTokens,omitempty"`
	TopP            float64 `json:"topP,omitempty"`
	TopK            int     `json:"topK,omitempty"`
}

type VeoResponse struct {
	Candidates []VeoCandidate `json:"candidates"`
}

type VeoCandidate struct {
	Content VeoContentResponse `json:"content"`
}

type VeoContentResponse struct {
	Parts []VeoPartResponse `json:"parts"`
}

type VeoPartResponse struct {
	VideoMetadata *VeoVideoMetadata `json:"videoMetadata,omitempty"`
	InlineData    *VeoInlineData    `json:"inlineData,omitempty"`
}

type VeoVideoMetadata struct {
	GeneratedVideoUri string `json:"generatedVideoUri"`
	DurationMs       int64  `json:"durationMs"`
}

func (a *Adaptor) Init(info *relaycommon.RelayInfo) {
	modelName := info.UpstreamModelName
	
	switch {
	case strings.HasPrefix(modelName, "claude"):
		a.RequestMode = RequestModeClaude
	case strings.HasPrefix(modelName, "gemini"):
		a.RequestMode = RequestModeGemini
	case strings.HasPrefix(modelName, "veo"):
		a.RequestMode = RequestModeVeo
	case strings.HasPrefix(modelName, "imagen"):
		a.RequestMode = RequestModeImagen
	case strings.HasPrefix(modelName, "deepseek-ai/"):
		a.RequestMode = RequestModeDeepSeek
	case strings.HasPrefix(modelName, "lyria"):
		a.RequestMode = RequestModeLyria
	case strings.HasPrefix(modelName, "text-embedding") || 
		 strings.HasPrefix(modelName, "textembedding"):
		a.RequestMode = RequestModeEmbedding
	case strings.HasPrefix(modelName, "text-to-speech") ||
		 strings.HasPrefix(modelName, "tts-"):
		a.RequestMode = RequestModeTTS
	case strings.Contains(modelName, "llama"):
		a.RequestMode = RequestModeLlama
	default:
		a.RequestMode = RequestModeGemini // 默认
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
	} else if a.RequestMode == RequestModeVeo {
		// Veo API URL
		model := info.UpstreamModelName
		if v, ok := veoModelMap[info.UpstreamModelName]; ok {
			model = v
		}
		if region == "global" {
			return fmt.Sprintf(
				"https://aiplatform.googleapis.com/v1/projects/%s/locations/global/publishers/google/models/%s:generateContent",
				adc.ProjectID, model), nil
		} else {
			return fmt.Sprintf(
				"https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:generateContent",
				region, adc.ProjectID, region, model), nil
		}
	} else if a.RequestMode == RequestModeImagen {
		// Imagen API URL
		model := info.UpstreamModelName
		if v, ok := imagenModelMap[info.UpstreamModelName]; ok {
			model = v
		}
		if region == "global" {
			return fmt.Sprintf(
				"https://aiplatform.googleapis.com/v1/projects/%s/locations/global/publishers/google/models/%s:predict",
				adc.ProjectID, model), nil
		} else {
			return fmt.Sprintf(
				"https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:predict",
				region, adc.ProjectID, region, model), nil
		}
	} else if a.RequestMode == RequestModeDeepSeek {
		// DeepSeek API URL - 使用OpenAI兼容端点
		if region == "global" {
			return fmt.Sprintf(
				"https://aiplatform.googleapis.com/v1/projects/%s/locations/global/endpoints/openapi/chat/completions",
				adc.ProjectID), nil
		} else {
			return fmt.Sprintf(
				"https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/endpoints/openapi/chat/completions",
				region, adc.ProjectID, region), nil
		}
	} else if a.RequestMode == RequestModeLyria {
		// Lyria API URL
		model := info.UpstreamModelName
		if v, ok := lyriaModelMap[info.UpstreamModelName]; ok {
			model = v
		}
		if region == "global" {
			return fmt.Sprintf(
				"https://aiplatform.googleapis.com/v1/projects/%s/locations/global/publishers/google/models/%s:predict",
				adc.ProjectID, model), nil
		} else {
			return fmt.Sprintf(
				"https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:predict",
				region, adc.ProjectID, region, model), nil
		}
	} else if a.RequestMode == RequestModeEmbedding {
		// Embedding API URL
		model := info.UpstreamModelName
		if v, ok := embeddingModelMap[info.UpstreamModelName]; ok {
			model = v
		}
		if region == "global" {
			return fmt.Sprintf(
				"https://aiplatform.googleapis.com/v1/projects/%s/locations/global/publishers/google/models/%s:predict",
				adc.ProjectID, model), nil
		} else {
			return fmt.Sprintf(
				"https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:predict",
				region, adc.ProjectID, region, model), nil
		}
	} else if a.RequestMode == RequestModeTTS {
		// Text-to-Speech API URL - 使用专门的TTS服务
		if region == "global" {
			return fmt.Sprintf(
				"https://texttospeech.googleapis.com/v1/text:synthesize"), nil
		} else {
			return fmt.Sprintf(
				"https://%s-texttospeech.googleapis.com/v1/text:synthesize",
				region), nil
		}
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
	} else if a.RequestMode == RequestModeVeo {
		// 对于 Veo API，转换为 Veo 格式
		veoRequest, err := a.ConvertVeoRequest(c, info, request)
		if err != nil {
			return nil, err
		}
		c.Set("request_model", request.Model)
		return veoRequest, nil
	} else if a.RequestMode == RequestModeDeepSeek {
		// 对于 DeepSeek API，使用标准 OpenAI 格式，只需验证模型名称
		if v, ok := deepseekModelMap[request.Model]; ok {
			request.Model = v
		}
		c.Set("request_model", request.Model)
		return request, nil
	} else if a.RequestMode == RequestModeLyria {
		// Lyria API 音频生成请求
		c.Set("request_model", request.Model)
		return request, nil
	} else if a.RequestMode == RequestModeEmbedding {
		// Embedding API 文本嵌入请求
		c.Set("request_model", request.Model)
		return request, nil
	} else if a.RequestMode == RequestModeTTS {
		// Text-to-Speech API 语音合成请求
		c.Set("request_model", request.Model)
		return request, nil
	}
	return nil, errors.New("unsupported request mode")
}

func (a *Adaptor) ConvertRerankRequest(c *gin.Context, relayMode int, request dto.RerankRequest) (any, error) {
	return nil, nil
}

func (a *Adaptor) ConvertEmbeddingRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.EmbeddingRequest) (any, error) {
	if a.RequestMode != RequestModeEmbedding {
		return nil, errors.New("not embedding mode")
	}
	
	// 解析输入内容
	inputTexts := request.ParseInput()
	instances := make([]EmbeddingInstance, 0, len(inputTexts))
	
	for _, input := range inputTexts {
		instance := EmbeddingInstance{
			Content: input,
			Task:    determineEmbeddingTask(request.Model),
		}
		instances = append(instances, instance)
	}

	parameters := &EmbeddingParameters{
		AutoTruncate: true,
	}

	if request.Dimensions > 0 {
		parameters.OutputDimensionality = request.Dimensions
	}

	return &EmbeddingRequest{
		Instances:  instances,
		Parameters: parameters,
	}, nil
}

func determineEmbeddingTask(model string) string {
	taskMap := map[string]string{
		"text-embedding-004":               "RETRIEVAL_DOCUMENT",
		"text-multilingual-embedding-002":  "SEMANTIC_SIMILARITY", 
		"textembedding-gecko":              "RETRIEVAL_DOCUMENT",
		"textembedding-gecko-multilingual": "SEMANTIC_SIMILARITY",
	}
	
	if task, exists := taskMap[model]; exists {
		return task
	}
	return "RETRIEVAL_DOCUMENT"
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
		case RequestModeVeo:
			// Veo API 不支持流式响应，返回错误
			err = types.NewErrorWithStatusCode(errors.New("Veo API does not support streaming"), "unsupported_operation", http.StatusBadRequest)
		case RequestModeImagen:
			// Imagen API 不支持流式响应，返回错误
			err = types.NewErrorWithStatusCode(errors.New("Imagen API does not support streaming"), "unsupported_operation", http.StatusBadRequest)
		case RequestModeDeepSeek:
			// DeepSeek API 支持流式响应，使用 OpenAI 处理器
			usage, err = openai.OaiStreamHandler(c, info, resp)
		case RequestModeLyria:
			// Lyria API 不支持流式响应
			err = types.NewErrorWithStatusCode(errors.New("Lyria API does not support streaming"), "unsupported_operation", http.StatusBadRequest)
		case RequestModeEmbedding:
			// Embedding API 不支持流式响应
			err = types.NewErrorWithStatusCode(errors.New("Embedding API does not support streaming"), "unsupported_operation", http.StatusBadRequest)
		case RequestModeTTS:
			// Text-to-Speech API 不支持流式响应
			err = types.NewErrorWithStatusCode(errors.New("Text-to-Speech API does not support streaming"), "unsupported_operation", http.StatusBadRequest)
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
		case RequestModeVeo:
			usage, err = a.handleVeoResponse(c, resp, info)
		case RequestModeImagen:
			usage, err = a.handleImagenResponse(c, resp, info)
		case RequestModeDeepSeek:
			// DeepSeek API 使用标准 OpenAI 格式响应
			usage, err = openai.OpenaiHandler(c, info, resp)
		case RequestModeLyria:
			// Lyria API 音频生成响应处理
			usage, err = a.handleLyriaResponse(c, resp, info)
		case RequestModeEmbedding:
			// Embedding API 文本嵌入响应处理
			usage, err = a.handleEmbeddingResponse(c, resp, info)
		case RequestModeTTS:
			// Text-to-Speech API 语音合成响应处理
			usage, err = a.handleTTSResponse(c, resp, info)
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

// handleVeoResponse 处理 Veo API 响应
func (a *Adaptor) handleVeoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage any, err *types.NewAPIError) {
	// 读取响应体
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, types.NewErrorWithStatusCode(readErr, "internal_error", http.StatusInternalServerError)
	}
	
	// 解析 Veo 响应
	var veoResp VeoResponse
	if parseErr := json.Unmarshal(body, &veoResp); parseErr != nil {
		return nil, types.NewErrorWithStatusCode(parseErr, "internal_error", http.StatusInternalServerError)
	}
	
	// 转换为 OpenAI 格式响应
	openaiResp := a.convertVeoToOpenAIResponse(&veoResp, info)
	
	// 返回响应
	c.JSON(http.StatusOK, openaiResp)
	
	// 计算实际的 usage - Veo API 按视频数量计费
	videoCount := len(veoResp.Candidates)
	if videoCount == 0 {
		videoCount = 1 // 默认至少生成一个视频
	}
	usage = dto.Usage{
		PromptTokens:     videoCount, // 将视频数量记录为 PromptTokens
		CompletionTokens: 0,          // 视频生成不产生completion tokens
		TotalTokens:      videoCount,
	}
	
	return usage, nil
}

// handleImagenResponse 处理 Imagen API 响应
func (a *Adaptor) handleImagenResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage any, err *types.NewAPIError) {
	// 读取响应体
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, types.NewErrorWithStatusCode(readErr, "internal_error", http.StatusInternalServerError)
	}
	
	// 解析 Imagen 响应
	var imagenResp ImagenResponse
	if parseErr := json.Unmarshal(body, &imagenResp); parseErr != nil {
		return nil, types.NewErrorWithStatusCode(parseErr, "internal_error", http.StatusInternalServerError)
	}
	
	// 转换为 OpenAI 格式响应
	openaiResp := a.convertImagenToOpenAIResponse(&imagenResp, info)
	
	// 返回响应
	c.JSON(http.StatusOK, openaiResp)
	
	// 计算实际的 usage - Imagen API 按图片数量计费
	imageCount := len(openaiResp.Data)
	usage = dto.Usage{
		PromptTokens:     imageCount, // 将图片数量记录为 PromptTokens
		CompletionTokens: 0,          // 图像生成不产生completion tokens
		TotalTokens:      imageCount,
	}
	
	return usage, nil
}

// convertVeoToOpenAIResponse 将 Veo 响应转换为 OpenAI 格式
func (a *Adaptor) convertVeoToOpenAIResponse(veoResp *VeoResponse, info *relaycommon.RelayInfo) *dto.OpenAITextResponse {
	// 这里需要根据 Veo 实际的响应格式来调整
	// 目前返回一个基本的响应结构
	return &dto.OpenAITextResponse{
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   info.OriginModelName,
		Choices: []dto.OpenAITextResponseChoice{
			{
				Index:        0,
				Message:      dto.Message{Role: "assistant", Content: "Video generation completed"},
				FinishReason: "stop",
			},
		},
	}
}

// convertImagenToOpenAIResponse 将 Imagen 响应转换为 OpenAI 格式
func (a *Adaptor) convertImagenToOpenAIResponse(imagenResp *ImagenResponse, info *relaycommon.RelayInfo) *dto.ImageResponse {
	imageResp := &dto.ImageResponse{
		Created: time.Now().Unix(),
		Data:    []dto.ImageData{},
	}

	for _, pred := range imagenResp.Predictions {
		// 检查安全过滤
		if pred.SafetyAttributes != nil && pred.SafetyAttributes.Blocked {
			continue
		}

		imageResp.Data = append(imageResp.Data, dto.ImageData{
			B64Json: pred.BytesBase64Encoded,
		})
	}

	return imageResp
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

// Lyria API 相关结构体
type LyriaRequest struct {
	Instances  []LyriaInstance  `json:"instances"`
	Parameters *LyriaParameters `json:"parameters,omitempty"`
}

type LyriaInstance struct {
	Prompt       string `json:"prompt"`                 // 文本描述
	Duration     int    `json:"duration,omitempty"`     // 持续时间(秒)
	Style        string `json:"style,omitempty"`        // 音乐风格
	Tempo        string `json:"tempo,omitempty"`        // 节拍
	Instrument   string `json:"instrument,omitempty"`   // 乐器
	VoiceClone   string `json:"voiceClone,omitempty"`   // 语音克隆数据
}

type LyriaParameters struct {
	Temperature       float64 `json:"temperature,omitempty"`       // 创造性
	MaxDurationSec   int     `json:"maxDurationSec,omitempty"`    // 最大时长
	AudioFormat      string  `json:"audioFormat,omitempty"`       // 音频格式
	SampleRate       int     `json:"sampleRate,omitempty"`        // 采样率
	ReturnAudioData  bool    `json:"returnAudioData,omitempty"`   // 返回音频数据
}

type LyriaResponse struct {
	Predictions []LyriaPrediction `json:"predictions"`
}

type LyriaPrediction struct {
	AudioContent string `json:"audioContent"` // base64编码的音频
	MimeType     string `json:"mimeType"`
	Duration     int    `json:"duration,omitempty"`
}

// Embedding API 相关结构体
type EmbeddingRequest struct {
	Instances []EmbeddingInstance `json:"instances"`
	Parameters *EmbeddingParameters `json:"parameters,omitempty"`
}

type EmbeddingInstance struct {
	Content   string `json:"content"`              // 文本内容
	Task      string `json:"task,omitempty"`       // 任务类型
	Title     string `json:"title,omitempty"`      // 文档标题
}

type EmbeddingParameters struct {
	AutoTruncate    bool   `json:"autoTruncate,omitempty"`    // 自动截断
	OutputDimensionality int `json:"outputDimensionality,omitempty"` // 输出维度
}

type EmbeddingResponse struct {
	Predictions []EmbeddingPrediction `json:"predictions"`
}

type EmbeddingPrediction struct {
	Embeddings    EmbeddingData `json:"embeddings"`
	TruncatedText bool          `json:"truncatedText,omitempty"`
}

type EmbeddingData struct {
	Values     []float64 `json:"values"`
	Statistics *EmbeddingStats `json:"statistics,omitempty"`
}

type EmbeddingStats struct {
	TokenCount int `json:"tokenCount,omitempty"`
}

// Text-to-Speech API 相关结构体
type TTSRequest struct {
	Input string           `json:"input"`    // 待合成文本
	Voice TTSVoiceConfig   `json:"voice"`    // 语音配置
	AudioConfig TTSAudioConfig `json:"audioConfig"` // 音频配置
}

type TTSVoiceConfig struct {
	LanguageCode string  `json:"languageCode"` // 语言代码
	Name         string  `json:"name,omitempty"` // 语音名称
	SsmlGender   string  `json:"ssmlGender,omitempty"` // 性别
}

type TTSAudioConfig struct {
	AudioEncoding    string  `json:"audioEncoding"`    // 音频编码格式
	SampleRateHertz  int     `json:"sampleRateHertz,omitempty"`  // 采样率
	SpeakingRate     float64 `json:"speakingRate,omitempty"`     // 语速 (0.25-4.0)
	Pitch           float64 `json:"pitch,omitempty"`            // 音调 (-20.0-20.0)
	VolumeGainDb    float64 `json:"volumeGainDb,omitempty"`     // 音量增益
	EffectsProfileId []string `json:"effectsProfileId,omitempty"` // 音效配置
}

type TTSResponse struct {
	AudioContent string `json:"audioContent"` // base64编码的音频
	AudioConfig  *TTSAudioConfig `json:"audioConfig,omitempty"`
}

// 自定义音频响应结构体，支持音频生成
type VertexAudioResponse struct {
	Created int64             `json:"created"`
	Data    []VertexAudioData `json:"data"`
}

type VertexAudioData struct {
	AudioBase64 string  `json:"audio_base64"`
	Format      string  `json:"format"`
	Duration    float64 `json:"duration,omitempty"`
	SampleRate  int     `json:"sample_rate,omitempty"`
}

// handleLyriaResponse 处理 Lyria API 响应
func (a *Adaptor) handleLyriaResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage any, err *types.NewAPIError) {
	// 读取响应体
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, types.NewErrorWithStatusCode(readErr, "internal_error", http.StatusInternalServerError)
	}
	
	// 解析 Lyria 响应
	var lyriaResp LyriaResponse
	if parseErr := json.Unmarshal(body, &lyriaResp); parseErr != nil {
		return nil, types.NewErrorWithStatusCode(parseErr, "internal_error", http.StatusInternalServerError)
	}
	
	// 转换为音频响应格式
	audioResp := a.convertLyriaToAudioResponse(&lyriaResp, info)
	
	// 返回响应
	c.JSON(http.StatusOK, audioResp)
	
	// 计算实际的 usage - Lyria API 按音频时长计费
	totalDuration := 0
	for _, data := range audioResp.Data {
		totalDuration += int(data.Duration)
	}
	if totalDuration == 0 {
		totalDuration = 30 // 默认30秒
	}
	usage = dto.Usage{
		PromptTokens:     totalDuration, // 将音频时长(秒)记录为 PromptTokens
		CompletionTokens: 0,             // 音频生成不产生completion tokens
		TotalTokens:      totalDuration,
	}
	
	return usage, nil
}

// handleEmbeddingResponse 处理 Embedding API 响应
func (a *Adaptor) handleEmbeddingResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage any, err *types.NewAPIError) {
	// 读取响应体
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, types.NewErrorWithStatusCode(readErr, "internal_error", http.StatusInternalServerError)
	}
	
	// 解析 Embedding 响应
	var embeddingResp EmbeddingResponse
	if parseErr := json.Unmarshal(body, &embeddingResp); parseErr != nil {
		return nil, types.NewErrorWithStatusCode(parseErr, "internal_error", http.StatusInternalServerError)
	}
	
	// 转换为 OpenAI 格式响应
	openaiResp := a.convertEmbeddingToOpenAIResponse(&embeddingResp, info)
	
	// 返回响应
	c.JSON(http.StatusOK, openaiResp)
	
	// 使用 convertEmbeddingToOpenAIResponse 中已计算的 usage
	usage = openaiResp.Usage
	
	return usage, nil
}

// handleTTSResponse 处理 Text-to-Speech API 响应
func (a *Adaptor) handleTTSResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage any, err *types.NewAPIError) {
	// 读取响应体
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, types.NewErrorWithStatusCode(readErr, "internal_error", http.StatusInternalServerError)
	}
	
	// 解析 TTS 响应
	var ttsResp TTSResponse
	if parseErr := json.Unmarshal(body, &ttsResp); parseErr != nil {
		return nil, types.NewErrorWithStatusCode(parseErr, "internal_error", http.StatusInternalServerError)
	}
	
	// 转换为音频响应格式
	audioResp := a.convertTTSToAudioResponse(&ttsResp, info)
	
	// 返回响应
	c.JSON(http.StatusOK, audioResp)
	
	// 计算实际的 usage - TTS API 按字符数计费
	// 从请求上下文中获取原始文本长度
	charCount := 0
	if originalText, exists := c.Get("tts_original_text"); exists {
		if text, ok := originalText.(string); ok {
			charCount = len(text)
		}
	}
	if charCount == 0 {
		charCount = 100 // 默认字符数
	}
	usage = dto.Usage{
		PromptTokens:     charCount, // 将字符数记录为 PromptTokens
		CompletionTokens: 0,         // TTS 不产生completion tokens
		TotalTokens:      charCount,
	}
	
	return usage, nil
}

// convertLyriaToAudioResponse 将 Lyria 响应转换为音频格式
func (a *Adaptor) convertLyriaToAudioResponse(lyriaResp *LyriaResponse, info *relaycommon.RelayInfo) *VertexAudioResponse {
	audioResp := &VertexAudioResponse{
		Created: time.Now().Unix(),
		Data:    []VertexAudioData{},
	}

	for _, pred := range lyriaResp.Predictions {
		audioData := VertexAudioData{
			AudioBase64: pred.AudioContent,
			Format:      pred.MimeType,
			Duration:    float64(pred.Duration),
		}
		audioResp.Data = append(audioResp.Data, audioData)
	}

	return audioResp
}

// convertEmbeddingToOpenAIResponse 将 Embedding 响应转换为 OpenAI 格式
func (a *Adaptor) convertEmbeddingToOpenAIResponse(embeddingResp *EmbeddingResponse, info *relaycommon.RelayInfo) *dto.EmbeddingResponse {
	openaiResp := &dto.EmbeddingResponse{
		Object: "list",
		Data:   []dto.EmbeddingResponseItem{},
		Model:  info.OriginModelName,
		Usage: dto.Usage{
			PromptTokens: 0,
			TotalTokens:  0,
		},
	}

	totalTokens := 0
	for i, pred := range embeddingResp.Predictions {
		embData := dto.EmbeddingResponseItem{
			Object:    "embedding",
			Index:     i,
			Embedding: pred.Embeddings.Values,
		}
		
		if pred.Embeddings.Statistics != nil {
			totalTokens += pred.Embeddings.Statistics.TokenCount
		}
		
		openaiResp.Data = append(openaiResp.Data, embData)
	}

	openaiResp.Usage.PromptTokens = totalTokens
	openaiResp.Usage.TotalTokens = totalTokens

	return openaiResp
}

// convertTTSToAudioResponse 将 TTS 响应转换为音频格式
func (a *Adaptor) convertTTSToAudioResponse(ttsResp *TTSResponse, info *relaycommon.RelayInfo) *VertexAudioResponse {
	audioResp := &VertexAudioResponse{
		Created: time.Now().Unix(),
		Data:    []VertexAudioData{},
	}

	audioData := VertexAudioData{
		AudioBase64: ttsResp.AudioContent,
		Format:      "mp3", // 默认格式
		SampleRate:  24000, // 默认采样率
	}

	if ttsResp.AudioConfig != nil {
		audioData.SampleRate = ttsResp.AudioConfig.SampleRateHertz
		// 根据编码格式设置format
		switch ttsResp.AudioConfig.AudioEncoding {
		case "MP3":
			audioData.Format = "mp3"
		case "WAV":
			audioData.Format = "wav"
		case "OGG":
			audioData.Format = "ogg"
		default:
			audioData.Format = "mp3"
		}
	}

	audioResp.Data = append(audioResp.Data, audioData)
	return audioResp
}
