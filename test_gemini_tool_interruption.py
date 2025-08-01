#!/usr/bin/env python3

import requests
import json
import time

# Test Gemini tool interruption issues
API_KEY = "sk-in9ARcpCuZVbhSkrtHRjrs4hz49Bg7f1ApkNzLiJII9OuBKR"
BASE_URL = "http://localhost:3002"

def test_gemini_multistep_task():
    """Test a multi-step task that might cause interruptions"""
    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }
    
    # Test with a complex task that requires multiple tool calls
    data = {
        "model": "claude-3-5-sonnet-20241022",
        "max_tokens": 2000,
        "messages": [
            {
                "role": "user", 
                "content": """/model gemini-2.5-flash

Please help me analyze the current directory structure:
1. First, list all files in the current directory
2. Then, find all .go files in the project
3. Finally, read the main.go file and tell me what it does

This is a multi-step task that should test tool continuity."""
            }
        ],
        "tools": [
            {"type": "bash_20250124", "name": "bash"},
            {"type": "LS", "name": "LS"}, 
            {"type": "Glob", "name": "Glob"},
            {"type": "Read", "name": "Read"}
        ]
    }
    
    print("🧪 Testing multi-step Gemini task...")
    print(f"Request: {json.dumps(data, indent=2)[:500]}...")
    
    try:
        response = requests.post(f"{BASE_URL}/v1/messages", 
                               headers=headers, 
                               json=data, 
                               timeout=60)
        
        print(f"\nStatus Code: {response.status_code}")
        
        if response.status_code == 200:
            try:
                response_json = response.json()
                print(f"\nResponse Type: {response_json.get('type', 'unknown')}")
                print(f"Stop Reason: {response_json.get('stop_reason', 'unknown')}")
                
                content = response_json.get('content', [])
                print(f"Content Items: {len(content)}")
                
                for i, item in enumerate(content):
                    item_type = item.get('type', 'unknown')
                    print(f"  Item {i+1}: {item_type}")
                    
                    if item_type == 'tool_use':
                        print(f"    Tool: {item.get('name', 'unknown')}")
                    elif item_type == 'text':
                        text = item.get('text', '')[:100]
                        print(f"    Text: {text}...")
                
                # Check if task was interrupted
                if response_json.get('stop_reason') == 'max_tokens':
                    print("\n❌ Task likely interrupted due to max_tokens")
                elif len(content) == 1 and content[0].get('type') == 'tool_use':
                    print("\n⚠️  Task may be incomplete - only one tool call made")
                else:
                    print("\n✅ Task appears complete")
                    
            except Exception as e:
                print(f"\n❌ Failed to parse response: {e}")
                print(f"Raw response: {response.text}")
        else:
            print(f"\n❌ HTTP Error: {response.status_code}")
            print(f"Response: {response.text}")
            
    except Exception as e:
        print(f"\n❌ Request failed: {e}")

def test_simple_vs_complex():
    """Compare simple vs complex tasks"""
    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }
    
    # Simple task
    simple_data = {
        "model": "claude-3-5-sonnet-20241022",
        "max_tokens": 1000,
        "messages": [
            {
                "role": "user", 
                "content": "/model gemini-2.5-flash\n\nJust say hello and tell me what model you are."
            }
        ]
    }
    
    print("\n" + "="*60)
    print("🔄 Testing Simple Task vs Complex Task...")
    
    print("\n1. Simple Task (no tools):")
    try:
        response = requests.post(f"{BASE_URL}/v1/messages", 
                               headers=headers, 
                               json=simple_data, 
                               timeout=30)
        
        if response.status_code == 200:
            result = response.json()
            print(f"✅ Simple task: {result.get('stop_reason', 'unknown')}")
        else:
            print(f"❌ Simple task failed: {response.status_code}")
    except Exception as e:
        print(f"❌ Simple task error: {e}")
    
    # Complex task
    complex_data = {
        "model": "claude-3-5-sonnet-20241022", 
        "max_tokens": 1000,
        "messages": [
            {
                "role": "user",
                "content": "/model gemini-2.5-flash\n\nPlease list the files in the current directory using the LS tool."
            }
        ],
        "tools": [{"type": "LS", "name": "LS"}]
    }
    
    print("\n2. Complex Task (with tools):")
    try:
        response = requests.post(f"{BASE_URL}/v1/messages",
                               headers=headers,
                               json=complex_data, 
                               timeout=30)
        
        if response.status_code == 200:
            result = response.json()
            print(f"✅ Complex task: {result.get('stop_reason', 'unknown')}")
            
            # Check if tool was actually used
            content = result.get('content', [])
            has_tool_use = any(item.get('type') == 'tool_use' for item in content)
            has_text = any(item.get('type') == 'text' for item in content)
            
            print(f"   Tool use detected: {has_tool_use}")
            print(f"   Text response: {has_text}")
            
        else:
            print(f"❌ Complex task failed: {response.status_code}")
            print(f"   Response: {response.text}")
    except Exception as e:
        print(f"❌ Complex task error: {e}")

if __name__ == "__main__":
    print("🚀 Testing Gemini Tool Interruption Issues")
    print("="*60)
    
    # Test 1: Multi-step task
    test_gemini_multistep_task()
    
    # Test 2: Simple vs Complex
    test_simple_vs_complex()
    
    print("\n" + "="*60)
    print("🏁 Gemini Tool Interruption Test Complete!")
    print("\nPossible issues to investigate:")
    print("1. Token limits causing premature stopping")
    print("2. Tool conversion causing API errors")
    print("3. Gemini not understanding multi-step instructions")
    print("4. Response format conversion issues")