#!/usr/bin/env python3

import requests
import json
import os

# Test the fixed message conversion
API_KEY = "sk-in9ARcpCuZVbhSkrtHRjrs4hz49Bg7f1ApkNzLiJII9OuBKR"
BASE_URL = "http://localhost:3002"

def test_gemini_with_model_command():
    """Test model switching with /model command"""
    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }
    
    # Test with /model command to switch to Gemini
    data = {
        "model": "claude-3-5-sonnet-20241022",
        "max_tokens": 100,
        "messages": [
            {
                "role": "user", 
                "content": "/model gemini-2.5-flash\n\nHello, what model are you?"
            }
        ]
    }
    
    print("Testing with /model command...")
    print(f"Request: {json.dumps(data, indent=2)}")
    
    try:
        response = requests.post(f"{BASE_URL}/v1/messages", 
                               headers=headers, 
                               json=data, 
                               timeout=30)
        
        print(f"Status Code: {response.status_code}")
        print(f"Response Headers: {dict(response.headers)}")
        
        if response.headers.get('content-type', '').startswith('application/json'):
            try:
                response_json = response.json()
                print(f"Response: {json.dumps(response_json, indent=2)}")
            except:
                print(f"Failed to parse JSON. Raw response: {response.text}")
        else:
            print(f"Non-JSON Response: {response.text}")
            
    except Exception as e:
        print(f"Request failed: {e}")

if __name__ == "__main__":
    test_gemini_with_model_command()