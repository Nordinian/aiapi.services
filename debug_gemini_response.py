#!/usr/bin/env python3

import requests
import json

# Debug Gemini response format issues
API_KEY = "sk-in9ARcpCuZVbhSkrtHRjrs4hz49Bg7f1ApkNzLiJII9OuBKR"
BASE_URL = "http://localhost:3002"

def test_detailed_gemini_response():
    """Test with detailed response analysis"""
    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }
    
    # Simple tool test
    data = {
        "model": "claude-3-5-sonnet-20241022",
        "max_tokens": 500,
        "messages": [
            {
                "role": "user", 
                "content": "/model gemini-2.5-flash\n\nPlease list the current directory using LS tool."
            }
        ],
        "tools": [
            {"type": "LS", "name": "LS"}
        ]
    }
    
    print("🔍 Detailed Gemini Response Analysis")
    print("="*50)
    
    try:
        response = requests.post(f"{BASE_URL}/v1/messages", 
                               headers=headers, 
                               json=data, 
                               timeout=30)
        
        print(f"Status Code: {response.status_code}")
        print(f"Response Headers: {dict(response.headers)}")
        
        if response.status_code == 200:
            response_json = response.json()
            
            print(f"\n📊 Response Structure:")
            print(f"  - Type: {response_json.get('type')}")
            print(f"  - Role: {response_json.get('role')}")
            print(f"  - Model: {response_json.get('model')}")
            print(f"  - Stop Reason: {response_json.get('stop_reason')}")
            print(f"  - ID: {response_json.get('id')}")
            
            content = response_json.get('content', [])
            print(f"\n📝 Content Analysis ({len(content)} items):")
            
            for i, item in enumerate(content):
                print(f"  Item {i+1}:")
                print(f"    Type: {item.get('type')}")
                
                if item.get('type') == 'tool_use':
                    print(f"    Name: {item.get('name')}")
                    print(f"    Input Keys: {list(item.get('input', {}).keys())}")
                    input_data = item.get('input', {})
                    for key, value in input_data.items():
                        value_str = str(value)[:50] + "..." if len(str(value)) > 50 else str(value)
                        print(f"      {key}: {value_str}")
                        
                elif item.get('type') == 'text':
                    text = item.get('text', '')
                    print(f"    Text Length: {len(text)}")
                    print(f"    Text Preview: {text[:100]}...")
                    
            usage = response_json.get('usage', {})
            print(f"\n📈 Usage Info:")
            print(f"  Input Tokens: {usage.get('input_tokens', 'N/A')}")
            print(f"  Output Tokens: {usage.get('output_tokens', 'N/A')}")
            
            # Check for issues
            print(f"\n🔍 Issue Detection:")
            
            if not content:
                print("  ❌ ISSUE: No content in response")
            elif len(content) == 1 and content[0].get('type') == 'tool_use':
                print("  ⚠️  ISSUE: Only tool call, no text explanation")
            elif response_json.get('stop_reason') == 'max_tokens':
                print("  ❌ ISSUE: Response truncated by max_tokens")
            elif all(item.get('type') == 'text' and not item.get('text', '').strip() for item in content):
                print("  ❌ ISSUE: Empty text content")
            else:
                print("  ✅ Response structure looks normal")
                
        else:
            print(f"\n❌ HTTP Error Response:")
            print(response.text)
            
    except Exception as e:
        print(f"\n❌ Request Error: {e}")

def compare_claude_vs_gemini():
    """Compare same request to Claude vs Gemini"""
    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }
    
    base_data = {
        "max_tokens": 500,
        "messages": [
            {
                "role": "user", 
                "content": "Please list the files in the current directory using the LS tool, then tell me what you found."
            }
        ],
        "tools": [{"type": "LS", "name": "LS"}]
    }
    
    print("\n" + "="*50)
    print("🆚 Claude vs Gemini Comparison")
    
    # Test Claude
    claude_data = {**base_data, "model": "claude-3-5-sonnet-20241022"}
    print("\n1. Claude Response:")
    
    try:
        response = requests.post(f"{BASE_URL}/v1/messages", 
                               headers=headers, 
                               json=claude_data, 
                               timeout=30)
        
        if response.status_code == 200:
            result = response.json()
            content = result.get('content', [])
            print(f"   ✅ Status: {response.status_code}")
            print(f"   Stop Reason: {result.get('stop_reason')}")
            print(f"   Content Items: {len(content)}")
            
            for i, item in enumerate(content):
                print(f"     {i+1}. {item.get('type')}")
        else:
            print(f"   ❌ Failed: {response.status_code}")
            
    except Exception as e:
        print(f"   ❌ Error: {e}")
    
    # Test Gemini
    gemini_data = {**base_data, "model": "claude-3-5-sonnet-20241022"}
    gemini_data["messages"][0]["content"] = "/model gemini-2.5-flash\n\n" + gemini_data["messages"][0]["content"]
    
    print("\n2. Gemini Response:")
    
    try:
        response = requests.post(f"{BASE_URL}/v1/messages", 
                               headers=headers, 
                               json=gemini_data, 
                               timeout=30)
        
        if response.status_code == 200:
            result = response.json()
            content = result.get('content', [])
            print(f"   ✅ Status: {response.status_code}")
            print(f"   Stop Reason: {result.get('stop_reason')}")
            print(f"   Content Items: {len(content)}")
            
            for i, item in enumerate(content):
                print(f"     {i+1}. {item.get('type')}")
        else:
            print(f"   ❌ Failed: {response.status_code}")
            print(f"   Response: {response.text}")
            
    except Exception as e:
        print(f"   ❌ Error: {e}")

if __name__ == "__main__":
    test_detailed_gemini_response()
    compare_claude_vs_gemini()
    
    print("\n" + "="*50)
    print("🎯 Debug Summary Complete!")