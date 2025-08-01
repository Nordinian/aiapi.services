#!/usr/bin/env python3

import requests
import json
import os

def test_environment_variable_model_routing():
    """测试环境变量模型路由"""
    print("=== 测试环境变量模型路由 ===")
    
    # 设置环境变量
    os.environ['CLAUDE_CUSTOM_MODEL'] = 'gemini-2.5-pro'
    
    url = "http://localhost:3002/v1/messages"
    headers = {
        "Authorization": "Bearer sk-in9ARcpCuZVbhSkrtHRjrs4hz49Bg7f1ApkNzLiJII9OuBKR",
        "Content-Type": "application/json",
        "anthropic-version": "2023-06-01"
    }
    
    # 简单消息测试（无工具）
    simple_data = {
        "model": "claude-3-5-sonnet-20241022",
        "max_tokens": 100,
        "messages": [
            {
                "role": "user", 
                "content": "Hello! What model are you?"
            }
        ]
    }
    
    print("发送简单消息请求...")
    response = requests.post(url, headers=headers, json=simple_data)
    print(f"状态码: {response.status_code}")
    
    if response.status_code == 200:
        try:
            result = response.json()
            print("成功！环境变量模型路由工作正常")
            print(f"响应: {result.get('content', [{}])[0].get('text', 'No response')}")
        except:
            print("响应解析失败")
    else:
        print(f"失败: {response.text}")

def test_tool_conversion():
    """测试完整的工具转换"""
    print("\n=== 测试Claude Code工具转换 ===")
    
    url = "http://localhost:3002/v1/messages"
    headers = {
        "Authorization": "Bearer sk-in9ARcpCuZVbhSkrtHRjrs4hz49Bg7f1ApkNzLiJII9OuBKR",
        "Content-Type": "application/json",
        "anthropic-version": "2023-06-01"
    }
    
    # 包含Claude Code工具的请求
    tool_data = {
        "model": "claude-3-5-sonnet-20241022",
        "max_tokens": 200,
        "messages": [
            {
                "role": "user", 
                "content": "model gemini-2.5-pro\n\nCan you help me list the files in the current directory using the bash tool?"
            }
        ],
        "tools": [
            {
                "type": "bash_20250124",
                "name": "bash",
                "description": "Execute bash commands",
                "input_schema": {
                    "type": "object",
                    "properties": {
                        "command": {
                            "type": "string",
                            "description": "Shell command to execute"
                        },
                        "timeout": {
                            "type": "number",
                            "description": "Timeout in milliseconds"
                        }
                    },
                    "required": ["command"]
                }
            },
            {
                "type": "str_replace_based_edit_tool",
                "name": "str_replace_based_edit_tool",
                "description": "Edit files with string replacement",
                "input_schema": {
                    "type": "object",
                    "properties": {
                        "command": {
                            "type": "string",
                            "description": "Command: view, str_replace, create"
                        },
                        "path": {
                            "type": "string",
                            "description": "File path"
                        }
                    },
                    "required": ["command", "path"]
                }
            }
        ]
    }
    
    print("发送包含工具的请求...")
    response = requests.post(url, headers=headers, json=tool_data)
    print(f"状态码: {response.status_code}")
    
    if response.status_code == 200:
        try:
            result = response.json()
            print("成功！工具转换正常工作")
            print(f"响应: {json.dumps(result, indent=2, ensure_ascii=False)}")
        except Exception as e:
            print(f"响应解析失败: {e}")
            print(f"原始响应: {response.text}")
    else:
        print(f"失败: {response.text}")

def test_message_command_routing():
    """测试消息内/model命令路由"""
    print("\n=== 测试消息内模型切换 ===")
    
    url = "http://localhost:3002/v1/messages"
    headers = {
        "Authorization": "Bearer sk-in9ARcpCuZVbhSkrtHRjrs4hz49Bg7f1ApkNzLiJII9OuBKR",
        "Content-Type": "application/json",
        "anthropic-version": "2023-06-01"
    }
    
    command_data = {
        "model": "claude-3-5-sonnet-20241022",
        "max_tokens": 50,
        "messages": [
            {
                "role": "user", 
                "content": "/model flash\n\nHello! What's 2+2?"
            }
        ]
    }
    
    print("发送带/model命令的请求...")
    response = requests.post(url, headers=headers, json=command_data)
    print(f"状态码: {response.status_code}")
    
    if response.status_code == 200:
        try:
            result = response.json()
            print("成功！消息内模型路由工作正常")
            print(f"响应: {result.get('content', [{}])[0].get('text', 'No response')}")
        except:
            print("响应解析失败")
    else:
        print(f"失败: {response.text}")

if __name__ == "__main__":
    print("Claude Code Gemini兼容性完整测试\n")
    
    # 测试环境变量支持
    test_environment_variable_model_routing()
    
    # 测试工具转换
    test_tool_conversion()
    
    # 测试消息命令路由
    test_message_command_routing()
    
    print("\n测试完成！")