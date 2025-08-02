package ratio_setting

import (
	"encoding/json"
	"one-api/common"
	"sync"
)

var defaultCacheRatio = map[string]float64{
	"gpt-4":                               0.5,
	"o1":                                  0.5,
	"o1-2024-12-17":                       0.5,
	"o1-preview-2024-09-12":               0.5,
	"o1-preview":                          0.5,
	"o1-mini-2024-09-12":                  0.5,
	"o1-mini":                             0.5,
	"o3-mini":                             0.5,
	"o3-mini-2025-01-31":                  0.5,
	"gpt-4o-2024-11-20":                   0.5,
	"gpt-4o-2024-08-06":                   0.5,
	"gpt-4o":                              0.5,
	"gpt-4o-mini-2024-07-18":              0.5,
	"gpt-4o-mini":                         0.5,
	"gpt-4o-realtime-preview":             0.5,
	"gpt-4o-mini-realtime-preview":        0.5,
	"gpt-4.5-preview":                     0.5,
	"gpt-4.5-preview-2025-02-27":          0.5,
	"deepseek-chat":                       0.25,
	"deepseek-reasoner":                   0.25,
	"deepseek-coder":                      0.25,
	"claude-3-sonnet-20240229":            0.1,
	"claude-3-opus-20240229":              0.1,
	"claude-3-haiku-20240307":             0.1,
	"claude-3-5-haiku-20241022":           0.1,
	"claude-3-5-sonnet-20240620":          0.1,
	"claude-3-5-sonnet-20241022":          0.1,
	"claude-3-7-sonnet-20250219":          0.1,
	"claude-3-7-sonnet-20250219-thinking": 0.1,
	"claude-sonnet-4-20250514":            0.1,
	"claude-sonnet-4-20250514-thinking":   0.1,
	"claude-opus-4-20250514":              0.1,
	"claude-opus-4-20250514-thinking":     0.1,
	
	// Vertex AI Gemini 系列缓存支持 (87.5% 折扣)
	"gemini-1.5-pro":                      0.125,  // $0.4375/1M vs $3.5/1M
	"gemini-1.5-pro-latest":               0.125,
	"gemini-1.5-flash":                    0.125,  // $0.046875/1M vs $0.375/1M
	"gemini-1.5-flash-latest":             0.125,
	"gemini-2.0-flash":                    0.125,
	"gemini-2.0-flash-exp":                0.125,
	"gemini-2.5-pro":                      0.125,  // 分层定价支持
	// "gemini-2.5-pro-exp-03-25":            0.125, // 已弃用，实验版本
	// "gemini-2.5-pro-preview-03-25":        0.125, // 已弃用，预览版本
	"gemini-2.5-flash":                    0.125,
	// "gemini-2.5-flash-preview-04-17":      0.125, // 已弃用，预览版本
	// "gemini-2.5-flash-preview-05-20":      0.125, // 已弃用，预览版本
	"gemini-2.5-flash-lite-preview-06-17": 0.125,
	
	// Vertex AI 嵌入模型缓存支持
	"text-embedding-004":                  0.125,  // $0.00005/1K vs $0.0004/1K
	"text-multilingual-embedding-002":     0.125,
	"textembedding-gecko":                 0.125,
	"textembedding-gecko-multilingual":    0.125,
}

var defaultCreateCacheRatio = map[string]float64{
	"claude-3-sonnet-20240229":            1.25,
	"claude-3-opus-20240229":              1.25,
	"claude-3-haiku-20240307":             1.25,
	"claude-3-5-haiku-20241022":           1.25,
	"claude-3-5-sonnet-20240620":          1.25,
	"claude-3-5-sonnet-20241022":          1.25,
	"claude-3-7-sonnet-20250219":          1.25,
	"claude-3-7-sonnet-20250219-thinking": 1.25,
	"claude-sonnet-4-20250514":            1.25,
	"claude-sonnet-4-20250514-thinking":   1.25,
	"claude-opus-4-20250514":              1.25,
	"claude-opus-4-20250514-thinking":     1.25,
	
	// Vertex AI Gemini 系列缓存创建 (标准价格，无额外费用)
	"gemini-1.5-pro":                      1.0,  // 缓存写入 = 标准价格
	"gemini-1.5-pro-latest":               1.0,
	"gemini-1.5-flash":                    1.0,
	"gemini-1.5-flash-latest":             1.0,
	"gemini-2.0-flash":                    1.0,
	"gemini-2.0-flash-exp":                1.0,
	"gemini-2.5-pro":                      1.0,  // 分层定价支持
	// "gemini-2.5-pro-exp-03-25":            1.0, // 已弃用，实验版本
	// "gemini-2.5-pro-preview-03-25":        1.0, // 已弃用，预览版本
	"gemini-2.5-flash":                    1.0,
	// "gemini-2.5-flash-preview-04-17":      1.0, // 已弃用，预览版本
	// "gemini-2.5-flash-preview-05-20":      1.0, // 已弃用，预览版本
	"gemini-2.5-flash-lite-preview-06-17": 1.0,
	
	// Vertex AI 嵌入模型缓存创建
	"text-embedding-004":                  1.0,
	"text-multilingual-embedding-002":     1.0,
	"textembedding-gecko":                 1.0,
	"textembedding-gecko-multilingual":    1.0,
}

