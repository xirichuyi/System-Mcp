package tools

import (
	"fmt"
	"time"

	"mcp-example/internal/types"

	"github.com/shirou/gopsutil/v3/host"
)

// SystemTool 系统信息工具
type SystemTool struct {
	cache types.Cache
}

// NewSystemTool 创建新的系统信息工具
func NewSystemTool(cache types.Cache) *SystemTool {
	return &SystemTool{
		cache: cache,
	}
}

// GetName 获取工具名称
func (st *SystemTool) GetName() string {
	return "system_overview"
}

// GetDescription 获取工具描述
func (st *SystemTool) GetDescription() string {
	return "获取系统综合概览信息"
}

// GetInputSchema 获取输入模式
func (st *SystemTool) GetInputSchema() types.InputSchema {
	return types.InputSchema{
		Type: "object",
		Properties: map[string]types.Property{
			"include_load": {
				Type:        "string",
				Description: "是否包含系统负载信息",
				Enum:        []string{"true", "false"},
				Default:     "true",
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

// Execute 执行系统信息获取
func (st *SystemTool) Execute(args map[string]interface{}) (string, error) {
	// 解析参数
	includeLoadStr, _ := args["include_load"].(string)
	includeLoad := includeLoadStr != "false" // 默认为 true

	useCacheStr, _ := args["use_cache"].(string)
	useCache := useCacheStr == "true"

	// 检查缓存
	cacheKey := fmt.Sprintf("system_overview_%t", includeLoad)
	if useCache {
		if cachedData, found := st.cache.Get(cacheKey); found {
			if sysInfo, ok := cachedData.(types.SystemInfo); ok {
				return st.formatSystemInfo(sysInfo, includeLoad), nil
			}
		}
	}

	// 获取系统信息
	sysInfo, err := st.getSystemInfo(includeLoad)
	if err != nil {
		return "", fmt.Errorf("获取系统信息失败: %v", err)
	}

	// 缓存结果（缓存60秒）
	st.cache.Set(cacheKey, sysInfo, 60*time.Second)

	return st.formatSystemInfo(sysInfo, includeLoad), nil
}

// getSystemInfo 获取系统信息
func (st *SystemTool) getSystemInfo(includeLoad bool) (types.SystemInfo, error) {
	var sysInfo types.SystemInfo

	// 获取主机信息
	hostInfo, err := host.Info()
	if err != nil {
		return sysInfo, fmt.Errorf("获取主机信息失败: %v", err)
	}

	// 填充系统信息
	sysInfo.Hostname = hostInfo.Hostname
	sysInfo.OS = hostInfo.OS
	sysInfo.Platform = hostInfo.Platform
	sysInfo.KernelVersion = hostInfo.KernelVersion
	sysInfo.Architecture = hostInfo.KernelArch
	sysInfo.Uptime = hostInfo.Uptime
	sysInfo.ProcessCount = hostInfo.Procs
	sysInfo.LastUpdated = time.Now()

	return sysInfo, nil
}

// formatSystemInfo 格式化系统信息输出
func (st *SystemTool) formatSystemInfo(sysInfo types.SystemInfo, includeLoad bool) string {
	var result string

	result += "🖥️  系统概览\n"
	result += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	result += fmt.Sprintf("主机名: %s\n", sysInfo.Hostname)
	result += fmt.Sprintf("操作系统: %s\n", sysInfo.OS)
	result += fmt.Sprintf("平台: %s\n", sysInfo.Platform)
	result += fmt.Sprintf("内核版本: %s\n", sysInfo.KernelVersion)
	result += fmt.Sprintf("架构: %s\n", sysInfo.Architecture)

	// 格式化运行时间
	uptime := time.Duration(sysInfo.Uptime) * time.Second
	days := int(uptime.Hours()) / 24
	hours := int(uptime.Hours()) % 24
	minutes := int(uptime.Minutes()) % 60
	result += fmt.Sprintf("运行时间: %d天 %d小时 %d分钟\n", days, hours, minutes)

	result += fmt.Sprintf("进程数: %d\n", sysInfo.ProcessCount)

	// 包含负载信息 (在某些系统上可能不可用)
	if includeLoad {
		// 注意：LoadAvg 在某些系统上可能不可用，这里暂时注释掉
		// 可以根据需要实现替代方案
		result += "\n📊 系统负载\n"
		result += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
		result += "系统负载信息在此平台暂不可用\n"
	}

	result += fmt.Sprintf("\n📅 更新时间: %s\n", sysInfo.LastUpdated.Format("2006-01-02 15:04:05"))

	return result
}

// GetSystemData 获取系统数据（供其他组件使用）
func (st *SystemTool) GetSystemData(includeLoad bool) (types.SystemInfo, error) {
	return st.getSystemInfo(includeLoad)
}

// GetBootTime 获取系统启动时间
func (st *SystemTool) GetBootTime() (time.Time, error) {
	bootTime, err := host.BootTime()
	if err != nil {
		return time.Time{}, fmt.Errorf("获取系统启动时间失败: %v", err)
	}

	return time.Unix(int64(bootTime), 0), nil
}

// GetSystemUsers 获取当前登录的用户
func (st *SystemTool) GetSystemUsers() ([]map[string]interface{}, error) {
	users, err := host.Users()
	if err != nil {
		return nil, fmt.Errorf("获取系统用户失败: %v", err)
	}

	var result []map[string]interface{}
	for _, user := range users {
		userInfo := map[string]interface{}{
			"user":     user.User,
			"terminal": user.Terminal,
			"host":     user.Host,
			"started":  time.Unix(int64(user.Started), 0).Format("2006-01-02 15:04:05"),
		}
		result = append(result, userInfo)
	}

	return result, nil
}

// GetSystemTemperature 获取系统温度信息
func (st *SystemTool) GetSystemTemperature() ([]map[string]interface{}, error) {
	temps, err := host.SensorsTemperatures()
	if err != nil {
		return nil, fmt.Errorf("获取系统温度失败: %v", err)
	}

	var result []map[string]interface{}
	for _, temp := range temps {
		tempInfo := map[string]interface{}{
			"sensor_key":  temp.SensorKey,
			"temperature": temp.Temperature,
			"high":        temp.High,
			"critical":    temp.Critical,
		}
		result = append(result, tempInfo)
	}

	return result, nil
}

// GetComprehensiveOverview 获取综合系统概览（包含所有监控数据）
func (st *SystemTool) GetComprehensiveOverview(
	cpuTool *CPUTool,
	memTool *MemoryTool,
	diskTool *DiskTool,
	netTool *NetworkTool,
) (types.MonitorData, error) {
	var monitorData types.MonitorData

	// 获取系统信息
	sysInfo, err := st.getSystemInfo(true)
	if err != nil {
		return monitorData, fmt.Errorf("获取系统信息失败: %v", err)
	}
	monitorData.System = sysInfo

	// 获取 CPU 信息
	if cpuTool != nil {
		cpuInfo, err := cpuTool.GetCPUData(time.Second)
		if err == nil {
			monitorData.CPU = cpuInfo
		}
	}

	// 获取内存信息
	if memTool != nil {
		memInfo, err := memTool.GetMemoryData()
		if err == nil {
			monitorData.Memory = memInfo
		}
	}

	// 获取磁盘信息
	if diskTool != nil {
		diskInfo, err := diskTool.GetDiskData(false)
		if err == nil {
			monitorData.Disk = diskInfo
		}
	}

	// 获取网络信息
	if netTool != nil {
		netInfo, err := netTool.GetNetworkData(false, "")
		if err == nil {
			monitorData.Network = netInfo
		}
	}

	monitorData.Timestamp = time.Now()

	return monitorData, nil
}
