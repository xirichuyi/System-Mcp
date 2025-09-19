package tools

import (
	"fmt"
	"time"

	"mcp-example/internal/types"

	"github.com/shirou/gopsutil/v3/mem"
)

// MemoryTool 内存监控工具
type MemoryTool struct {
	cache types.Cache
}

// NewMemoryTool 创建新的内存监控工具
func NewMemoryTool(cache types.Cache) *MemoryTool {
	return &MemoryTool{
		cache: cache,
	}
}

// GetName 获取工具名称
func (mt *MemoryTool) GetName() string {
	return "memory_info"
}

// GetDescription 获取工具描述
func (mt *MemoryTool) GetDescription() string {
	return "获取内存使用情况详细信息"
}

// GetInputSchema 获取输入模式
func (mt *MemoryTool) GetInputSchema() types.InputSchema {
	return types.InputSchema{
		Type: "object",
		Properties: map[string]types.Property{
			"use_cache": {
				Type:        "string",
				Description: "是否使用缓存数据",
				Enum:        []string{"true", "false"},
				Default:     "false",
			},
		},
	}
}

// Execute 执行内存监控
func (mt *MemoryTool) Execute(args map[string]interface{}) (string, error) {
	// 解析参数
	useCacheStr, _ := args["use_cache"].(string)
	useCache := useCacheStr == "true"

	// 检查缓存
	cacheKey := "memory_info"
	if useCache {
		if cachedData, found := mt.cache.Get(cacheKey); found {
			if memInfo, ok := cachedData.(types.MemoryInfo); ok {
				return mt.formatMemoryInfo(memInfo), nil
			}
		}
	}

	// 获取内存信息
	memInfo, err := mt.getMemoryInfo()
	if err != nil {
		return "", fmt.Errorf("获取内存信息失败: %v", err)
	}

	// 缓存结果（缓存15秒）
	mt.cache.Set(cacheKey, memInfo, 15*time.Second)

	return mt.formatMemoryInfo(memInfo), nil
}

// getMemoryInfo 获取内存信息
func (mt *MemoryTool) getMemoryInfo() (types.MemoryInfo, error) {
	var memInfo types.MemoryInfo

	// 获取虚拟内存信息
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return memInfo, fmt.Errorf("获取虚拟内存信息失败: %v", err)
	}

	// 获取交换内存信息
	swapStat, err := mem.SwapMemory()
	if err != nil {
		return memInfo, fmt.Errorf("获取交换内存信息失败: %v", err)
	}

	// 填充内存信息
	memInfo.Total = vmStat.Total
	memInfo.Used = vmStat.Used
	memInfo.Available = vmStat.Available
	memInfo.Free = vmStat.Free
	memInfo.Buffers = vmStat.Buffers
	memInfo.Cached = vmStat.Cached
	memInfo.UsedPercent = vmStat.UsedPercent

	// 填充交换内存信息
	memInfo.Swap.Total = swapStat.Total
	memInfo.Swap.Used = swapStat.Used
	memInfo.Swap.Free = swapStat.Free
	memInfo.Swap.UsedPercent = swapStat.UsedPercent

	memInfo.LastUpdated = time.Now()

	return memInfo, nil
}

// formatMemoryInfo 格式化内存信息输出
func (mt *MemoryTool) formatMemoryInfo(memInfo types.MemoryInfo) string {
	var result string

	result += "💾 内存信息\n"
	result += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	result += fmt.Sprintf("总内存: %s\n", formatBytes(memInfo.Total))
	result += fmt.Sprintf("已使用: %s (%.2f%%)\n", formatBytes(memInfo.Used), memInfo.UsedPercent)
	result += fmt.Sprintf("可用内存: %s\n", formatBytes(memInfo.Available))
	result += fmt.Sprintf("空闲内存: %s\n", formatBytes(memInfo.Free))
	result += fmt.Sprintf("缓冲区: %s\n", formatBytes(memInfo.Buffers))
	result += fmt.Sprintf("缓存: %s\n", formatBytes(memInfo.Cached))

	result += "\n🔄 交换内存\n"
	result += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	result += fmt.Sprintf("总交换: %s\n", formatBytes(memInfo.Swap.Total))
	result += fmt.Sprintf("已使用: %s (%.2f%%)\n", formatBytes(memInfo.Swap.Used), memInfo.Swap.UsedPercent)
	result += fmt.Sprintf("空闲交换: %s\n", formatBytes(memInfo.Swap.Free))

	result += fmt.Sprintf("\n📅 更新时间: %s\n", memInfo.LastUpdated.Format("2006-01-02 15:04:05"))

	return result
}

// GetMemoryData 获取内存数据（供其他组件使用）
func (mt *MemoryTool) GetMemoryData() (types.MemoryInfo, error) {
	return mt.getMemoryInfo()
}

// formatBytes 格式化字节数
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
