# 我的第一个 MCP 项目开发记录

最近花了几天时间学习 MCP（Model Context Protocol），并且成功搭建了一个系统监控服务器。记录一下这个过程中的一些经历和踩过的坑，希望对后来者有帮助。

## 刚开始的困惑

第一次接触 MCP 时完全不知道从哪里入手：
- 看文档说是个协议，但不知道具体怎么实现
- 不确定是要启动一个 HTTP 服务器，还是别的什么
- Cursor 怎么知道调用我的工具？

折腾了一阵子才搞明白，MCP 其实是通过 stdin/stdout 进行通信的，不是 HTTP。就是两个进程通过管道传递 JSON 消息。

## 项目结构的演变

一开始想偷懒，把所有代码都塞在一个 main.go 文件里。写到一半发现完全维护不了，只能重构。

最后搞了个分层的结构：
- `main.go` - 程序入口
- `internal/router/` - 处理 JSON-RPC 协议
- `internal/tools/` - 各种监控工具
- `internal/storage/` - 数据存储和缓存
- `internal/types/` - 类型定义

这样每个模块职责清楚，后面添加新功能也方便。

## 踩过的大坑

### 坑1：stdout 被污染了

这个坑卡了我好久。一开始程序启动时会打印一些日志，类似这样：

```
🖥️ 系统监控 MCP 服务器启动中...
{"jsonrpc":"2.0","id":1,"result":{...}}
```

结果 Cursor 一直报错说解析 JSON 失败。后来才明白，MCP 要求 stdout 只能输出 JSON-RPC 消息，其他日志都要输出到 stderr。

解决方法就是把所有日志都改成 `fmt.Fprintf(os.Stderr, ...)`。

### 坑2：通知不需要响应

JSON-RPC 有两种消息：请求和通知。请求有 ID 字段，需要返回响应；通知没有 ID，不需要响应。

一开始我对所有消息都返回响应，结果 `notifications/initialized` 这个通知也被我回复了，导致协议出错。

后来加了个判断：
```go
isNotification := req.ID == nil
if response != nil && !isNotification {
    r.sendResponse(response)
}
```

### 坑3：库函数返回类型不对

用 `gopsutil` 库获取进程状态时踩了个坑：

```go
// 我以为 Status() 返回 string，结果是 []string
status, _ := p.Status()  // 编译错误

// 正确的写法
statusSlice, _ := p.Status()
status := ""
if len(statusSlice) > 0 {
    status = statusSlice[0]
}
```

这种类型错误编译器会报错，但如果不仔细看文档很容易踩坑。

## 配置文件的优化

最开始配置文件特别复杂，需要指定一堆参数：

```json
{
  "mcpServers": {
    "system-monitor": {
      "command": "/Users/ml/Work/Go/Mcp/system-monitor",
      "args": ["--name", "system-monitor-mcp", "--version", "1.0.0", "--data-dir", "/Users/ml/Work/Go/Mcp/data", "--cache", "true"],
      "env": {"PATH": "/usr/local/bin:/usr/bin:/bin"}
    }
  }
}
```

后来觉得这样对用户太不友好了，就在代码里加了默认值，配置文件简化成：

```json
{
  "mcpServers": {
    "system-monitor": {
      "command": "/Users/ml/Work/Go/Mcp/system-monitor"
    }
  }
}
```

这样用户只需要指定可执行文件路径就行了。

## 性能还不错

做完之后用 `ps` 看了一下，这个服务器还挺轻量的：
- 内存只用了 12.6 MB
- CPU 使用率基本是 0%
- 启动很快

因为大部分时间都在等待输入，所以资源占用很低。

另外还加了个简单的缓存，避免频繁调用系统 API：

```go
type MemoryCache struct {
    items map[string]*CacheItem
    mutex sync.RWMutex
}
```

缓存 5 分钟，对于监控数据来说够用了。

## 最终效果

配置好之后，在 Cursor 里直接问"帮我查看系统的 CPU 使用情况"，就能得到这样的回复：

```
🖥️ CPU 信息
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
型号: Apple M4
核心数: 10 物理核心, 10 逻辑核心
总体使用率: 18.98%

各核心使用率:
  核心 1: 42.42%
  核心 2: 35.64%
  ...
```

还能查看内存、磁盘、网络、进程等信息，基本的系统监控需求都能满足。

## 总结

整个过程下来，主要学到了：

1. **MCP 协议本质**：就是通过 stdin/stdout 传递 JSON-RPC 消息
2. **调试很重要**：stdout 不能有杂音，通知不需要响应
3. **用户体验**：配置越简单越好
4. **代码结构**：分层架构比单文件好维护

做完这个项目对 MCP 的理解深入了很多，也算是入门了。代码放在 GitHub 上了：[https://github.com/xirichuyi/System-Mcp](https://github.com/xirichuyi/System-Mcp)

如果你也想试试 MCP 开发，建议从一个简单的 echo 工具开始，不要一上来就搞复杂的功能。
