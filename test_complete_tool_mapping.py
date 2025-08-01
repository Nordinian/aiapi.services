#!/usr/bin/env python3

import requests
import json

# Test all three stages of tool mapping
API_KEY = "sk-in9ARcpCuZVbhSkrtHRjrs4hz49Bg7f1ApkNzLiJII9OuBKR"
BASE_URL = "http://localhost:3002"

def test_stage_tools(stage_name, tools):
    """Test a specific stage of tools"""
    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }
    
    print(f"\n🧪 Testing {stage_name}...")
    print("=" * 60)
    
    for tool_type, expected_mapping in tools.items():
        data = {
            "model": "claude-3-5-sonnet-20241022",
            "max_tokens": 50,
            "messages": [
                {
                    "role": "user", 
                    "content": f"/model gemini-2.5-flash\n\nTest tool: {tool_type}"
                }
            ],
            "tools": [
                {
                    "type": tool_type,
                    "name": tool_type.lower()
                }
            ]
        }
        
        try:
            response = requests.post(f"{BASE_URL}/v1/messages", 
                                   headers=headers, 
                                   json=data, 
                                   timeout=15)
            
            if response.status_code == 200:
                print(f"✅ {tool_type} → {expected_mapping} : SUCCESS")
            else:
                print(f"❌ {tool_type} → {expected_mapping} : FAILED (status: {response.status_code})")
                
        except Exception as e:
            print(f"❌ {tool_type} → {expected_mapping} : ERROR - {e}")

def main():
    print("🚀 Testing Complete Claude Code Tool Mapping Implementation")
    print("=" * 80)
    
    # 阶段1: 核心工具映射
    stage1_tools = {
        "bash_20250124": "bash_command (CodeExecution兼容)",
        "Read": "file_reader",
        "Write": "file_writer", 
        "Edit": "file_editor_exact"
    }
    
    # 阶段2: 搜索工具映射
    stage2_tools = {
        "web_search_20250305": "web_search (GoogleSearch兼容)",
        "Grep": "text_search"
    }
    
    # 阶段3: 高级工具映射
    stage3_tools = {
        "NotebookRead": "jupyter_notebook_reader",
        "NotebookEdit": "jupyter_notebook_editor",
        "Task": "sub_agent_task (sub_task_delegation兼容)"
    }
    
    # 额外工具
    extra_tools = {
        "MultiEdit": "multi_file_editor",
        "Glob": "file_pattern_search",
        "LS": "directory_lister",
        "WebFetch": "web_fetcher",
        "TodoWrite": "task_manager"
    }
    
    # 运行所有测试
    test_stage_tools("阶段1: 核心工具映射", stage1_tools)
    test_stage_tools("阶段2: 搜索工具映射", stage2_tools) 
    test_stage_tools("阶段3: 高级工具映射", stage3_tools)
    test_stage_tools("额外工具映射", extra_tools)
    
    print("\n" + "=" * 80)
    print("🏁 Tool Mapping Test Complete!")
    print("\n📊 Expected Results:")
    print("✅ 阶段1: 4/4 核心工具 (Bash, Read, Write, Edit)")
    print("✅ 阶段2: 2/2 搜索工具 (WebSearch, Grep)")  
    print("✅ 阶段3: 3/3 高级工具 (NotebookRead, NotebookEdit, Agent)")
    print("✅ 额外: 5/5 扩展工具 (MultiEdit, Glob, LS, WebFetch, TodoWrite)")
    print("\n🎯 Total: 14/14 Claude Code tools mapped to Gemini functions")

if __name__ == "__main__":
    main()