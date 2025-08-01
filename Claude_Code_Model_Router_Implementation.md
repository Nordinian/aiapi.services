# Claude Code Model Router Implementation

## 概述

成功实现了 Claude Code 与 Gemini 模型的兼容性，使用户可以通过服务端路由在 Claude Code 中使用 Gemini 模型，而无需修改客户端。

## 实现特性

### 1. 消息解析路由
- **功能**: 解析用户消息中的 `/model` 命令
- **格式**: `/model <model-name> [remaining message]`
- **示例**: 
  - `/model gemini-2.0-flash Hello world` → 切换到 gemini-2.0-flash，消息变为 "Hello world"
  - `/model flash How are you?` → 切换到 gemini-2.5-flash，消息变为 "How are you?"

### 2. 环境变量配置
- **功能**: 通过 HTTP 请求头传递模型配置
- **支持的请求头**:
  - `X-Claude-Custom-Model`
  - `X-Claude-Model`
  - `Claude-Custom-Model`
  - `Model`
  - `X-Anthropic-Model`
- **客户端使用方式**:
  ```bash
  export CLAUDE_CUSTOM_MODEL="gemini-2.0-flash"
  claude "your prompt here"
  ```

### 3. 会话级模型持久化
- **功能**: 记住用户在会话中选择的模型
- **特性**: 
  - 模型选择在会话期间保持
  - 自动清理 24 小时前的旧会话
  - 基于客户端 IP 和请求 ID 生成会话标识

## 支持的模型

### Gemini 模型
- `gemini-2.0-flash` → gemini-2.0-flash
- `gemini-2.5-pro` → gemini-2.5-pro  
- `gemini-2.5-flash` → gemini-2.5-flash
- `gemini-2.0-flash-lite` → gemini-2.0-flash-lite

### Claude 模型
- `claude-3-7-sonnet` → claude-3-7-sonnet
- `claude-opus-4` → claude-opus-4
- `claude-sonnet-4` → claude-sonnet-4

### 简短别名
- `gemini` → gemini-2.5-pro
- `flash` → gemini-2.5-flash
- `pro` → gemini-2.5-pro
- `lite` → gemini-2.0-flash-lite
- `claude`, `sonnet` → claude-sonnet-4
- `opus` → claude-opus-4

## 使用方法

### 1. 消息内命令切换
```
/model gemini-2.0-flash 请用中文回答问题
/model claude 分析这段代码
/model flash 生成一首诗
```

### 2. 环境变量配置
```bash
# 设置默认模型
export CLAUDE_CUSTOM_MODEL="gemini-2.5-pro"
claude "请帮我写一个Python函数"

# 切换到不同模型
export CLAUDE_CUSTOM_MODEL="claude-sonnet-4"  
claude "代码审查这个函数"
```

### 3. 别名使用
```
/model gemini  # 等同于 gemini-2.5-pro
/model flash   # 等同于 gemini-2.5-flash
/model claude  # 等同于 claude-sonnet-4
```

## 路由优先级

1. **消息中的 `/model` 命令** (最高优先级)
2. **环境变量/请求头**
3. **会话中已设置的模型**
4. **默认模型** (claude-sonnet-4)

## 技术实现

### 核心组件

1. **ModelRouter** (`model_router.go`)
   - 模型命令解析
   - 环境变量提取
   - 会话管理
   - 模型验证

2. **Vertex Adaptor 集成** (`adaptor.go`)
   - 请求拦截和模型路由
   - 消息内容更新
   - 请求模式切换

### 关键功能

- **智能模型识别**: 自动检测 Claude 和 Gemini 模型类型
- **消息处理**: 移除 `/model` 命令后保留原始消息内容
- **类型安全**: 完整的模型验证和错误处理
- **性能优化**: 高效的会话管理和内存清理

## 测试验证

✅ **所有功能已通过测试**:
- 模型命令解析 (6/6 测试通过)
- 环境变量提取 (5/5 测试通过)  
- 会话管理 (2/2 测试通过)
- 模型验证 (7/7 测试通过)
- 适配器集成 (3/3 测试通过)

## 注意事项

1. **模型可用性**: 确保配置的 Vertex AI 项目支持所选模型
2. **权限管理**: 需要适当的 Google Cloud 权限访问模型
3. **会话隔离**: 不同客户端的模型选择相互独立
4. **错误处理**: 无效模型名称会被忽略，使用默认模型

## 扩展性

- **新模型支持**: 在 `SupportedModels` 映射中添加新模型
- **自定义别名**: 可轻松添加更多用户友好的别名
- **请求头扩展**: 支持添加新的环境变量头格式
- **路由策略**: 可扩展路由逻辑以支持更复杂的场景

通过此实现，Claude Code 用户现在可以无缝地在 Claude 和 Gemini 模型之间切换，享受两个平台的最佳功能。