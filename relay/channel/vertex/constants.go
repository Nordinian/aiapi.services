package vertex

var ModelList = []string{
	//"claude-3-sonnet-20240229",
	//"claude-3-opus-20240229",
	//"claude-3-haiku-20240307",
	//"claude-3-5-sonnet-20240620",

	//"gemini-1.5-pro-latest", "gemini-1.5-flash-latest",
	//"gemini-1.5-pro-001", "gemini-1.5-flash-001", "gemini-pro", "gemini-pro-vision",

	"meta/llama3-405b-instruct-maas",
	
	// Veo 视频生成模型
	"veo-2-generate-001",
	"veo-3-generate-001",
	"veo-3-fast-generate-001",
	
	// Imagen 图像生成模型
	// Imagen 4 系列 (Preview 版本)
	"imagen-4.0-generate-preview-06-06",
	"imagen-4.0-fast-generate-preview-06-06",
	"imagen-4.0-ultra-generate-preview-06-06",
	
	// Imagen 3 系列 (已验证可用)
	"imagen-3.0-generate-002",
	// "imagen-3.0-generate-001", // 已弃用，由 imagen-3.0-generate-002 替代
	"imagen-3.0-fast-generate-001",
	"imagen-3.0-capability-001",
	
	// DeepSeek 推理模型
	"deepseek-ai/deepseek-r1-0528-maas",
	
	// Lyria 音频生成模型
	"lyria-music-generate-001",
	"lyria-audio-generate-001",
	"lyria-voice-clone-001",
	"lyria-sound-effects-001",
	
	// Embedding 文本嵌入模型
	"text-embedding-004",
	"text-multilingual-embedding-002",
	"textembedding-gecko",
	"textembedding-gecko-multilingual",
	"text-embedding-preview-0815",
	
	// Text-to-Speech 语音合成模型
	"text-to-speech-001",
	"text-to-speech-multilingual",
	"text-to-speech-neural",
	"text-to-speech-standard",
}

var ChannelName = "vertex-ai"
