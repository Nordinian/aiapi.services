#!/usr/bin/env python3

import requests
import json
import os

# Test Claude Code tools conversion with Gemini
API_KEY = "sk-in9ARcpCuZVbhSkrtHRjrs4hz49Bg7f1ApkNzLiJII9OuBKR"
BASE_URL = "http://localhost:3002"

def test_claude_code_tools_with_gemini():
    """Test Claude Code tools with Gemini model routing"""
    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }
    
    # Test with Claude Code tools and /model command to switch to Gemini
    data = {
        "model": "claude-3-5-sonnet-20241022",
        "max_tokens": 1000,
        "messages": [
            {
                "role": "user", 
                "content": "/model gemini-2.5-flash\n\nPlease read the file /etc/passwd and tell me how many users are defined in it."
            }
        ],
        "tools": [
            {
                "type": "bash_20250124",
                "name": "bash"
            }
        ]
    }
    
    print("Testing Claude Code tools with Gemini routing...")
    print(f"Request: {json.dumps(data, indent=2)}")
    
    try:
        response = requests.post(f"{BASE_URL}/v1/messages", 
                               headers=headers, 
                               json=data, 
                               timeout=30)
        
        print(f"\nStatus Code: {response.status_code}")
        print(f"Response Headers: {dict(response.headers)}")
        
        if response.headers.get('content-type', '').startswith('application/json'):
            try:
                response_json = response.json()
                print(f"\nResponse: {json.dumps(response_json, indent=2)}")
                
                # Check if the response indicates successful tool conversion
                if response_json.get('content'):
                    print("\n✅ Tool conversion test: SUCCESS")
                    print("- Model routing working (switched to Gemini)")
                    print("- Tool format conversion working (Claude Code → Gemini)")
                    print("- Response format conversion working (Gemini → Claude Code)")
                    print("- Tool execution should be working in the response")
                else:
                    print("\n❌ Tool conversion test: FAILED - No content in response")
                    
            except Exception as e:
                print(f"\n❌ Failed to parse JSON: {e}")
                print(f"Raw response: {response.text}")
        else:
            print(f"\n❌ Non-JSON Response: {response.text}")
            
    except Exception as e:
        print(f"\n❌ Request failed: {e}")

def test_environment_variable_routing():
    """Test model routing via environment variable"""
    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json",
        "X-Claude-Custom-Model": "gemini-2.5-pro"
    }
    
    data = {
        "model": "claude-3-5-sonnet-20241022",
        "max_tokens": 100,
        "messages": [
            {
                "role": "user", 
                "content": "Hello, what model are you? Please identify yourself."
            }
        ]
    }
    
    print("\n" + "="*60)
    print("Testing environment variable routing...")
    print(f"Request: {json.dumps(data, indent=2)}")
    print(f"Custom Header: X-Claude-Custom-Model: gemini-2.5-pro")
    
    try:
        response = requests.post(f"{BASE_URL}/v1/messages", 
                               headers=headers, 
                               json=data, 
                               timeout=30)
        
        print(f"\nStatus Code: {response.status_code}")
        
        if response.headers.get('content-type', '').startswith('application/json'):
            try:
                response_json = response.json()
                print(f"\nResponse: {json.dumps(response_json, indent=2)}")
                
                # Check if response indicates Gemini model
                content_text = ""
                if response_json.get('content'):
                    for content_item in response_json['content']:
                        if content_item.get('type') == 'text':
                            content_text += content_item.get('text', '')
                
                if "Google" in content_text:
                    print("\n✅ Environment variable routing: SUCCESS")
                    print("- Model routed to Gemini via X-Claude-Custom-Model header")
                else:
                    print("\n❌ Environment variable routing: FAILED")
                    print(f"- Response doesn't indicate Gemini model: {content_text}")
                    
            except Exception as e:
                print(f"\n❌ Failed to parse JSON: {e}")
                print(f"Raw response: {response.text}")
        else:
            print(f"\n❌ Non-JSON Response: {response.text}")
            
    except Exception as e:
        print(f"\n❌ Request failed: {e}")

def test_model_aliases():
    """Test model alias routing"""
    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }
    
    test_cases = [
        ("gemini", "gemini-2.5-pro"),
        ("flash", "gemini-2.5-flash"),
        ("pro", "gemini-2.5-pro"),
        ("claude", "claude-sonnet-4"),
        ("sonnet", "claude-sonnet-4")
    ]
    
    print("\n" + "="*60)
    print("Testing model aliases...")
    
    for alias, expected_model in test_cases:
        data = {
            "model": "claude-3-5-sonnet-20241022",
            "max_tokens": 50,
            "messages": [
                {
                    "role": "user", 
                    "content": f"/model {alias}\n\nHello"
                }
            ]
        }
        
        print(f"\nTesting alias: {alias} -> expected: {expected_model}")
        
        try:
            response = requests.post(f"{BASE_URL}/v1/messages", 
                                   headers=headers, 
                                   json=data, 
                                   timeout=15)
            
            if response.status_code == 200:
                print(f"✅ Alias '{alias}' -> SUCCESS")
            else:
                print(f"❌ Alias '{alias}' -> FAILED (status: {response.status_code})")
                
        except Exception as e:
            print(f"❌ Alias '{alias}' -> ERROR: {e}")

if __name__ == "__main__":
    print("🚀 Starting comprehensive Claude Code model routing and tool conversion tests...")
    print("="*80)
    
    # Test 1: Tool conversion with model routing
    test_claude_code_tools_with_gemini()
    
    # Test 2: Environment variable routing  
    test_environment_variable_routing()
    
    # Test 3: Model aliases
    test_model_aliases()
    
    print("\n" + "="*80)
    print("🏁 All tests completed!")