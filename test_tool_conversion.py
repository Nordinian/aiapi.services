#!/usr/bin/env python3

import requests
import json

def test_gemini_with_simple_message():
    """测试Gemini模型的简单消息（不带工具）"""
    print("=== 测试Gemini模型简单消息 ===")
    
    url = "http://localhost:3002/v1/messages"
    headers = {
        "Authorization": "Bearer sk-in9ARcpCuZVbhSkrtHRjrs4hz49Bg7f1ApkNzLiJII9OuBKR",
        "Content-Type": "application/json",
        "anthropic-version": "2023-06-01"
    }
    
    data = {
        "model": "claude-3-5-sonnet-20241022",
        "max_tokens": 50,
        "messages": [
            {
                "role": "user", 
                "content": "/model gemini-2.5-pro\n\nHello! What's 2+2? Just give me a short answer."
            }
        ]
    }
    
    print("发送请求到Gemini...")
    response = requests.post(url, headers=headers, json=data)
    print(f"状态码: {response.status_code}")
    
    if response.status_code == 200:
        try:
            result = response.json()
            print("✅ 成功！响应格式转换正常")
            print(f"响应内容: {json.dumps(result, ensure_ascii=False, indent=2)}")
        except json.JSONDecodeError:
            print("❌ JSON解析失败")
            print(f"原始响应: {response.text}")
    else:
        print(f"❌ 请求失败: {response.text}")

if __name__ == "__main__":
    test_gemini_with_simple_message()