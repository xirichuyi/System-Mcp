package router

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"mcp-example/internal/storage"
	"mcp-example/internal/tools"
	"mcp-example/internal/types"
)

// Router MCP 路由器
type Router struct {
	handler *MCPHandler
	storage types.DataStorage
	cache   types.Cache
	running bool
	input   io.Reader
	output  io.Writer
}

// NewRouter 创建新的路由器
func NewRouter(serverName, serverVersion string, dataStorage types.DataStorage, cache types.Cache) *Router {
	return &Router{
		handler: NewMCPHandler(serverName, serverVersion),
		storage: dataStorage,
		cache:   cache,
		input:   os.Stdin,
		output:  os.Stdout,
	}
}

// SetIO 设置输入输出流（用于测试）
func (r *Router) SetIO(input io.Reader, output io.Writer) {
	r.input = input
	r.output = output
}

// InitializeTools 初始化所有监控工具
func (r *Router) InitializeTools() error {
	// 初始化监控工具，但不输出日志避免干扰 JSON-RPC

	// 创建工具实例
	cpuTool := tools.NewCPUTool(r.cache)
	memoryTool := tools.NewMemoryTool(r.cache)
	processTool := tools.NewProcessTool(r.cache)
	networkTool := tools.NewNetworkTool(r.cache)
	diskTool := tools.NewDiskTool(r.cache)
	systemTool := tools.NewSystemTool(r.cache)

	// 注册工具
	r.handler.RegisterTool(cpuTool)
	r.handler.RegisterTool(memoryTool)
	r.handler.RegisterTool(processTool)
	r.handler.RegisterTool(networkTool)
	r.handler.RegisterTool(diskTool)
	r.handler.RegisterTool(systemTool)

	// 工具初始化完成，但不输出日志避免干扰 JSON-RPC

	return nil
}

// Start 启动路由器
func (r *Router) Start() error {
	if r.running {
		return fmt.Errorf("路由器已经在运行")
	}

	// 启动 MCP 路由器，但不输出日志避免干扰 JSON-RPC
	r.running = true

	// 初始化工具
	if err := r.InitializeTools(); err != nil {
		return fmt.Errorf("初始化工具失败: %v", err)
	}

	// 启动消息处理循环
	return r.messageLoop()
}

// Stop 停止路由器
func (r *Router) Stop() {
	// 停止 MCP 路由器，但不输出日志避免干扰 JSON-RPC
	r.running = false
}

// messageLoop 消息处理循环
func (r *Router) messageLoop() error {
	scanner := bufio.NewScanner(r.input)

	// 不输出到 stdout，避免干扰 JSON-RPC 通信

	for r.running && scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// 解析 JSON-RPC 请求
		var req types.JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			// 解析失败，但不输出日志到避免干扰 JSON-RPC
			// 发送解析错误响应（只有在有ID的情况下）
			var rawMessage map[string]interface{}
			json.Unmarshal([]byte(line), &rawMessage)
			if id, hasID := rawMessage["id"]; hasID {
				errorResp := &types.JSONRPCResponse{
					JSONRPC: "2.0",
					ID:      id,
					Error: &types.RPCError{
						Code:    -32700,
						Message: "Parse error: " + err.Error(),
					},
				}
				r.sendResponse(errorResp)
			}
			continue
		}

		// 检查是否是通知（没有 ID 字段）
		isNotification := req.ID == nil

		// 处理请求
		response := r.handler.HandleRequest(&req)

		// 发送响应（只有非通知的请求才发送响应）
		if response != nil && !isNotification {
			r.sendResponse(response)
		}
	}

	if err := scanner.Err(); err != nil {
		// 扫描错误，但不输出到 stdout
		return fmt.Errorf("扫描输入时出错: %v", err)
	}

	return nil
}

// sendResponse 发送响应
func (r *Router) sendResponse(response *types.JSONRPCResponse) {
	respBytes, err := json.Marshal(response)
	if err != nil {
		// 序列化失败，但不输出日志避免干扰 JSON-RPC
		return
	}

	if _, err := fmt.Fprintln(r.output, string(respBytes)); err != nil {
		// 发送失败，但不输出日志避免干扰 JSON-RPC
	}
}

// ProcessSingleRequest 处理单个请求（用于测试）
func (r *Router) ProcessSingleRequest(reqJSON string) (string, error) {
	var req types.JSONRPCRequest
	if err := json.Unmarshal([]byte(reqJSON), &req); err != nil {
		return "", fmt.Errorf("解析请求失败: %v", err)
	}

	response := r.handler.HandleRequest(&req)
	if response == nil {
		return "", nil
	}

	respBytes, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("序列化响应失败: %v", err)
	}

	return string(respBytes), nil
}

// SaveMonitorData 保存监控数据到存储
func (r *Router) SaveMonitorData(key string, data interface{}) error {
	if r.storage == nil {
		return fmt.Errorf("存储未初始化")
	}

	return r.storage.Save(key, data)
}

// LoadMonitorData 从存储加载监控数据
func (r *Router) LoadMonitorData(key string, data interface{}) error {
	if r.storage == nil {
		return fmt.Errorf("存储未初始化")
	}

	return r.storage.Load(key, data)
}

// GetCacheStats 获取缓存统计信息
func (r *Router) GetCacheStats() map[string]interface{} {
	stats := make(map[string]interface{})

	if memCache, ok := r.cache.(*storage.MemoryCache); ok {
		stats["size"] = memCache.Size()
		stats["keys"] = memCache.Keys()
	}

	return stats
}

// GetStorageStats 获取存储统计信息
func (r *Router) GetStorageStats() map[string]interface{} {
	stats := make(map[string]interface{})

	if jsonStorage, ok := r.storage.(*storage.JSONStorage); ok {
		stats["data_dir"] = jsonStorage.GetDataDir()
		if keys, err := jsonStorage.ListKeys(); err == nil {
			stats["keys"] = keys
			stats["count"] = len(keys)
		}
	}

	return stats
}

// IsRunning 检查路由器是否正在运行
func (r *Router) IsRunning() bool {
	return r.running
}

// GetHandler 获取 MCP 处理器（用于测试）
func (r *Router) GetHandler() *MCPHandler {
	return r.handler
}
