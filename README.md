# 🖥️ System-MCP：系统监控 MCP 服务器

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![MCP Protocol](https://img.shields.io/badge/MCP-2024--11--05-green.svg)](https://modelcontextprotocol.io)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

一个功能完整、轻量级的系统监控 MCP 服务器，使用 Go 语言实现，为 AI 助手提供实时系统监控能力。

## ✨ 功能特性

### 🔧 监控工具
- **🖥️ CPU 监控** - 实时 CPU 使用率和核心状态
- **💾 内存监控** - 内存使用情况和交换空间状态  
- **📊 进程监控** - CPU/内存占用最高的进程列表
- **🌐 网络监控** - 网络接口状态和连接统计
- **💽 磁盘监控** - 磁盘使用情况和分区信息
- **📈 系统概览** - 系统整体状态和运行时间

### 🏗️ 技术特性
- ⚡ **零配置启动** - 无需任何参数即可运行
- 🔄 **实时数据** - 支持缓存和实时数据获取
- 📡 **标准协议** - 完整的 MCP 协议实现 (JSON-RPC 2.0)
- 🏃 **高性能** - 轻量级设计，资源占用极低
- 🔧 **易扩展** - 模块化架构，易于添加新监控工具

## 🚀 快速开始

### 前置要求
- Go 1.21 或更高版本
- macOS/Linux/Windows 系统

### 安装与编译

```bash
# 克隆仓库
git clone https://github.com/xirichuyi/System-Mcp.git
cd System-Mcp

# 安装依赖
go mod tidy

# 编译项目
go build -o system-monitor main.go
```

### 运行服务器

```bash
# 使用默认配置启动
./system-monitor

# 或者自定义配置
./system-monitor --name my-monitor --data-dir ./data
```

### 查看帮助

```bash
./system-monitor --help
```

## 🔧 配置指南

### Cursor IDE 配置

在 Cursor 中使用此 MCP 服务器，需要配置 `~/.cursor/mcp.json` 文件：

#### 方案一：极简配置（推荐）
```json
{
  "mcpServers": {
    "system-monitor": {
      "command": "/path/to/your/system-monitor"
    }
  }
}
```

#### 方案二：完整配置
```json
{
  "mcpServers": {
    "system-monitor": {
      "command": "/path/to/your/system-monitor",
      "args": [
        "--name", "system-monitor-mcp",
        "--data-dir", "/path/to/data",
        "--cache", "true"
      ],
      "env": {
        "PATH": "/usr/local/bin:/usr/bin:/bin"
      }
    }
  }
}
```

### 其他 MCP 客户端配置

项目根目录提供了多个配置文件模板：

- `mcp-simple.json` - 最简配置
- `cursor-mcp-config.json` - Cursor 专用配置示例

## 📖 使用方法

### 在 Cursor 中使用

配置完成后，您可以在 Cursor 中直接询问：

```
请帮我查看系统的 CPU 使用情况
显示内存占用最高的 10 个进程
检查磁盘空间使用情况
查看当前网络连接状态
```

### 命令行测试

您也可以通过命令行直接测试服务器功能：

#### 1. 初始化连接
```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}' | ./system-monitor
```

#### 2. 获取工具列表
```bash
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | ./system-monitor
```

#### 3. 调用 CPU 监控
```bash
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"cpu_info","arguments":{"duration":"1s"}}}' | ./system-monitor
```

#### 4. 查看内存使用
```bash
echo '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"memory_info","arguments":{}}}' | ./system-monitor
```

## 🛠️ 工具参数说明

### CPU 监控 (cpu_info)
```json
{
  "duration": "1s|5s|10s",    // 监控持续时间
  "use_cache": "true|false"   // 是否使用缓存
}
```

### 内存监控 (memory_info)
```json
{
  "use_cache": "true|false"   // 是否使用缓存
}
```

### 进程监控 (top_processes)
```json
{
  "limit": "10",              // 返回进程数量
  "sort_by": "cpu|memory",    // 排序方式
  "use_cache": "true|false"   // 是否使用缓存
}
```

### 网络监控 (network_stats)
```json
{
  "interface_filter": "",     // 网络接口过滤器
  "show_connections": "true|false", // 是否显示连接详情
  "use_cache": "true|false"   // 是否使用缓存
}
```

### 磁盘监控 (disk_info)
```json
{
  "show_all": "true|false",   // 是否显示所有分区
  "use_cache": "true|false"   // 是否使用缓存
}
```

### 系统概览 (system_overview)
```json
{
  "include_load": "true|false", // 是否包含负载信息
  "use_cache": "true|false"   // 是否使用缓存
}
```

## 📁 项目结构

```
System-Mcp/
├── main.go                    # 主程序入口
├── go.mod                     # Go 模块定义
├── go.sum                     # 依赖版本锁定
├── README.md                  # 项目说明文档
├── .gitignore                 # Git 忽略文件
├── mcp-simple.json           # 简化配置模板
├── internal/                  # 内部模块
│   ├── router/               # MCP 路由和协议处理
│   │   ├── router.go         # 主路由器
│   │   └── mcp_handler.go    # JSON-RPC 处理器
│   ├── tools/                # 监控工具实现
│   │   ├── cpu.go            # CPU 监控
│   │   ├── memory.go         # 内存监控
│   │   ├── process.go        # 进程监控
│   │   ├── network.go        # 网络监控
│   │   ├── disk.go           # 磁盘监控
│   │   └── system.go         # 系统概览
│   ├── storage/              # 数据存储
│   │   ├── json_store.go     # JSON 文件存储
│   │   └── cache.go          # 内存缓存
│   └── types/                # 类型定义
│       ├── mcp_types.go      # MCP 协议类型
│       └── monitor_types.go  # 监控数据类型
├── configs/                  # 配置文件
│   └── server_config.json    # 服务器配置
└── data/                     # 数据存储目录
```

## 🔧 开发指南

### 添加新的监控工具

1. 在 `internal/tools/` 目录创建新工具文件
2. 实现 `MonitorTool` 接口：
   ```go
   type MonitorTool interface {
       GetName() string
       GetDescription() string
       GetInputSchema() map[string]interface{}
       Execute(args map[string]interface{}) (string, error)
   }
   ```
3. 在 `router.go` 的 `InitializeTools()` 中注册新工具

### 自定义数据存储

项目支持自定义存储后端，只需实现 `DataStorage` 接口：

```go
type DataStorage interface {
    Save(key string, data interface{}) error
    Load(key string, data interface{}) error
}
```

## 🐛 故障排除

### 常见问题

1. **服务器无法启动**
   - 检查端口是否被占用
   - 确认有足够的系统权限

2. **Cursor 无法连接**
   - 检查配置文件路径是否正确
   - 确认可执行文件权限
   - 查看 Cursor 的错误日志

3. **监控数据不准确**
   - 尝试禁用缓存：`"use_cache": "false"`
   - 检查系统权限

### 调试模式

使用以下命令启用详细日志：

```bash
./system-monitor --help  # 查看所有可用参数
```

## 🤝 贡献指南

欢迎贡献代码！请遵循以下步骤：

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

- [Model Context Protocol](https://modelcontextprotocol.io) - 提供优秀的协议规范
- [gopsutil](https://github.com/shirou/gopsutil) - 跨平台系统信息库
- [Cursor](https://cursor.sh) - 优秀的 AI 编程环境

## 📞 联系方式

如有问题或建议，请通过以下方式联系：

- GitHub Issues: [提交问题](https://github.com/xirichuyi/System-Mcp/issues)
- 项目维护者: [@xirichuyi](https://github.com/xirichuyi)

---

⭐ 如果这个项目对您有帮助，请给个 Star 支持一下！