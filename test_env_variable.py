#!/usr/bin/env python3

import requests
import json
import os

def test_environment_variable_simple():
    """简单测试环境变量是否工作"""
    print("当前Python进程的CLAUDE_CUSTOM_MODEL:", os.environ.get('CLAUDE_CUSTOM_MODEL', 'NOT_SET'))
    
    url = "http://localhost:3002/v1/messages"
    headers = {
        "Authorization": "Bearer sk-in9ARcpCuZVbhSkrtHRjrs4hz49Bg7f1ApkNzLiJII9OuBKR",
        "Content-Type": "application/json",
        "anthropic-version": "2023-06-01"
    }
    
    # 简单消息测试
    data = {
        "model": "claude-3-5-sonnet-20241022",  # 原始模型，应该被环境变量覆盖
        "max_tokens": 50,
        "messages": [
            {
                "role": "user", 
                "content": "Hello, what model are you running on?"
            }
        ]
    }
    
    print("发送请求...")
    response = requests.post(url, headers=headers, json=data)
    print(f"状态码: {response.status_code}")
    print(f"响应: {response.text}")

if __name__ == "__main__":
    # 确保环境变量设置
    os.environ['CLAUDE_CUSTOM_MODEL'] = 'gemini-2.5-pro'
    test_environment_variable_simple()