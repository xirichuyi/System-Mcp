package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mcp-example/internal/router"
	"mcp-example/internal/storage"
)

const (
	DefaultServerName    = "system-monitor-mcp"
	DefaultServerVersion = "1.0.0"
	DefaultDataDir       = "data"
)

type ServerConfig struct {
	ServerName    string `json:"server_name"`
	ServerVersion string `json:"server_version"`
	DataDir       string `json:"data_dir"`
	CacheEnabled  bool   `json:"cache_enabled"`
}

func getDefaultConfig() *ServerConfig {
	return &ServerConfig{
		ServerName:    DefaultServerName,
		ServerVersion: DefaultServerVersion,
		DataDir:       DefaultDataDir,
		CacheEnabled:  true,
	}
}

func initializeStorage(config *ServerConfig) (*storage.JSONStorage, error) {
	// 初始化数据存储，输出到 stderr 避免干扰 JSON-RPC
	fmt.Fprintf(os.Stderr, "初始化数据存储，数据目录: %s\n", config.DataDir)

	if err := os.MkdirAll(config.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %v", err)
	}

	jsonStorage, err := storage.NewJSONStorage(config.DataDir)
	if err != nil {
		return nil, fmt.Errorf("初始化 JSON 存储失败: %v", err)
	}

	fmt.Fprintf(os.Stderr, "数据存储初始化成功\n")
	return jsonStorage, nil
}

func initializeCache(config *ServerConfig) *storage.MemoryCache {
	fmt.Fprintf(os.Stderr, "初始化内存缓存\n")
	cache := storage.NewMemoryCache()
	fmt.Fprintf(os.Stderr, "内存缓存初始化成功\n")
	return cache
}

func initializeRouter(config *ServerConfig, dataStorage *storage.JSONStorage, cache *storage.MemoryCache) (*router.Router, error) {
	fmt.Fprintf(os.Stderr, "初始化路由器，服务器: %s v%s\n", config.ServerName, config.ServerVersion)

	mcpRouter := router.NewRouter(config.ServerName, config.ServerVersion, dataStorage, cache)

	fmt.Fprintf(os.Stderr, "路由器初始化成功\n")
	return mcpRouter, nil
}

func setupSignalHandling(mcpRouter *router.Router) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		fmt.Fprintf(os.Stderr, "接收到信号: %v\n", sig)
		fmt.Fprintf(os.Stderr, "正在关闭服务器...\n")

		mcpRouter.Stop()

		fmt.Fprintf(os.Stderr, "服务器已关闭\n")
		os.Exit(0)
	}()
}

func parseFlags() *ServerConfig {
	config := getDefaultConfig()

	flag.StringVar(&config.ServerName, "name", config.ServerName, "服务器名称")
	flag.StringVar(&config.ServerVersion, "version", config.ServerVersion, "服务器版本")
	flag.StringVar(&config.DataDir, "data-dir", config.DataDir, "数据目录")
	flag.BoolVar(&config.CacheEnabled, "cache", config.CacheEnabled, "启用缓存")

	help := flag.Bool("help", false, "显示帮助信息")
	version := flag.Bool("v", false, "显示版本信息")

	flag.Parse()

	if *help {
		fmt.Printf("系统监控 MCP 服务器 v%s\n\n", config.ServerVersion)
		fmt.Println("💡 零配置启动：直接运行即可，无需任何参数！")
		fmt.Println("\n用法:")
		fmt.Printf("  %s                    # 使用默认配置启动\n", os.Args[0])
		fmt.Printf("  %s --name my-monitor  # 自定义服务器名称\n\n", os.Args[0])
		fmt.Println("可选参数:")
		flag.PrintDefaults()
		fmt.Println("\n支持的监控工具:")
		fmt.Println("  • cpu_info      - CPU 使用率和详细信息")
		fmt.Println("  • memory_info   - 内存使用情况")
		fmt.Println("  • top_processes - CPU/内存占用最高的进程")
		fmt.Println("  • network_stats - 网络连接状态和传输速度")
		fmt.Println("  • disk_info     - 磁盘使用情况")
		fmt.Println("  • system_overview - 系统综合概览")
		os.Exit(0)
	}

	if *version {
		fmt.Printf("%s v%s\n", config.ServerName, config.ServerVersion)
		os.Exit(0)
	}

	return config
}

func main() {
	// 将日志输出到 stderr，避免干扰 stdout 的 JSON-RPC 通信
	log.SetOutput(os.Stderr)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	config := parseFlags()

	// 打印启动信息到 stderr，避免干扰 stdout 的 JSON 通信
	fmt.Fprintf(os.Stderr, "\n🖥️  系统监控 MCP 服务器\n")
	fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Fprintf(os.Stderr, "服务器名称: %s\n", config.ServerName)
	fmt.Fprintf(os.Stderr, "服务器版本: %s\n", config.ServerVersion)
	fmt.Fprintf(os.Stderr, "数据目录: %s\n", config.DataDir)
	fmt.Fprintf(os.Stderr, "启动时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Fprintf(os.Stderr, "🚀 服务器启动中...\n\n")

	dataStorage, err := initializeStorage(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化数据存储失败: %v\n", err)
		os.Exit(1)
	}

	cache := initializeCache(config)

	mcpRouter, err := initializeRouter(config, dataStorage, cache)
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化路由器失败: %v\n", err)
		os.Exit(1)
	}

	setupSignalHandling(mcpRouter)

	fmt.Fprintf(os.Stderr, "✅ 服务器初始化完成，开始处理 MCP 请求...\n")
	fmt.Fprintf(os.Stderr, "✅ 服务器已启动，等待 MCP 客户端连接...\n")
	fmt.Fprintf(os.Stderr, "💡 使用 Ctrl+C 停止服务器\n\n")

	if err := mcpRouter.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "启动路由器失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "服务器正常退出\n")
}
