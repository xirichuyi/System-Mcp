package tools

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"mcp-example/internal/types"

	"github.com/shirou/gopsutil/v3/process"
)

// ProcessTool 进程监控工具
type ProcessTool struct {
	cache types.Cache
}

// NewProcessTool 创建新的进程监控工具
func NewProcessTool(cache types.Cache) *ProcessTool {
	return &ProcessTool{
		cache: cache,
	}
}

// GetName 获取工具名称
func (pt *ProcessTool) GetName() string {
	return "top_processes"
}

// GetDescription 获取工具描述
func (pt *ProcessTool) GetDescription() string {
	return "获取 CPU 或内存占用最高的进程"
}

// GetInputSchema 获取输入模式
func (pt *ProcessTool) GetInputSchema() types.InputSchema {
	return types.InputSchema{
		Type: "object",
		Properties: map[string]types.Property{
			"sort_by": {
				Type:        "string",
				Description: "排序方式: cpu 或 memory",
				Enum:        []string{"cpu", "memory"},
				Default:     "memory",
			},
			"limit": {
				Type:        "string",
				Description: "返回进程数量",
				Default:     "10",
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

// Execute 执行进程监控
func (pt *ProcessTool) Execute(args map[string]interface{}) (string, error) {
	// 解析参数
	sortBy, _ := args["sort_by"].(string)
	if sortBy == "" {
		sortBy = "memory"
	}

	limitStr, _ := args["limit"].(string)
	if limitStr == "" {
		limitStr = "10"
	}
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	useCacheStr, _ := args["use_cache"].(string)
	useCache := useCacheStr == "true"

	// 检查缓存
	cacheKey := fmt.Sprintf("top_processes_%s_%d", sortBy, limit)
	if useCache {
		if cachedData, found := pt.cache.Get(cacheKey); found {
			if processList, ok := cachedData.(types.ProcessList); ok {
				return pt.formatProcessList(processList, sortBy, limit), nil
			}
		}
	}

	// 获取进程信息
	processList, err := pt.getTopProcesses(sortBy, limit)
	if err != nil {
		return "", fmt.Errorf("获取进程信息失败: %v", err)
	}

	// 缓存结果（缓存20秒）
	pt.cache.Set(cacheKey, processList, 20*time.Second)

	return pt.formatProcessList(processList, sortBy, limit), nil
}

// getTopProcesses 获取进程信息
func (pt *ProcessTool) getTopProcesses(sortBy string, limit int) (types.ProcessList, error) {
	var processList types.ProcessList

	// 获取所有进程
	processes, err := process.Processes()
	if err != nil {
		return processList, fmt.Errorf("获取进程列表失败: %v", err)
	}

	var procInfos []types.ProcessInfo
	for _, p := range processes {
		name, err := p.Name()
		if err != nil || name == "" {
			continue
		}

		// 获取进程信息
		memInfo, _ := p.MemoryInfo()
		cpuPercent, _ := p.CPUPercent()
		statusSlice, _ := p.Status()
		status := ""
		if len(statusSlice) > 0 {
			status = statusSlice[0]
		}
		createTime, _ := p.CreateTime()

		var memBytes uint64
		var memMB float64
		if memInfo != nil {
			memBytes = memInfo.RSS
			memMB = float64(memBytes) / (1024 * 1024)
		}

		procInfo := types.ProcessInfo{
			PID:         p.Pid,
			Name:        name,
			Status:      status,
			CPUPercent:  cpuPercent,
			MemoryBytes: memBytes,
			MemoryMB:    memMB,
			CreateTime:  createTime,
			LastUpdated: time.Now(),
		}

		procInfos = append(procInfos, procInfo)
	}

	// 排序
	if sortBy == "cpu" {
		sort.Slice(procInfos, func(i, j int) bool {
			return procInfos[i].CPUPercent > procInfos[j].CPUPercent
		})
	} else {
		sort.Slice(procInfos, func(i, j int) bool {
			return procInfos[i].MemoryBytes > procInfos[j].MemoryBytes
		})
	}

	// 限制数量
	if len(procInfos) > limit {
		procInfos = procInfos[:limit]
	}

	processList.Processes = procInfos
	processList.Total = len(processes)
	processList.LastUpdated = time.Now()

	return processList, nil
}

// formatProcessList 格式化进程列表输出
func (pt *ProcessTool) formatProcessList(processList types.ProcessList, sortBy string, limit int) string {
	var result string

	if sortBy == "cpu" {
		result += fmt.Sprintf("🚀 CPU 占用最高的 %d 个进程\n", limit)
	} else {
		result += fmt.Sprintf("💾 内存占用最高的 %d 个进程\n", limit)
	}
	result += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	result += fmt.Sprintf("%-8s %-25s %-10s %-12s %-10s\n", "PID", "进程名", "CPU%", "内存(MB)", "状态")
	result += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"

	for _, proc := range processList.Processes {
		// 截断过长的进程名
		name := proc.Name
		if len(name) > 25 {
			name = name[:22] + "..."
		}

		result += fmt.Sprintf("%-8d %-25s %-10.2f %-12.2f %-10s\n",
			proc.PID,
			name,
			proc.CPUPercent,
			proc.MemoryMB,
			proc.Status,
		)
	}

	result += fmt.Sprintf("\n📊 总进程数: %d\n", processList.Total)
	result += fmt.Sprintf("📅 更新时间: %s\n", processList.LastUpdated.Format("2006-01-02 15:04:05"))

	return result
}

// GetProcessData 获取进程数据（供其他组件使用）
func (pt *ProcessTool) GetProcessData(sortBy string, limit int) (types.ProcessList, error) {
	return pt.getTopProcesses(sortBy, limit)
}

// GetProcessByPID 根据 PID 获取特定进程信息
func (pt *ProcessTool) GetProcessByPID(pid int32) (types.ProcessInfo, error) {
	var procInfo types.ProcessInfo

	p, err := process.NewProcess(pid)
	if err != nil {
		return procInfo, fmt.Errorf("找不到 PID 为 %d 的进程: %v", pid, err)
	}

	name, err := p.Name()
	if err != nil {
		return procInfo, fmt.Errorf("获取进程名失败: %v", err)
	}

	memInfo, _ := p.MemoryInfo()
	cpuPercent, _ := p.CPUPercent()
	statusSlice, _ := p.Status()
	status := ""
	if len(statusSlice) > 0 {
		status = statusSlice[0]
	}
	createTime, _ := p.CreateTime()

	var memBytes uint64
	var memMB float64
	if memInfo != nil {
		memBytes = memInfo.RSS
		memMB = float64(memBytes) / (1024 * 1024)
	}

	procInfo = types.ProcessInfo{
		PID:         pid,
		Name:        name,
		Status:      status,
		CPUPercent:  cpuPercent,
		MemoryBytes: memBytes,
		MemoryMB:    memMB,
		CreateTime:  createTime,
		LastUpdated: time.Now(),
	}

	return procInfo, nil
}
