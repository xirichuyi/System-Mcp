# MCP Go 服务器示例

这是一个使用 Go 语言实现的 Model Context Protocol (MCP) 服务器示例。

## 功能特性

### 工具 (Tools)
- **echo**: 回显输入消息
- **current_time**: 获取当前系统时间
- **calculator**: 简单的数学计算器

### 资源 (Resources)
- **file://example.txt**: 示例文本文件资源

## 使用方法

### 编译
```bash
go build -o mcp-server main.go
```

### 运行
```bash
./mcp-server
```

### 测试
服务器通过 stdio 接收 JSON-RPC 2.0 格式的消息。

#### 初始化请求示例：
```json
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}
```

#### 列出工具：
```json
{"jsonrpc":"2.0","id":2,"method":"tools/list"}
```

#### 调用工具：
```json
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"echo","arguments":{"message":"Hello MCP!"}}}
```

## 项目结构

- `main.go`: 完整的 MCP 服务器实现
- `go.mod`: Go 模块定义
- `README.md`: 项目说明文档

