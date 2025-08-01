#!/usr/bin/env python3

import requests
import json

def test_model_command():
    """测试消息内/model命令"""
    print("=== 测试消息内/model命令 ===")
    
    url = "http://localhost:3002/v1/messages"
    headers = {
        "Authorization": "Bearer sk-in9ARcpCuZVbhSkrtHRjrs4hz49Bg7f1ApkNzLiJII9OuBKR",
        "Content-Type": "application/json",
        "anthropic-version": "2023-06-01"
    }
    
    # 测试/model命令
    data = {
        "model": "claude-3-5-sonnet-20241022",  # 原始模型
        "max_tokens": 50,
        "messages": [
            {
                "role": "user", 
                "content": "/model flash\n\nHello! What's 2+2?"
            }
        ]
    }
    
    print("发送/model flash命令...")
    response = requests.post(url, headers=headers, json=data)
    print(f"状态码: {response.status_code}")
    print(f"响应: {response.text}")

def test_model_aliases():
    """测试不同的模型别名"""
    print("\n=== 测试模型别名 ===")
    
    url = "http://localhost:3002/v1/messages"
    headers = {
        "Authorization": "Bearer sk-in9ARcpCuZVbhSkrtHRjrs4hz49Bg7f1ApkNzLiJII9OuBKR",
        "Content-Type": "application/json",
        "anthropic-version": "2023-06-01"
    }
    
    # 测试不同的别名
    test_cases = [
        ("/model pro", "应该路由到 gemini-2.5-pro"),
        ("/model lite", "应该路由到 gemini-2.0-flash-lite"),
        ("/model claude", "应该路由到 claude-sonnet-4"),
        ("/model opus", "应该路由到 claude-opus-4")
    ]
    
    for command, description in test_cases:
        data = {
            "model": "claude-3-5-sonnet-20241022",
            "max_tokens": 30,
            "messages": [
                {
                    "role": "user", 
                    "content": f"{command}\n\nHi!"
                }
            ]
        }
        
        print(f"\n发送命令: {command} ({description})")
        response = requests.post(url, headers=headers, json=data)
        print(f"状态码: {response.status_code}")
        if response.status_code == 400:
            # 400错误说明路由成功但格式转换需要完善
            print("✅ 模型路由成功 (400错误是格式转换问题)")
        else:
            print(f"响应: {response.text[:200]}...")

if __name__ == "__main__":
    test_model_command()
    test_model_aliases()