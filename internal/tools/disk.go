package tools

import (
	"fmt"
	"time"

	"mcp-example/internal/types"

	"github.com/shirou/gopsutil/v3/disk"
)

// DiskTool 磁盘监控工具
type DiskTool struct {
	cache types.Cache
}

// NewDiskTool 创建新的磁盘监控工具
func NewDiskTool(cache types.Cache) *DiskTool {
	return &DiskTool{
		cache: cache,
	}
}

// GetName 获取工具名称
func (dt *DiskTool) GetName() string {
	return "disk_info"
}

// GetDescription 获取工具描述
func (dt *DiskTool) GetDescription() string {
	return "获取磁盘使用情况"
}

// GetInputSchema 获取输入模式
func (dt *DiskTool) GetInputSchema() types.InputSchema {
	return types.InputSchema{
		Type: "object",
		Properties: map[string]types.Property{
			"show_all": {
				Type:        "string",
				Description: "是否显示所有分区（包括系统分区）",
				Enum:        []string{"true", "false"},
				Default:     "false",
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

// Execute 执行磁盘监控
func (dt *DiskTool) Execute(args map[string]interface{}) (string, error) {
	// 解析参数
	showAllStr, _ := args["show_all"].(string)
	showAll := showAllStr == "true"

	useCacheStr, _ := args["use_cache"].(string)
	useCache := useCacheStr == "true"

	// 检查缓存
	cacheKey := fmt.Sprintf("disk_info_%t", showAll)
	if useCache {
		if cachedData, found := dt.cache.Get(cacheKey); found {
			if diskInfo, ok := cachedData.(types.DiskInfo); ok {
				return dt.formatDiskInfo(diskInfo), nil
			}
		}
	}

	// 获取磁盘信息
	diskInfo, err := dt.getDiskInfo(showAll)
	if err != nil {
		return "", fmt.Errorf("获取磁盘信息失败: %v", err)
	}

	// 缓存结果（缓存30秒）
	dt.cache.Set(cacheKey, diskInfo, 30*time.Second)

	return dt.formatDiskInfo(diskInfo), nil
}

// getDiskInfo 获取磁盘信息
func (dt *DiskTool) getDiskInfo(showAll bool) (types.DiskInfo, error) {
	var diskInfo types.DiskInfo

	// 获取磁盘分区
	partitions, err := disk.Partitions(showAll)
	if err != nil {
		return diskInfo, fmt.Errorf("获取磁盘分区失败: %v", err)
	}

	for _, partition := range partitions {
		// 获取分区使用情况
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			// 跳过无法访问的分区
			continue
		}

		// 过滤一些不需要显示的分区
		if !showAll && dt.shouldSkipPartition(partition.Mountpoint, partition.Fstype) {
			continue
		}

		diskPartition := types.DiskPartition{
			Device:      partition.Device,
			Mountpoint:  partition.Mountpoint,
			Fstype:      partition.Fstype,
			Total:       usage.Total,
			Used:        usage.Used,
			Free:        usage.Free,
			UsedPercent: usage.UsedPercent,
		}

		diskInfo.Partitions = append(diskInfo.Partitions, diskPartition)
	}

	diskInfo.LastUpdated = time.Now()

	return diskInfo, nil
}

// shouldSkipPartition 判断是否应该跳过某个分区
func (dt *DiskTool) shouldSkipPartition(mountpoint, fstype string) bool {
	// 跳过一些系统分区和虚拟文件系统
	skipMountpoints := []string{
		"/dev", "/proc", "/sys", "/run", "/boot/efi",
		"/snap", "/var/snap", "/tmp", "/dev/shm",
	}

	skipFstypes := []string{
		"tmpfs", "devtmpfs", "sysfs", "proc", "devfs",
		"squashfs", "overlay", "aufs", "fuse",
	}

	// 检查挂载点
	for _, skip := range skipMountpoints {
		if mountpoint == skip {
			return true
		}
	}

	// 检查文件系统类型
	for _, skip := range skipFstypes {
		if fstype == skip {
			return true
		}
	}

	return false
}

// formatDiskInfo 格式化磁盘信息输出
func (dt *DiskTool) formatDiskInfo(diskInfo types.DiskInfo) string {
	var result string

	result += "💽 磁盘信息\n"
	result += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"

	if len(diskInfo.Partitions) == 0 {
		result += "未找到可用的磁盘分区\n"
	} else {
		result += fmt.Sprintf("%-20s %-10s %-12s %-12s %-12s %-10s\n",
			"挂载点", "文件系统", "总大小", "已使用", "可用", "使用率")
		result += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"

		var totalSize, totalUsed, totalFree uint64
		for _, partition := range diskInfo.Partitions {
			// 截断过长的挂载点
			mountpoint := partition.Mountpoint
			if len(mountpoint) > 20 {
				mountpoint = mountpoint[:17] + "..."
			}

			result += fmt.Sprintf("%-20s %-10s %-12s %-12s %-12s %-10.1f%%\n",
				mountpoint,
				partition.Fstype,
				formatBytes(partition.Total),
				formatBytes(partition.Used),
				formatBytes(partition.Free),
				partition.UsedPercent,
			)

			// 累计总计
			totalSize += partition.Total
			totalUsed += partition.Used
			totalFree += partition.Free
		}

		// 显示总计
		if len(diskInfo.Partitions) > 1 {
			result += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
			totalUsedPercent := float64(totalUsed) / float64(totalSize) * 100
			result += fmt.Sprintf("%-20s %-10s %-12s %-12s %-12s %-10.1f%%\n",
				"总计",
				"-",
				formatBytes(totalSize),
				formatBytes(totalUsed),
				formatBytes(totalFree),
				totalUsedPercent,
			)
		}
	}

	result += fmt.Sprintf("\n📅 更新时间: %s\n", diskInfo.LastUpdated.Format("2006-01-02 15:04:05"))

	return result
}

// GetDiskData 获取磁盘数据（供其他组件使用）
func (dt *DiskTool) GetDiskData(showAll bool) (types.DiskInfo, error) {
	return dt.getDiskInfo(showAll)
}

// GetDiskUsageByPath 获取指定路径的磁盘使用情况
func (dt *DiskTool) GetDiskUsageByPath(path string) (types.DiskPartition, error) {
	var partition types.DiskPartition

	usage, err := disk.Usage(path)
	if err != nil {
		return partition, fmt.Errorf("获取路径 %s 的磁盘使用情况失败: %v", path, err)
	}

	partition = types.DiskPartition{
		Device:      "unknown",
		Mountpoint:  path,
		Fstype:      "unknown",
		Total:       usage.Total,
		Used:        usage.Used,
		Free:        usage.Free,
		UsedPercent: usage.UsedPercent,
	}

	return partition, nil
}

// GetDiskIOStats 获取磁盘 I/O 统计信息
func (dt *DiskTool) GetDiskIOStats() (map[string]interface{}, error) {
	ioStats, err := disk.IOCounters()
	if err != nil {
		return nil, fmt.Errorf("获取磁盘 I/O 统计失败: %v", err)
	}

	result := make(map[string]interface{})
	for device, stat := range ioStats {
		result[device] = map[string]interface{}{
			"read_count":  stat.ReadCount,
			"write_count": stat.WriteCount,
			"read_bytes":  stat.ReadBytes,
			"write_bytes": stat.WriteBytes,
			"read_time":   stat.ReadTime,
			"write_time":  stat.WriteTime,
		}
	}

	return result, nil
}
