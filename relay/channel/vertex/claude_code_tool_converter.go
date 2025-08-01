package vertex

import (
	"fmt"
	"one-api/dto"
)

// Claude Code工具类型到Gemini函数的映射表
var claudeCodeToolMappings = map[string]GeminiFunctionMapping{
	// 核心文件操作工具
	"bash_20250124": {
		Name:        "bash_command",
		Description: "Execute bash commands in the system shell",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"command": map[string]interface{}{
					"type":        "string",
					"description": "The bash command to execute",
				},
				"timeout": map[string]interface{}{
					"type":        "number",
					"description": "Optional timeout in milliseconds (max 600000)",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Clear description of what this command does in 5-10 words",
				},
			},
			"required": []string{"command"},
		},
	},
	
	// 文件读取工具
	"str_replace_based_edit_tool": {
		Name:        "file_editor",
		Description: "Read and edit files with string replacement",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"command": map[string]interface{}{
					"type":        "string",
					"description": "The operation: 'view', 'str_replace', 'create'",
				},
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Absolute path to the file",
				},
				"old_str": map[string]interface{}{
					"type":        "string",
					"description": "String to replace (for str_replace command)",
				},
				"new_str": map[string]interface{}{
					"type":        "string",
					"description": "Replacement string (for str_replace command)",
				},
				"file_text": map[string]interface{}{
					"type":        "string",
					"description": "File content (for create command)",
				},
			},
			"required": []string{"command", "path"},
		},
	},
	
	// Web搜索工具
	"web_search_20250305": {
		Name:        "web_search",
		Description: "Search the web for information using various search engines",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "The search query",
				},
				"max_results": map[string]interface{}{
					"type":        "number",
					"description": "Maximum number of results to return (default: 5)",
				},
			},
			"required": []string{"query"},
		},
	},
	
	// Task工具 - 代理任务
	"Task": {
		Name:        "sub_agent_task",
		Description: "Launch a specialized sub-agent for complex multi-step tasks",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Short description of the task (3-5 words)",
				},
				"prompt": map[string]interface{}{
					"type":        "string",
					"description": "Detailed task description for the agent",
				},
				"subagent_type": map[string]interface{}{
					"type":        "string",
					"description": "Type of specialized agent to use",
				},
			},
			"required": []string{"description", "prompt", "subagent_type"},
		},
	},
	
	// Grep工具 - 文本搜索
	"Grep": {
		Name:        "text_search",
		Description: "Search for patterns in files using ripgrep",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pattern": map[string]interface{}{
					"type":        "string",
					"description": "Regular expression pattern to search for",
				},
				"path": map[string]interface{}{
					"type":        "string",
					"description": "File or directory to search in",
				},
				"glob": map[string]interface{}{
					"type":        "string",
					"description": "Glob pattern to filter files",
				},
				"output_mode": map[string]interface{}{
					"type":        "string",
					"description": "Output mode: content, files_with_matches, count",
				},
			},
			"required": []string{"pattern"},
		},
	},
	
	// Glob工具 - 文件模式匹配
	"Glob": {
		Name:        "file_pattern_search",
		Description: "Find files by name patterns using glob syntax",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pattern": map[string]interface{}{
					"type":        "string",
					"description": "Glob pattern to match files against",
				},
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Directory to search in (optional)",
				},
			},
			"required": []string{"pattern"},
		},
	},
	
	// Read工具 - 文件读取
	"Read": {
		Name:        "file_reader",
		Description: "Read file contents from the filesystem",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"file_path": map[string]interface{}{
					"type":        "string",
					"description": "Absolute path to the file to read",
				},
				"limit": map[string]interface{}{
					"type":        "number",
					"description": "Number of lines to read (optional)",
				},
				"offset": map[string]interface{}{
					"type":        "number",
					"description": "Line number to start reading from (optional)",
				},
			},
			"required": []string{"file_path"},
		},
	},
	
	// Write工具 - 文件写入
	"Write": {
		Name:        "file_writer",
		Description: "Write content to files on the filesystem",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"file_path": map[string]interface{}{
					"type":        "string",
					"description": "Absolute path to the file to write",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "Content to write to the file",
				},
			},
			"required": []string{"file_path", "content"},
		},
	},
	
	// Edit工具 - 文件编辑
	"Edit": {
		Name:        "file_editor_exact",
		Description: "Perform exact string replacements in files",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"file_path": map[string]interface{}{
					"type":        "string",
					"description": "Absolute path to the file to modify",
				},
				"old_string": map[string]interface{}{
					"type":        "string",
					"description": "Text to replace",
				},
				"new_string": map[string]interface{}{
					"type":        "string",
					"description": "Replacement text",
				},
				"replace_all": map[string]interface{}{
					"type":        "boolean",
					"description": "Replace all occurrences (default: false)",
				},
			},
			"required": []string{"file_path", "old_string", "new_string"},
		},
	},
	
	// MultiEdit工具 - 多重文件编辑
	"MultiEdit": {
		Name:        "multi_file_editor",
		Description: "Perform multiple edits to a single file in one operation",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"file_path": map[string]interface{}{
					"type":        "string",
					"description": "Absolute path to the file to modify",
				},
				"edits": map[string]interface{}{
					"type":        "array",
					"description": "Array of edit operations",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"old_string": map[string]interface{}{
								"type":        "string",
								"description": "Text to replace",
							},
							"new_string": map[string]interface{}{
								"type":        "string",
								"description": "Replacement text",
							},
							"replace_all": map[string]interface{}{
								"type":        "boolean",
								"description": "Replace all occurrences (default: false)",
							},
						},
						"required": []string{"old_string", "new_string"},
					},
				},
			},
			"required": []string{"file_path", "edits"},
		},
	},
	
	// LS工具 - 目录列表
	"LS": {
		Name:        "directory_lister",
		Description: "List files and directories in a given path",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Absolute path to the directory to list",
				},
				"ignore": map[string]interface{}{
					"type":        "array",
					"description": "List of glob patterns to ignore",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
			},
			"required": []string{"path"},
		},
	},
	
	// WebFetch工具 - 网页获取
	"WebFetch": {
		Name:        "web_fetcher",
		Description: "Fetch and analyze content from web URLs",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"url": map[string]interface{}{
					"type":        "string",
					"description": "URL to fetch content from",
				},
				"prompt": map[string]interface{}{
					"type":        "string",
					"description": "Prompt to analyze the fetched content",
				},
			},
			"required": []string{"url", "prompt"},
		},
	},
	
	// TodoWrite工具 - 任务管理
	"TodoWrite": {
		Name:        "task_manager",
		Description: "Create and manage structured task lists",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"todos": map[string]interface{}{
					"type":        "array",
					"description": "Array of todo items",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"content": map[string]interface{}{
								"type":        "string",
								"description": "Task content description",
							},
							"status": map[string]interface{}{
								"type":        "string",
								"description": "Task status: pending, in_progress, completed",
							},
							"priority": map[string]interface{}{
								"type":        "string",
								"description": "Task priority: high, medium, low",
							},
							"id": map[string]interface{}{
								"type":        "string",
								"description": "Unique task identifier",
							},
						},
						"required": []string{"content", "status", "priority", "id"},
					},
				},
			},
			"required": []string{"todos"},
		},
	},
	
	// NotebookRead工具 - Jupyter笔记本读取
	"NotebookRead": {
		Name:        "jupyter_notebook_reader",
		Description: "Read Jupyter notebook (.ipynb file) and return all cells with their outputs",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"notebook_path": map[string]interface{}{
					"type":        "string",
					"description": "Absolute path to the Jupyter notebook file to read",
				},
				"cell_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of a specific cell to read (optional)",
				},
			},
			"required": []string{"notebook_path"},
		},
	},
	
	// NotebookEdit工具 - Jupyter笔记本编辑
	"NotebookEdit": {
		Name:        "jupyter_notebook_editor",
		Description: "Completely replace the contents of a specific cell in a Jupyter notebook",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"notebook_path": map[string]interface{}{
					"type":        "string",
					"description": "Absolute path to the Jupyter notebook file to edit",
				},
				"new_source": map[string]interface{}{
					"type":        "string",
					"description": "The new source for the cell",
				},
				"cell_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the cell to edit (optional)",
				},
				"cell_type": map[string]interface{}{
					"type":        "string",
					"description": "Type of the cell: code or markdown (optional)",
					"enum": []string{"code", "markdown"},
				},
				"edit_mode": map[string]interface{}{
					"type":        "string",
					"description": "Type of edit: replace, insert, delete (default: replace)",
					"enum": []string{"replace", "insert", "delete"},
				},
			},
			"required": []string{"notebook_path", "new_source"},
		},
	},
}

