#!/usr/bin/env python3

import requests
import json

# Test improved Gemini behavior with enhanced system prompts
API_KEY = "sk-in9ARcpCuZVbhSkrtHRjrs4hz49Bg7f1ApkNzLiJII9OuBKR"
BASE_URL = "http://localhost:3002"

def test_improved_gemini_behavior():
    """Test if Gemini now behaves more like Claude with enhanced prompts"""
    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }
    
    # Test cases to verify improved behavior
    test_cases = [
        {
            "name": "Simple Directory Listing",
            "content": "/model gemini-2.5-flash\n\nPlease list the files in the current directory using the LS tool.",
            "expected": "Should explain what it's doing, use LS tool, then explain results"
        },
        {
            "name": "Multi-step Task",
            "content": "/model gemini-2.5-flash\n\nHelp me understand this project: first list the files, then read the main.go file.",
            "expected": "Should explain the plan, execute steps with explanations"
        },
        {
            "name": "File Analysis",
            "content": "/model gemini-2.5-flash\n\nFind all .go files in this project and tell me what the main functionality is.",
            "expected": "Should explain the search strategy, use tools, provide analysis"
        }
    ]
    
    print("🧪 Testing Improved Gemini Behavior")
    print("="*60)
    
    for i, test_case in enumerate(test_cases, 1):
        print(f"\n{i}. {test_case['name']}")
        print(f"Expected: {test_case['expected']}")
        print("-" * 50)
        
        data = {
            "model": "claude-3-5-sonnet-20241022",
            "max_tokens": 1500,
            "messages": [
                {
                    "role": "user", 
                    "content": test_case["content"]
                }
            ],
            "tools": [
                {"type": "LS", "name": "LS"},
                {"type": "Glob", "name": "Glob"},
                {"type": "Read", "name": "Read"}
            ]
        }
        
        try:
            response = requests.post(f"{BASE_URL}/v1/messages", 
                                   headers=headers, 
                                   json=data, 
                                   timeout=45)
            
            if response.status_code == 200:
                result = response.json()
                content = result.get('content', [])
                
                print(f"✅ Status: {response.status_code}")
                print(f"Stop Reason: {result.get('stop_reason')}")
                print(f"Content Items: {len(content)}")
                
                # Analyze response structure
                has_text_before_tools = False
                has_text_after_tools = False
                tool_calls = []
                text_parts = []
                
                for item in content:
                    if item.get('type') == 'text':
                        text_parts.append(item.get('text', ''))
                    elif item.get('type') == 'tool_use':
                        tool_calls.append(item.get('name', 'unknown'))
                
                # Check for improved behavior
                all_text = ' '.join(text_parts)
                
                if tool_calls:
                    print(f"🔧 Tools Used: {', '.join(tool_calls)}")
                
                if text_parts:
                    print(f"📝 Text Response: Yes ({len(all_text)} chars)")
                    print(f"   Preview: {all_text[:150]}...")
                else:
                    print(f"📝 Text Response: None ❌")
                
                # Behavior analysis
                behavior_score = 0
                if text_parts:
                    behavior_score += 1
                    if len(text_parts) > 1:  # Multiple text parts suggest explanation
                        behavior_score += 1
                if tool_calls:
                    behavior_score += 1
                
                if behavior_score >= 2:
                    print(f"🎯 Behavior: Improved ✅ (score: {behavior_score}/3)")
                else:
                    print(f"🎯 Behavior: Still Basic ⚠️ (score: {behavior_score}/3)")
                    
            else:
                print(f"❌ HTTP Error: {response.status_code}")
                print(f"Response: {response.text}")
                
        except Exception as e:
            print(f"❌ Error: {e}")
        
        print()

if __name__ == "__main__":
    test_improved_gemini_behavior()
    
    print("="*60)
    print("🏁 Improved Gemini Behavior Test Complete!")
    print("\nLook for:")
    print("1. ✅ Text explanations before and after tool use")
    print("2. ✅ Multiple content items (not just single tool call)")
    print("3. ✅ Conversational, Claude-like behavior")
    print("4. ❌ Silent tool execution without explanation")