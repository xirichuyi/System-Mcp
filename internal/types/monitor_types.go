package types

import "time"

// 监控数据相关类型定义

// 系统监控数据结构
type SystemInfo struct {
	Hostname      string    `json:"hostname"`
	OS            string    `json:"os"`
	Platform      string    `json:"platform"`
	KernelVersion string    `json:"kernel_version"`
	Architecture  string    `json:"architecture"`
	Uptime        uint64    `json:"uptime"`
	ProcessCount  uint64    `json:"process_count"`
	LastUpdated   time.Time `json:"last_updated"`
}

// CPU 监控数据
type CPUInfo struct {
	ModelName    string    `json:"model_name"`
	Cores        int32     `json:"cores"`
	LogicalCores int       `json:"logical_cores"`
	Frequency    float64   `json:"frequency_ghz"`
	Usage        CPUUsage  `json:"usage"`
	LastUpdated  time.Time `json:"last_updated"`
}

type CPUUsage struct {
	Total   float64   `json:"total_percent"`
	PerCore []float64 `json:"per_core_percent"`
}

// 内存监控数据
type MemoryInfo struct {
	Total       uint64    `json:"total_bytes"`
	Used        uint64    `json:"used_bytes"`
	Available   uint64    `json:"available_bytes"`
	Free        uint64    `json:"free_bytes"`
	Buffers     uint64    `json:"buffers_bytes"`
	Cached      uint64    `json:"cached_bytes"`
	UsedPercent float64   `json:"used_percent"`
	Swap        SwapInfo  `json:"swap"`
	LastUpdated time.Time `json:"last_updated"`
}

type SwapInfo struct {
	Total       uint64  `json:"total_bytes"`
	Used        uint64  `json:"used_bytes"`
	Free        uint64  `json:"free_bytes"`
	UsedPercent float64 `json:"used_percent"`
}

// 进程监控数据
type ProcessInfo struct {
	PID         int32     `json:"pid"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	CPUPercent  float64   `json:"cpu_percent"`
	MemoryBytes uint64    `json:"memory_bytes"`
	MemoryMB    float64   `json:"memory_mb"`
	CreateTime  int64     `json:"create_time"`
	LastUpdated time.Time `json:"last_updated"`
}

type ProcessList struct {
	Processes   []ProcessInfo `json:"processes"`
	Total       int           `json:"total_count"`
	LastUpdated time.Time     `json:"last_updated"`
}

// 网络监控数据
type NetworkInfo struct {
	Interfaces  []NetworkInterface `json:"interfaces"`
	Connections NetworkConnections `json:"connections"`
	LastUpdated time.Time          `json:"last_updated"`
}

type NetworkInterface struct {
	Name        string `json:"name"`
	BytesSent   uint64 `json:"bytes_sent"`
	BytesRecv   uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
	ErrorsIn    uint64 `json:"errors_in"`
	ErrorsOut   uint64 `json:"errors_out"`
	DropIn      uint64 `json:"drop_in"`
	DropOut     uint64 `json:"drop_out"`
}

type NetworkConnections struct {
	Total      int                `json:"total"`
	ByStatus   map[string]int     `json:"by_status"`
	ByProtocol map[string]int     `json:"by_protocol"`
	Details    []ConnectionDetail `json:"details,omitempty"`
}

type ConnectionDetail struct {
	Protocol   string `json:"protocol"`
	LocalIP    string `json:"local_ip"`
	LocalPort  uint32 `json:"local_port"`
	RemoteIP   string `json:"remote_ip"`
	RemotePort uint32 `json:"remote_port"`
	Status     string `json:"status"`
	PID        int32  `json:"pid"`
}

// 磁盘监控数据
type DiskInfo struct {
	Partitions  []DiskPartition `json:"partitions"`
	LastUpdated time.Time       `json:"last_updated"`
}

type DiskPartition struct {
	Device      string  `json:"device"`
	Mountpoint  string  `json:"mountpoint"`
	Fstype      string  `json:"fstype"`
	Total       uint64  `json:"total_bytes"`
	Used        uint64  `json:"used_bytes"`
	Free        uint64  `json:"free_bytes"`
	UsedPercent float64 `json:"used_percent"`
}

// 综合监控数据
type MonitorData struct {
	System    SystemInfo  `json:"system"`
	CPU       CPUInfo     `json:"cpu"`
	Memory    MemoryInfo  `json:"memory"`
	Network   NetworkInfo `json:"network"`
	Disk      DiskInfo    `json:"disk"`
	Processes ProcessList `json:"processes"`
	Timestamp time.Time   `json:"timestamp"`
}

// 工具接口定义
type MonitorTool interface {
	GetName() string
	GetDescription() string
	GetInputSchema() InputSchema
	Execute(args map[string]interface{}) (string, error)
}

// 数据存储接口
type DataStorage interface {
	Save(key string, data interface{}) error
	Load(key string, data interface{}) error
	Delete(key string) error
	Exists(key string) bool
}

// 缓存接口
type Cache interface {
	Set(key string, value interface{}, duration time.Duration) error
	Get(key string) (interface{}, bool)
	Delete(key string)
	Clear()
}