// GeminiFunctionMapping 定义Gemini函数映射结构
type GeminiFunctionMapping struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ConvertClaudeCodeToolsToGemini 将Claude Code工具转换为Gemini Function Declarations
func ConvertClaudeCodeToolsToGemini(claudeTools []interface{}) ([]map[string]interface{}, error) {
	if claudeTools == nil {
		return nil, nil
	}
	
	var geminiTools []map[string]interface{}
	
	for _, tool := range claudeTools {
		toolMap, ok := tool.(map[string]interface{})
		if !ok {
			continue
		}
		
		// 获取工具类型和名称
		toolType, hasType := toolMap["type"].(string)
		toolName, hasName := toolMap["name"].(string)
		
		if !hasType {
			continue
		}
		
		// 查找映射
		mapping, exists := claudeCodeToolMappings[toolType]
		if !exists {
			// 如果没有映射，尝试使用工具名称查找
			if hasName {
				mapping, exists = claudeCodeToolMappings[toolName]
			}
			if !exists {
				// 如果仍然没有找到，跳过这个工具
				continue
			}
		}
		
		// 创建Gemini函数声明
		geminiFunction := map[string]interface{}{
			"name":        mapping.Name,
			"description": mapping.Description,
			"parameters":  mapping.Parameters,
		}
		
		geminiTools = append(geminiTools, geminiFunction)
	}
	
	return geminiTools, nil
}

