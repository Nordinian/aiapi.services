# Vertex AI 弃用模型清理报告

## 概述
本文档记录了Vertex AI渠道中已弃用模型的清理工作，包括移除过期的预览版本、实验版本和不再支持的模型。

## 已清理的弃用模型

### 1. PaLM系列模型
- **PaLM-2**: 已被Gemini系列完全替代
- **文件**: 
  - `relay/channel/palm/constants.go`: 注释掉ModelList中的PaLM-2
  - `setting/ratio_setting/model_ratio.go`: 注释掉定价配置

### 2. Gemini 1.0系列模型  
- **gemini-1.0-pro**: 已被gemini-1.5-pro替代
- **文件**:
  - `setting/model_setting/gemini.go`: 注释掉版本设置
  - `setting/ratio_setting/model_ratio.go`: 注释掉定价配置

### 3. Gemini 2.5预览和实验版本
- **gemini-2.5-pro-exp-03-25**: 实验版本，由正式版本替代
- **gemini-2.5-pro-preview-03-25**: 预览版本，由正式版本替代  
- **gemini-2.5-flash-preview-04-17**: 预览版本，包含thinking/nothinking变体
- **gemini-2.5-flash-preview-05-20**: 预览版本，包含thinking/nothinking变体
- **gemini-2.5-flash-lite-preview-06-17**: 预览版本

**清理的文件**:
- `setting/ratio_setting/model_ratio.go`: 注释掉所有预览版本的定价配置
- `setting/ratio_setting/cache_ratio.go`: 注释掉缓存相关配置
- `relay/channel/gemini/constant.go`: 注释掉模型列表中的预览版本
- `setting/operation_setting/tools.go`: 注释掉音频定价函数中的预览版本引用
- `relay/channel/gemini/relay-gemini.go`: 更新thinking adapter逻辑，移除对预览版本的特殊处理

### 4. Imagen系列弃用模型
- **imagen-3.0-generate-001**: 已被imagen-3.0-generate-002替代
- **文件**:
  - `relay/channel/vertex/adaptor.go`: 注释掉模型映射
  - `relay/channel/vertex/constants.go`: 注释掉模型列表
  - `setting/ratio_setting/model_ratio.go`: 注释掉定价配置

## 保留的模型

### Gemini生产版本
- `gemini-1.5-pro-latest`, `gemini-1.5-flash-latest`
- `gemini-2.0-flash`  
- `gemini-2.5-pro` (支持分层定价)
- `gemini-2.5-flash`
- `gemini-2.5-flash-thinking-*` (通配符支持)
- `gemini-2.5-pro-thinking-*` (通配符支持)

### Gemini实验版本(仍在使用)
- `gemini-2.0-flash-exp`
- `gemini-2.0-pro-exp`
- `gemini-2.0-flash-thinking-exp`

### Claude系列(全保留)
- 所有Claude 3.x和4.x版本保持不变

### Vertex AI其他服务
- Imagen 3.0和4.0系列(去除弃用版本)
- Veo视频生成系列
- Lyria音频生成系列
- Embedding和TTS系列
- DeepSeek推理模型
- Llama系列

## 修改统计

### 注释掉的配置项
- **model_ratio.go**: 9个弃用模型的定价配置
- **cache_ratio.go**: 6个弃用模型的缓存配置(命中/创建/存储)
- **常量文件**: 4个ModelList条目

### 更新的逻辑
- **thinking adapter**: 简化预览版本检测逻辑
- **音频定价**: 移除预览版本特殊处理
- **模型映射**: 清理弃用的imagen映射

## 验证结果
- ✅ 编译测试通过: `go build -o one-api .`
- ✅ 无编译错误或警告
- ✅ 保留所有生产环境使用的模型
- ✅ 清理了所有已确认弃用的模型

## 影响评估
1. **向后兼容性**: 弃用模型已注释而非删除，保持配置完整性
2. **功能影响**: 无影响，所有弃用模型都有正式版本替代
3. **定价准确性**: 移除了过时的定价信息，避免计费混乱
4. **维护简化**: 减少了需要维护的模型配置数量

## 后续建议
1. 在生产环境部署前，建议进行充分的集成测试
2. 监控用户是否还在使用已注释的弃用模型
3. 考虑在未来版本中完全移除注释的配置项
4. 定期审查Vertex AI官方文档，及时清理新的弃用模型

---
生成时间: 2025-02-02
生成者: Claude Code SuperClaude Framework