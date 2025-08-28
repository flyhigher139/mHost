package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/flyhigher139/mhost/internal/helper"
	"github.com/flyhigher139/mhost/pkg/logger"
)

const (
	// Version Helper Tool版本
	Version = "1.0.0"
	// ServiceName XPC服务名称
	ServiceName = "com.mhost.helper"
)

func main() {
	// 打印版本信息
	fmt.Printf("mHost Helper Tool v%s\n", Version)

	// 初始化日志
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting mHost Helper Tool...")

	// 创建增强日志器
	logger := logger.NewEnhancedLogger(logger.LogLevelInfo, false)

	// 创建Helper Tool实例
	helperTool, err := helper.NewHostsHelper(ServiceName, logger)
	if err != nil {
		log.Fatalf("Failed to create HostsHelper: %v", err)
	}

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动Helper Tool
	if err := helperTool.Start(); err != nil {
		log.Fatalf("Failed to start HostsHelper: %v", err)
	}

	log.Println("mHost Helper Tool started successfully")

	// 等待信号
	<-sigChan
	log.Println("Received shutdown signal")

	// 停止Helper Tool
	if err := helperTool.Stop(); err != nil {
		log.Printf("Error stopping HostsHelper: %v", err)
	}

	log.Println("mHost Helper Tool stopped")
}