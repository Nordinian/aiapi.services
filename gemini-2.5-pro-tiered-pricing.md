# Gemini 2.5 Pro 分层定价实现记录

## 概述

实现了 Gemini 2.5 Pro 模型基于输入token数量的分层定价逻辑，解决了不同context window下计费不一致的问题。

## 问题描述

**原问题**: Gemini 2.5 Pro 在不同的context window（不同数量的输入词元）的计费是不同的，但现有代码逻辑使用固定价格 0.625，无法处理这种分层定价情况。

**官方定价结构**:
- **≤200k input tokens**: $1.25/1M tokens input, $10/1M tokens output
- **>200k input tokens**: $2.5/1M tokens input, $15/1M tokens output
- **Batch API**: 还有不同的定价

## 实现方案

### 1. 修改 ModelPriceHelper 函数

**文件**: `/relay/helper/price.go`

**修改位置**: `ModelPriceHelper` 函数中的价格计算逻辑

```go
// 特殊处理 gemini-2.5-pro 的分层定价
if info.OriginModelName == "gemini-2.5-pro" {
    if promptTokens <= 200000 {
        // ≤200k tokens: $1.25/1M tokens input, $10/1M tokens output
        modelRatio = 0.625  // $1.25 / 1M tokens 转换为系统倍率
    } else {
        // >200k tokens: $2.5/1M tokens input, $15/1M tokens output  
        modelRatio = 1.25   // $2.5 / 1M tokens 转换为系统倍率
    }
    success = true
    matchName = info.OriginModelName
}
```

### 2. 同时处理输出token倍率

**同文件位置**: 在 `completionRatio` 计算部分添加特殊处理

```go
// 特殊处理 gemini-2.5-pro 的输出token倍率
if info.OriginModelName == "gemini-2.5-pro" {
    if promptTokens <= 200000 {
        // ≤200k tokens: $10/1M tokens output / $1.25/1M tokens input = 8倍
        completionRatio = 8.0  
    } else {
        // >200k tokens: $15/1M tokens output / $2.5/1M tokens input = 6倍  
        completionRatio = 6.0
    }
}
```

### 3. 更新配置文件注释

**文件**: `/setting/ratio_setting/model_ratio.go`

**修改1**: 更新基础价格注释
```go
"gemini-2.5-pro": 0.625, // 基础价格，实际根据输入token数量分层：≤200k:0.625, >200k:1.25
```

**修改2**: 更新完成倍率注释
```go
} else if strings.HasPrefix(name, "gemini-2.5-pro") { // 基础倍率，实际根据输入token数量动态计算：≤200k:8倍, >200k:6倍
    return 8, false
```

## 分层定价规则

| 输入Token数量 | 输入价格 | 输出价格 | 系统倍率 | 输出倍率 | 说明 |
|--------------|---------|---------|----------|----------|------|
| ≤200k tokens | $1.25/1M | $10/1M | 0.625 | 8.0 | 标准定价层 |
| >200k tokens | $2.5/1M | $15/1M | 1.25 | 6.0 | 高容量定价层 |

## 技术实现细节

### 调用链路

1. **Token计算**: `gemini_handler.go` → `getGeminiInputTokens()` → 获取实际输入token数量
2. **价格分层**: `ModelPriceHelper()` → 检查模型名称和token数量 → 动态设置价格倍率
3. **费用计算**: 系统根据分层倍率计算实际消费配额

### 关键参数

- **`promptTokens`**: 输入token数量，用于判断价格层级
- **`modelRatio`**: 输入token价格倍率
- **`completionRatio`**: 输出token相对于输入token的倍率

### 判断逻辑

```go
if promptTokens <= 200000 {
    // 低容量层 (≤200k tokens)
    modelRatio = 0.625      // $1.25/1M tokens
    completionRatio = 8.0   // $10/1M ÷ $1.25/1M = 8倍
} else {
    // 高容量层 (>200k tokens)  
    modelRatio = 1.25       // $2.5/1M tokens
    completionRatio = 6.0   // $15/1M ÷ $2.5/1M = 6倍
}
```

## 实现特点

### ✅ 优势

1. **精确计费**: 严格按照 Vertex AI 官方分层定价执行
2. **动态调整**: 实时根据请求的输入token数量确定价格层级
3. **完整覆盖**: 同时处理输入和输出token的不同价格倍率
4. **向后兼容**: 不影响其他模型的现有定价逻辑
5. **单一责任**: 只影响 `gemini-2.5-pro` 模型

### 🔧 技术优势

1. **最小侵入**: 在现有架构基础上最小化修改
2. **性能友好**: 在价格计算阶段处理，无额外查询开销
3. **易于维护**: 逻辑集中在 `ModelPriceHelper` 函数中
4. **扩展性强**: 如需添加其他分层定价模型，可复用相同模式

## 测试验证

### 编译测试
```bash
# 验证 price.go 文件编译
go build -o /dev/null ./relay/helper/

# 验证整个项目编译  
go build -o /dev/null .
```

**结果**: ✅ 编译成功，无错误

### 功能测试场景

**场景1**: 小容量请求 (≤200k tokens)
- **输入**: 100k tokens
- **预期**: modelRatio=0.625, completionRatio=8.0

**场景2**: 大容量请求 (>200k tokens)
- **输入**: 300k tokens  
- **预期**: modelRatio=1.25, completionRatio=6.0

**场景3**: 边界值测试
- **输入**: 200000 tokens (边界值)
- **预期**: modelRatio=0.625, completionRatio=8.0

## 相关文件

### 主要修改文件
1. `/relay/helper/price.go` - 核心分层定价逻辑
2. `/setting/ratio_setting/model_ratio.go` - 配置注释更新

### 相关依赖文件
1. `/relay/gemini_handler.go` - Token计算和价格调用
2. `/relay/common/relay_info.go` - RelayInfo结构体定义

## 官方文档参考

- **Vertex AI Gemini API定价**: https://cloud.google.com/vertex-ai/generative-ai/pricing
- **Gemini 2.5 Pro官方定价表**: https://cloud.google.com/vertex-ai/generative-ai/docs/partner-models/use-gemini

## 后续改进建议

### 1. 监控和日志
- 添加分层定价决策的日志记录
- 监控不同价格层级的使用分布

### 2. Batch API支持
- 考虑添加 Batch API 的不同定价支持
- 扩展分层定价逻辑以支持更多价格变体

### 3. 配置化
- 考虑将分层阈值 (200k) 和价格倍率设为可配置参数
- 支持通过配置文件动态调整分层策略

### 4. 其他模型扩展
- 为其他可能需要分层定价的模型提供相同支持
- 建立通用的分层定价框架

## 版本信息

- **实现日期**: 2025-01-02
- **影响版本**: aiapi.services-alpha
- **向后兼容**: 是
- **破坏性变更**: 无

## 总结

成功实现了 Gemini 2.5 Pro 的官方分层定价逻辑，解决了不同context window下计费不准确的问题。实现方案具有良好的扩展性和维护性，为后续其他模型的分层定价需求奠定了基础。