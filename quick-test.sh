#!/bin/bash

# 快速测试 MCP 服务器功能

echo "🚀 测试系统监控 MCP 服务器..."

# 测试初始化
echo "1. 测试初始化..."
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}' | timeout 5 ./system-monitor | head -1

echo -e "\n2. 测试工具列表..."
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | timeout 5 ./system-monitor | head -1

echo -e "\n3. 测试 CPU 信息..."
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"cpu_info","arguments":{"duration":"1s"}}}' | timeout 5 ./system-monitor | head -1

echo -e "\n4. 测试内存信息..."
echo '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"memory_info","arguments":{}}}' | timeout 5 ./system-monitor | head -1

echo -e "\n5. 测试系统概览..."
echo '{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"system_overview","arguments":{}}}' | timeout 5 ./system-monitor | head -1

echo -e "\n✅ 快速测试完成！"