// ConvertGeminiFunctionCallToClaudeCode 将Gemini函数调用转换为Claude Code工具调用
func ConvertGeminiFunctionCallToClaudeCode(geminiCall map[string]interface{}) (map[string]interface{}, error) {
	functionCall, ok := geminiCall["functionCall"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid Gemini function call format")
	}
	
	functionName, ok := functionCall["name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing function name in Gemini call")
	}
	
	args, ok := functionCall["args"].(map[string]interface{})
	if !ok {
		args = make(map[string]interface{})
	}
	
	// 反向映射：从Gemini函数名找到Claude Code工具类型
	var claudeToolType string
	for toolType, mapping := range claudeCodeToolMappings {
		if mapping.Name == functionName {
			claudeToolType = toolType
			break
		}
	}
	
	if claudeToolType == "" {
		return nil, fmt.Errorf("no Claude Code tool mapping found for Gemini function: %s", functionName)
	}
	
	// 创建Claude Code工具调用格式
	claudeCall := map[string]interface{}{
		"type":  "tool_use",
		"id":    fmt.Sprintf("toolu_%d", getCurrentTimestamp()),
		"name":  getClaudeToolName(claudeToolType),
		"input": args,
	}
	
	return claudeCall, nil
}

// getClaudeToolName 获取Claude Code工具的标准名称
func getClaudeToolName(toolType string) string {
	switch toolType {
	case "bash_20250124":
		return "bash"
	case "str_replace_based_edit_tool":
		return "str_replace_based_edit_tool"
	case "web_search_20250305":
		return "web_search"
	case "Task":
		return "Task"
	case "Grep":
		return "Grep"
	case "Glob":
		return "Glob"
	case "Read":
		return "Read"
	case "Write":
		return "Write"
	case "Edit":
		return "Edit"
	case "MultiEdit":
		return "MultiEdit"
	case "LS":
		return "LS"
	case "WebFetch":
		return "WebFetch"
	case "TodoWrite":
		return "TodoWrite"
	case "NotebookRead":
		return "NotebookRead"
	case "NotebookEdit":
		return "NotebookEdit"
	default:
		return toolType
	}
}