// Vertex AI 缓存存储倍率 (按小时计费)
var defaultCacheStorageRatio = map[string]float64{
	// Vertex AI Gemini 系列缓存存储费用
	"gemini-1.5-pro":                      0.5,    // $1/1M tokens/hour
	"gemini-1.5-pro-latest":               0.5,
	"gemini-1.5-flash":                    0.05,   // $0.1/1M tokens/hour
	"gemini-1.5-flash-latest":             0.05,
	"gemini-2.0-flash":                    0.05,
	"gemini-2.0-flash-exp":                0.05,
	"gemini-2.5-pro":                      0.25,   // $0.5/1M tokens/hour
	// "gemini-2.5-pro-exp-03-25":            0.25, // 已弃用，实验版本
	// "gemini-2.5-pro-preview-03-25":        0.25, // 已弃用，预览版本
	"gemini-2.5-flash":                    0.05,
	// "gemini-2.5-flash-preview-04-17":      0.05, // 已弃用，预览版本
	// "gemini-2.5-flash-preview-05-20":      0.05, // 已弃用，预览版本
	"gemini-2.5-flash-lite-preview-06-17": 0.025,  // $0.05/1M tokens/hour
	
	// Vertex AI 嵌入模型缓存存储
	"text-embedding-004":                  0.025,  // $0.05/1M tokens/hour
	"text-multilingual-embedding-002":     0.025,
	"textembedding-gecko":                 0.025,
	"textembedding-gecko-multilingual":    0.025,
}

//var defaultCreateCacheRatio = map[string]float64{}

var cacheRatioMap map[string]float64
var cacheRatioMapMutex sync.RWMutex

// GetCacheRatioMap returns the cache ratio map
func GetCacheRatioMap() map[string]float64 {
	cacheRatioMapMutex.RLock()
	defer cacheRatioMapMutex.RUnlock()
	return cacheRatioMap
}

// CacheRatio2JSONString converts the cache ratio map to a JSON string
func CacheRatio2JSONString() string {
	cacheRatioMapMutex.RLock()
	defer cacheRatioMapMutex.RUnlock()
	jsonBytes, err := json.Marshal(cacheRatioMap)
	if err != nil {
		common.SysError("error marshalling cache ratio: " + err.Error())
	}
	return string(jsonBytes)
}

// UpdateCacheRatioByJSONString updates the cache ratio map from a JSON string
func UpdateCacheRatioByJSONString(jsonStr string) error {
	cacheRatioMapMutex.Lock()
	defer cacheRatioMapMutex.Unlock()
	cacheRatioMap = make(map[string]float64)
	err := json.Unmarshal([]byte(jsonStr), &cacheRatioMap)
	if err == nil {
		InvalidateExposedDataCache()
	}
	return err
}

// GetCacheRatio returns the cache ratio for a model
func GetCacheRatio(name string) (float64, bool) {
	cacheRatioMapMutex.RLock()
	defer cacheRatioMapMutex.RUnlock()
	ratio, ok := cacheRatioMap[name]
	if !ok {
		return 1, false // Default to 1 if not found
	}
	return ratio, true
}

func GetCreateCacheRatio(name string) (float64, bool) {
	ratio, ok := defaultCreateCacheRatio[name]
	if !ok {
		return 1.25, false // Default to 1.25 if not found
	}
	return ratio, true
}

func GetCacheRatioCopy() map[string]float64 {
	cacheRatioMapMutex.RLock()
	defer cacheRatioMapMutex.RUnlock()
	copyMap := make(map[string]float64, len(cacheRatioMap))
	for k, v := range cacheRatioMap {
		copyMap[k] = v
	}
	return copyMap
}

// GetCacheStorageRatio 获取缓存存储倍率 (Vertex AI 按小时计费)
func GetCacheStorageRatio(name string) (float64, bool) {
	ratio, ok := defaultCacheStorageRatio[name]
	if !ok {
		return 0, false // 不支持缓存存储计费的模型返回 0
	}
	return ratio, true
}

// CalculateCacheStorageCost 计算缓存存储费用
// modelName: 模型名称
// cacheTokens: 缓存的token数量
// storageHours: 存储小时数
// returns: 存储费用 (以配额单位计算)
func CalculateCacheStorageCost(modelName string, cacheTokens int, storageHours float64) float64 {
	storageRatio, exists := GetCacheStorageRatio(modelName)
	if !exists || storageHours <= 0 || cacheTokens <= 0 {
		return 0
	}
	
	// 存储费用 = (缓存tokens / 1M) × 存储倍率 × 存储小时数 × 配额单位
	return float64(cacheTokens) * storageRatio * storageHours / 1000000 * 500
}
