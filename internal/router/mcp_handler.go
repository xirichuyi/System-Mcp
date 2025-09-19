package router

import (
	"encoding/json"

	"mcp-example/internal/types"
)

// MCPHandler MCP 协议处理器
type MCPHandler struct {
	serverName    string
	serverVersion string
	tools         map[string]types.MonitorTool
}

// NewMCPHandler 创建新的 MCP 处理器
func NewMCPHandler(serverName, serverVersion string) *MCPHandler {
	return &MCPHandler{
		serverName:    serverName,
		serverVersion: serverVersion,
		tools:         make(map[string]types.MonitorTool),
	}
}

// RegisterTool 注册工具
func (h *MCPHandler) RegisterTool(tool types.MonitorTool) {
	h.tools[tool.GetName()] = tool
	// 工具注册成功，但不输出日志避免干扰 JSON-RPC
}

// HandleRequest 处理 MCP 请求
func (h *MCPHandler) HandleRequest(req *types.JSONRPCRequest) *types.JSONRPCResponse {
	// 处理请求，但不输出日志避免干扰 JSON-RPC

	switch req.Method {
	case types.MethodInitialize:
		return h.handleInitialize(req)
	case types.MethodInitialized, types.MethodNotificationInitialized:
		return h.handleInitialized(req)
	case types.MethodListTools:
		return h.handleListTools(req)
	case types.MethodCallTool:
		return h.handleCallTool(req)
	case types.MethodListPrompts:
		return h.handleListPrompts(req)
	case types.MethodListResources:
		return h.handleListResources(req)
	case types.MethodReadResource:
		return h.handleReadResource(req)
	default:
		return h.errorResponse(req, -32601, "Method not found: "+req.Method)
	}
}

// handleInitialize 处理初始化请求
func (h *MCPHandler) handleInitialize(req *types.JSONRPCRequest) *types.JSONRPCResponse {
	// 初始化服务器，但不输出日志避免干扰 JSON-RPC

	result := types.InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: types.ServerCapabilities{
			Tools: &types.ToolsCapability{
				ListChanged: true,
			},
			Resources: &types.ResourcesCapability{
				Subscribe:   false,
				ListChanged: false,
			},
			Prompts: &types.PromptsCapability{
				ListChanged: false,
			},
		},
		ServerInfo: types.ServerInfo{
			Name:    h.serverName,
			Version: h.serverVersion,
		},
	}

	return &types.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

// handleInitialized 处理初始化完成通知
func (h *MCPHandler) handleInitialized(req *types.JSONRPCRequest) *types.JSONRPCResponse {
	// 服务器初始化完成，但不输出日志避免干扰 JSON-RPC
	// 初始化完成通知通常不需要返回响应
	return nil
}

// handleListTools 处理工具列表请求
func (h *MCPHandler) handleListTools(req *types.JSONRPCRequest) *types.JSONRPCResponse {
	// 列出工具，但不输出日志避免干扰 JSON-RPC

	var tools []types.Tool
	for _, tool := range h.tools {
		mcpTool := types.Tool{
			Name:        tool.GetName(),
			Description: tool.GetDescription(),
			InputSchema: tool.GetInputSchema(),
		}
		tools = append(tools, mcpTool)
	}

	result := map[string]interface{}{
		"tools": tools,
	}

	return &types.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

// handleCallTool 处理工具调用请求
func (h *MCPHandler) handleCallTool(req *types.JSONRPCRequest) *types.JSONRPCResponse {
	var params types.CallToolParams
	if req.Params != nil {
		paramBytes, err := json.Marshal(req.Params)
		if err != nil {
			return h.errorResponse(req, -32602, "Invalid params: "+err.Error())
		}
		if err := json.Unmarshal(paramBytes, &params); err != nil {
			return h.errorResponse(req, -32602, "Invalid params: "+err.Error())
		}
	}

	// 调用工具，但不输出日志避免干扰 JSON-RPC

	// 查找工具
	tool, exists := h.tools[params.Name]
	if !exists {
		return h.errorResponse(req, -32602, "Unknown tool: "+params.Name)
	}

	// 执行工具
	result, err := tool.Execute(params.Arguments)
	if err != nil {
		// 工具执行失败，但不输出日志避免干扰 JSON-RPC
		return &types.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: types.CallToolResult{
				Content: []types.Content{
					{Type: "text", Text: "❌ " + err.Error()},
				},
				IsError: true,
			},
		}
	}

	// 工具执行成功，但不输出日志避免干扰 JSON-RPC

	return &types.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: types.CallToolResult{
			Content: []types.Content{
				{Type: "text", Text: result},
			},
		},
	}
}

// handleListPrompts 处理提示列表请求
func (h *MCPHandler) handleListPrompts(req *types.JSONRPCRequest) *types.JSONRPCResponse {
	// 列出提示，但不输出日志避免干扰 JSON-RPC

	result := map[string]interface{}{
		"prompts": []interface{}{},
	}

	return &types.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

// handleListResources 处理资源列表请求
func (h *MCPHandler) handleListResources(req *types.JSONRPCRequest) *types.JSONRPCResponse {
	// 列出资源，但不输出日志避免干扰 JSON-RPC

	result := map[string]interface{}{
		"resources": []interface{}{},
	}

	return &types.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

// handleReadResource 处理资源读取请求
func (h *MCPHandler) handleReadResource(req *types.JSONRPCRequest) *types.JSONRPCResponse {
	return h.errorResponse(req, -32601, "Resource reading not implemented")
}

// errorResponse 创建错误响应
func (h *MCPHandler) errorResponse(req *types.JSONRPCRequest, code int, message string) *types.JSONRPCResponse {
	// 创建错误响应，但不输出日志避免干扰 JSON-RPC

	return &types.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Error: &types.RPCError{
			Code:    code,
			Message: message,
		},
	}
}

// GetRegisteredTools 获取已注册的工具列表
func (h *MCPHandler) GetRegisteredTools() []string {
	var toolNames []string
	for name := range h.tools {
		toolNames = append(toolNames, name)
	}
	return toolNames
}

// GetServerInfo 获取服务器信息
func (h *MCPHandler) GetServerInfo() (string, string) {
	return h.serverName, h.serverVersion
}