// ConvertClaudeRequestToGeminiWithTools 完整转换Claude请求到Gemini格式（包含工具）
func ConvertClaudeRequestToGeminiWithTools(claudeRequest map[string]interface{}) (map[string]interface{}, error) {
	fmt.Printf("[DEBUG] Converting Claude request to Gemini format\n")
	fmt.Printf("[DEBUG] Claude request keys: %v\n", getKeys(claudeRequest))
	
	geminiRequest := make(map[string]interface{})
	
	// 复制基本字段
	if model, ok := claudeRequest["model"]; ok {
		geminiRequest["model"] = model
	}
	
	if maxTokens, ok := claudeRequest["max_tokens"]; ok {
		geminiRequest["generationConfig"] = map[string]interface{}{
			"maxOutputTokens": maxTokens,
		}
	}
	
	if temp, ok := claudeRequest["temperature"]; ok {
		if genConfig, exists := geminiRequest["generationConfig"]; exists {
			genConfig.(map[string]interface{})["temperature"] = temp
		} else {
			geminiRequest["generationConfig"] = map[string]interface{}{
				"temperature": temp,
			}
		}
	}
	
	// 转换消息
	messagesField := claudeRequest["messages"]
	fmt.Printf("[DEBUG] Messages field type: %T\n", messagesField)
	
	var geminiMessages []map[string]interface{}
	
	// 处理 []dto.ClaudeMessage 类型
	if claudeMessages, ok := claudeRequest["messages"].([]dto.ClaudeMessage); ok {
		fmt.Printf("[DEBUG] Found %d ClaudeMessage to convert\n", len(claudeMessages))
		geminiMessages = make([]map[string]interface{}, 0, len(claudeMessages))
		
		for i, msg := range claudeMessages {
			fmt.Printf("[DEBUG] ClaudeMessage %d: role=%s\n", i, msg.Role)
			
			// 转换角色映射
			geminiRole := msg.Role
			if geminiRole == "user" {
				geminiRole = "user"
			} else if geminiRole == "assistant" {
				geminiRole = "model"
			}
			
			geminiMsg := map[string]interface{}{
				"role": geminiRole,
			}
			
			// 获取消息内容
			contentStr := msg.GetStringContent()
			if contentStr != "" {
				fmt.Printf("[DEBUG] ClaudeMessage %d content: %s\n", i, contentStr)
				geminiMsg["parts"] = []map[string]interface{}{
					{"text": contentStr},
				}
			} else {
				fmt.Printf("[DEBUG] ClaudeMessage %d has empty content\n", i)
				// 即使内容为空，也要为Gemini创建一个空的parts
				geminiMsg["parts"] = []map[string]interface{}{
					{"text": ""},
				}
			}
			
			geminiMessages = append(geminiMessages, geminiMsg)
		}
	} else if messages, ok := claudeRequest["messages"].([]interface{}); ok {
		// 处理通用 []interface{} 类型（备用方案）
		fmt.Printf("[DEBUG] Found %d interface{} messages to convert\n", len(messages))
		geminiMessages = make([]map[string]interface{}, 0, len(messages))
		
		for i, msg := range messages {
			if msgMap, ok := msg.(map[string]interface{}); ok {
				role, hasRole := msgMap["role"].(string)
				content, hasContent := msgMap["content"]
				fmt.Printf("[DEBUG] Message %d: role=%s, hasContent=%v\n", i, role, hasContent)
				
				if hasRole && hasContent {
					geminiMsg := map[string]interface{}{
						"role": role,
					}
					
					// 处理内容
					if contentStr, ok := content.(string); ok {
						fmt.Printf("[DEBUG] Message %d content (string): %s\n", i, contentStr)
						geminiMsg["parts"] = []map[string]interface{}{
							{"text": contentStr},
						}
					} else {
						fmt.Printf("[DEBUG] Message %d content type: %T\n", i, content)
					}
					
					geminiMessages = append(geminiMessages, geminiMsg)
				}
			}
		}
	} else {
		fmt.Printf("[DEBUG] No messages field found or unrecognized type\n")
	}
	
	// 处理系统消息 - 为Gemini添加行为指导
	var finalSystemMessage string
	if systemMessage, ok := claudeRequest["system"].(string); ok && systemMessage != "" {
		finalSystemMessage = systemMessage
	}
	
	// 为Gemini添加特定的行为指导，使其更像Claude
	claudeStyleGuidance := `When using tools, always:
1. First explain what you're about to do and why
2. Use the appropriate tool
3. After getting the result, explain what you found and how it answers the user's question
Be conversational and helpful like Claude.`
	
	if finalSystemMessage != "" {
		finalSystemMessage = finalSystemMessage + "\n\n" + claudeStyleGuidance
	} else {
		finalSystemMessage = claudeStyleGuidance
	}
	
	fmt.Printf("[DEBUG] Adding enhanced system message to Gemini request\n")
	// 在Gemini中，系统消息需要作为第一条消息添加
	systemMsg := map[string]interface{}{
		"role": "user",
		"parts": []map[string]interface{}{
			{"text": finalSystemMessage},
		},
	}
	// 将系统消息插入到消息列表的开头
	geminiMessages = append([]map[string]interface{}{systemMsg}, geminiMessages...)

	fmt.Printf("[DEBUG] Generated %d Gemini messages (including system)\n", len(geminiMessages))
	geminiRequest["contents"] = geminiMessages
	
	// 转换工具
	if tools, ok := claudeRequest["tools"].([]interface{}); ok {
		geminiTools, err := ConvertClaudeCodeToolsToGemini(tools)
		if err != nil {
			return nil, fmt.Errorf("failed to convert tools: %w", err)
		}
		
		if len(geminiTools) > 0 {
			geminiRequest["tools"] = []map[string]interface{}{
				{
					"functionDeclarations": geminiTools,
				},
			}
		}
	}
	
	return geminiRequest, nil
}

