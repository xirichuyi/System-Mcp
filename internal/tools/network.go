package tools

import (
	"fmt"
	"time"

	"mcp-example/internal/types"

	"github.com/shirou/gopsutil/v3/net"
)

// NetworkTool 网络监控工具
type NetworkTool struct {
	cache types.Cache
}

// NewNetworkTool 创建新的网络监控工具
func NewNetworkTool(cache types.Cache) *NetworkTool {
	return &NetworkTool{
		cache: cache,
	}
}

// GetName 获取工具名称
func (nt *NetworkTool) GetName() string {
	return "network_stats"
}

// GetDescription 获取工具描述
func (nt *NetworkTool) GetDescription() string {
	return "获取网络连接状态和传输速度"
}

// GetInputSchema 获取输入模式
func (nt *NetworkTool) GetInputSchema() types.InputSchema {
	return types.InputSchema{
		Type: "object",
		Properties: map[string]types.Property{
			"show_connections": {
				Type:        "string",
				Description: "是否显示连接详情",
				Enum:        []string{"true", "false"},
				Default:     "false",
			},
			"interface_filter": {
				Type:        "string",
				Description: "网络接口过滤器（为空则显示所有）",
				Default:     "",
			},
			"use_cache": {
				Type:        "string",
				Description: "是否使用缓存数据",
				Enum:        []string{"true", "false"},
				Default:     "false",
			},
		},
	}
}

// Execute 执行网络监控
func (nt *NetworkTool) Execute(args map[string]interface{}) (string, error) {
	// 解析参数
	showConnStr, _ := args["show_connections"].(string)
	showConnections := showConnStr == "true"

	interfaceFilter, _ := args["interface_filter"].(string)

	useCacheStr, _ := args["use_cache"].(string)
	useCache := useCacheStr == "true"

	// 检查缓存
	cacheKey := fmt.Sprintf("network_stats_%t_%s", showConnections, interfaceFilter)
	if useCache {
		if cachedData, found := nt.cache.Get(cacheKey); found {
			if netInfo, ok := cachedData.(types.NetworkInfo); ok {
				return nt.formatNetworkInfo(netInfo, showConnections), nil
			}
		}
	}

	// 获取网络信息
	netInfo, err := nt.getNetworkInfo(showConnections, interfaceFilter)
	if err != nil {
		return "", fmt.Errorf("获取网络信息失败: %v", err)
	}

	// 缓存结果（缓存10秒）
	nt.cache.Set(cacheKey, netInfo, 10*time.Second)

	return nt.formatNetworkInfo(netInfo, showConnections), nil
}

// getNetworkInfo 获取网络信息
func (nt *NetworkTool) getNetworkInfo(showConnections bool, interfaceFilter string) (types.NetworkInfo, error) {
	var netInfo types.NetworkInfo

	// 获取网络接口统计
	netStats, err := net.IOCounters(true)
	if err != nil {
		return netInfo, fmt.Errorf("获取网络接口统计失败: %v", err)
	}

	// 过滤网络接口
	var filteredStats []net.IOCountersStat
	for _, stat := range netStats {
		// 跳过回环接口
		if stat.Name == "lo" || stat.Name == "lo0" {
			continue
		}

		// 应用接口过滤器
		if interfaceFilter != "" && stat.Name != interfaceFilter {
			continue
		}

		filteredStats = append(filteredStats, stat)
	}

	// 转换为内部类型
	for _, stat := range filteredStats {
		netInterface := types.NetworkInterface{
			Name:        stat.Name,
			BytesSent:   stat.BytesSent,
			BytesRecv:   stat.BytesRecv,
			PacketsSent: stat.PacketsSent,
			PacketsRecv: stat.PacketsRecv,
			ErrorsIn:    stat.Errin,
			ErrorsOut:   stat.Errout,
			DropIn:      stat.Dropin,
			DropOut:     stat.Dropout,
		}
		netInfo.Interfaces = append(netInfo.Interfaces, netInterface)
	}

	// 获取网络连接信息
	if showConnections {
		connections, err := net.Connections("all")
		if err == nil {
			netInfo.Connections = nt.processConnections(connections)
		}
	}

	netInfo.LastUpdated = time.Now()

	return netInfo, nil
}

// processConnections 处理网络连接信息
func (nt *NetworkTool) processConnections(connections []net.ConnectionStat) types.NetworkConnections {
	var netConn types.NetworkConnections

	netConn.Total = len(connections)
	netConn.ByStatus = make(map[string]int)
	netConn.ByProtocol = make(map[string]int)

	for _, conn := range connections {
		// 按状态统计
		netConn.ByStatus[conn.Status]++

		// 按协议统计
		protocol := fmt.Sprintf("%d-%d", conn.Type, conn.Family)
		netConn.ByProtocol[protocol]++

		// 添加连接详情（限制数量避免输出过多）
		if len(netConn.Details) < 20 {
			detail := types.ConnectionDetail{
				Protocol:   protocol,
				LocalIP:    conn.Laddr.IP,
				LocalPort:  conn.Laddr.Port,
				RemoteIP:   conn.Raddr.IP,
				RemotePort: conn.Raddr.Port,
				Status:     conn.Status,
				PID:        conn.Pid,
			}
			netConn.Details = append(netConn.Details, detail)
		}
	}

	return netConn
}

