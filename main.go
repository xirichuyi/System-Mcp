package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"mcp-example/internal/router"
	"mcp-example/internal/storage"
)

const (
	DefaultServerName    = "system-monitor-mcp"
	DefaultServerVersion = "1.0.0"
	DefaultDataDir       = "data"
)

type ServerConfig struct {
	ServerName    string
	ServerVersion string
	DataDir       string
	CacheEnabled  bool
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
	if err := os.MkdirAll(config.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %v", err)
	}

	jsonStorage, err := storage.NewJSONStorage(config.DataDir)
	if err != nil {
		return nil, fmt.Errorf("初始化存储失败: %v", err)
	}

	return jsonStorage, nil
}

func initializeCache() *storage.MemoryCache {
	return storage.NewMemoryCache()
}

func initializeRouter(config *ServerConfig, dataStorage *storage.JSONStorage, cache *storage.MemoryCache) *router.Router {
	return router.NewRouter(config.ServerName, config.ServerVersion, dataStorage, cache)
}

func setupSignalHandling(mcpRouter *router.Router) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		mcpRouter.Stop()
		os.Exit(0)
	}()
}

func parseFlags() *ServerConfig {
	config := getDefaultConfig()

	flag.StringVar(&config.ServerName, "name", config.ServerName, "服务器名称")
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
	log.SetOutput(os.Stderr)

	config := parseFlags()

	// 初始化组件
	dataStorage, err := initializeStorage(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "存储初始化失败: %v\n", err)
		os.Exit(1)
	}

	cache := initializeCache()
	mcpRouter := initializeRouter(config, dataStorage, cache)

	setupSignalHandling(mcpRouter)

	// 启动服务器
	if err := mcpRouter.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "服务器启动失败: %v\n", err)
		os.Exit(1)
	}
}