// ConvertGeminiResponseToClaudeFormat 将Gemini响应转换为Claude格式
func ConvertGeminiResponseToClaudeFormat(geminiResponse map[string]interface{}) (map[string]interface{}, error) {
	fmt.Printf("[DEBUG] Converting Gemini response to Claude format\n")
	fmt.Printf("[DEBUG] Gemini response keys: %v\n", getKeys(geminiResponse))
	
	claudeResponse := make(map[string]interface{})
	
	// 设置基本字段
	claudeResponse["id"] = "claude_code_gemini_" + fmt.Sprintf("%d", getCurrentTimestamp())
	claudeResponse["type"] = "message"
	claudeResponse["role"] = "assistant"
	claudeResponse["model"] = getModelFromResponse(geminiResponse)
	
	// 初始化content数组
	var content []map[string]interface{}
	
	// 处理candidates
	if candidates, ok := geminiResponse["candidates"].([]interface{}); ok && len(candidates) > 0 {
		fmt.Printf("[DEBUG] Found %d candidates\n", len(candidates))
		for _, candidate := range candidates {
			if candidateMap, ok := candidate.(map[string]interface{}); ok {
				fmt.Printf("[DEBUG] Candidate keys: %v\n", getKeys(candidateMap))
				if contentObj, ok := candidateMap["content"].(map[string]interface{}); ok {
					fmt.Printf("[DEBUG] Content keys: %v\n", getKeys(contentObj))
					if parts, ok := contentObj["parts"].([]interface{}); ok {
						fmt.Printf("[DEBUG] Found %d parts\n", len(parts))
						for i, part := range parts {
							if partMap, ok := part.(map[string]interface{}); ok {
								fmt.Printf("[DEBUG] Part %d keys: %v\n", i, getKeys(partMap))
								// 处理文本内容
								if text, ok := partMap["text"].(string); ok && text != "" {
									content = append(content, map[string]interface{}{
										"type": "text",
										"text": text,
									})
								}
								
								// 处理函数调用
								if _, ok := partMap["functionCall"].(map[string]interface{}); ok {
									claudeToolUse, err := ConvertGeminiFunctionCallToClaudeCode(partMap)
									if err != nil {
										// 如果转换失败，记录错误但继续处理
										continue
									}
									content = append(content, claudeToolUse)
								}
							}
						}
					}
				}
			}
		}
	}
	
	claudeResponse["content"] = content
	
	// 处理usage信息
	if usageMetadata, ok := geminiResponse["usageMetadata"].(map[string]interface{}); ok {
		usage := make(map[string]interface{})
		
		if promptTokenCount, ok := usageMetadata["promptTokenCount"]; ok {
			usage["input_tokens"] = promptTokenCount
		}
		
		if candidatesTokenCount, ok := usageMetadata["candidatesTokenCount"]; ok {
			usage["output_tokens"] = candidatesTokenCount
		}
		
		claudeResponse["usage"] = usage
	}
	
	// 设置stop_reason
	claudeResponse["stop_reason"] = getStopReason(geminiResponse)
	claudeResponse["stop_sequence"] = nil
	
	return claudeResponse, nil
}

// getCurrentTimestamp 获取当前时间戳
func getCurrentTimestamp() int64 {
	return 1707839285 // 固定时间戳，可以改为time.Now().Unix()
}

// getModelFromResponse 从Gemini响应中获取模型名
func getModelFromResponse(response map[string]interface{}) string {
	// 尝试从响应中获取模型信息
	if model, ok := response["model"].(string); ok {
		return model
	}
	return "gemini-model" // 默认值
}

// getStopReason 获取停止原因
func getStopReason(response map[string]interface{}) string {
	if candidates, ok := response["candidates"].([]interface{}); ok && len(candidates) > 0 {
		if candidate, ok := candidates[0].(map[string]interface{}); ok {
			if finishReason, ok := candidate["finishReason"].(string); ok {
				switch finishReason {
				case "STOP":
					return "end_turn"
				case "MAX_TOKENS":
					return "max_tokens"
				case "SAFETY":
					return "stop_sequence"
				case "RECITATION":
					return "stop_sequence"
				default:
					return "end_turn"
				}
			}
		}
	}
	return "end_turn"
}

// getKeys helper function to get map keys for debugging
func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}