// formatNetworkInfo 格式化网络信息输出
func (nt *NetworkTool) formatNetworkInfo(netInfo types.NetworkInfo, showConnections bool) string {
	var result string

	result += "🌐 网络状态\n"
	result += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"

	// 网络接口统计
	if len(netInfo.Interfaces) > 0 {
		result += "网络接口统计:\n"
		result += fmt.Sprintf("%-15s %-12s %-12s %-12s %-12s %-8s %-8s\n",
			"接口", "发送(MB)", "接收(MB)", "发送包数", "接收包数", "发送错误", "接收错误")
		result += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"

		for _, iface := range netInfo.Interfaces {
			result += fmt.Sprintf("%-15s %-12.2f %-12.2f %-12d %-12d %-8d %-8d\n",
				iface.Name,
				float64(iface.BytesSent)/(1024*1024),
				float64(iface.BytesRecv)/(1024*1024),
				iface.PacketsSent,
				iface.PacketsRecv,
				iface.ErrorsOut,
				iface.ErrorsIn,
			)
		}
	}

	// 网络连接统计
	if showConnections && netInfo.Connections.Total > 0 {
		result += "\n🔗 网络连接统计:\n"
		result += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
		result += fmt.Sprintf("总连接数: %d\n", netInfo.Connections.Total)

		if len(netInfo.Connections.ByStatus) > 0 {
			result += "\n按状态分类:\n"
			for status, count := range netInfo.Connections.ByStatus {
				result += fmt.Sprintf("  %s: %d\n", status, count)
			}
		}

		if len(netInfo.Connections.ByProtocol) > 0 {
			result += "\n按协议分类:\n"
			for protocol, count := range netInfo.Connections.ByProtocol {
				result += fmt.Sprintf("  %s: %d\n", protocol, count)
			}
		}

		// 显示部分连接详情
		if len(netInfo.Connections.Details) > 0 {
			result += "\n连接详情 (前20个):\n"
			result += fmt.Sprintf("%-10s %-15s %-6s %-15s %-6s %-12s\n",
				"协议", "本地IP", "端口", "远程IP", "端口", "状态")
			result += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"

			for _, detail := range netInfo.Connections.Details {
				result += fmt.Sprintf("%-10s %-15s %-6d %-15s %-6d %-12s\n",
					detail.Protocol,
					detail.LocalIP,
					detail.LocalPort,
					detail.RemoteIP,
					detail.RemotePort,
					detail.Status,
				)
			}
		}
	}

	result += fmt.Sprintf("\n📅 更新时间: %s\n", netInfo.LastUpdated.Format("2006-01-02 15:04:05"))

	return result
}

// GetNetworkData 获取网络数据（供其他组件使用）
func (nt *NetworkTool) GetNetworkData(showConnections bool, interfaceFilter string) (types.NetworkInfo, error) {
	return nt.getNetworkInfo(showConnections, interfaceFilter)
}

// GetNetworkSpeed 计算网络传输速度（需要两次采样）
func (nt *NetworkTool) GetNetworkSpeed(interfaceName string, interval time.Duration) (float64, float64, error) {
	// 第一次采样
	stats1, err := net.IOCounters(true)
	if err != nil {
		return 0, 0, fmt.Errorf("获取第一次网络统计失败: %v", err)
	}

	var stat1 *net.IOCountersStat
	for _, stat := range stats1 {
		if stat.Name == interfaceName {
			stat1 = &stat
			break
		}
	}

	if stat1 == nil {
		return 0, 0, fmt.Errorf("找不到网络接口: %s", interfaceName)
	}

	// 等待间隔
	time.Sleep(interval)

	// 第二次采样
	stats2, err := net.IOCounters(true)
	if err != nil {
		return 0, 0, fmt.Errorf("获取第二次网络统计失败: %v", err)
	}

	var stat2 *net.IOCountersStat
	for _, stat := range stats2 {
		if stat.Name == interfaceName {
			stat2 = &stat
			break
		}
	}

	if stat2 == nil {
		return 0, 0, fmt.Errorf("找不到网络接口: %s", interfaceName)
	}

	// 计算速度 (bytes/second)
	seconds := interval.Seconds()
	uploadSpeed := float64(stat2.BytesSent-stat1.BytesSent) / seconds
	downloadSpeed := float64(stat2.BytesRecv-stat1.BytesRecv) / seconds

	return uploadSpeed, downloadSpeed, nil
}
