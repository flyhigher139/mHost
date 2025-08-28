package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/flyhigher139/mhost/internal/helper"
	"github.com/flyhigher139/mhost/pkg/logger"
)

// main XPC客户端示例程序
func main() {
	fmt.Println("=== XPC Client Example ===")

	// 创建增强日志器
	logger := logger.NewEnhancedLogger(logger.LogLevelInfo, false)

	// 创建XPC客户端
	client := helper.NewXPCClient("com.mhost.helper", logger)

	// 连接到Helper Tool
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect to Helper Tool: %v", err)
	}
	defer client.Disconnect()

	fmt.Println("Connected to Helper Tool successfully")

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 测试获取状态
	fmt.Println("\n1. Testing GetStatus...")
	status, err := client.GetStatus(ctx)
	if err != nil {
		log.Printf("GetStatus failed: %v", err)
	} else {
		fmt.Printf("Helper Tool Status: %+v\n", status)
	}

	// 测试验证hosts文件
	fmt.Println("\n2. Testing ValidateHosts...")
	if err := client.ValidateHosts(ctx); err != nil {
		log.Printf("ValidateHosts failed: %v", err)
	} else {
		fmt.Println("Hosts file validation successful")
	}

	// 测试备份hosts文件
	fmt.Println("\n3. Testing BackupHosts...")
	backupPath, err := client.BackupHosts(ctx)
	if err != nil {
		log.Printf("BackupHosts failed: %v", err)
	} else {
		fmt.Printf("Hosts file backed up to: %s\n", backupPath)
	}

	// 测试写入hosts文件
	fmt.Println("\n4. Testing WriteHosts...")
	testEntries := []helper.HostEntry{
		{
			IP:       "127.0.0.1",
			Hostname: "test.local",
			Comment:  "Test entry from XPC client",
			Enabled:  true,
		},
		{
			IP:       "192.168.1.100",
			Hostname: "dev.local",
			Comment:  "Development server",
			Enabled:  false,
		},
	}

	if err := client.WriteHosts(ctx, testEntries); err != nil {
		log.Printf("WriteHosts failed: %v", err)
	} else {
		fmt.Printf("Successfully wrote %d host entries\n", len(testEntries))
	}

	// 测试恢复hosts文件
	if backupPath != "" {
		fmt.Println("\n5. Testing RestoreHosts...")
		if err := client.RestoreHosts(ctx, backupPath); err != nil {
			log.Printf("RestoreHosts failed: %v", err)
		} else {
			fmt.Printf("Hosts file restored from: %s\n", backupPath)
		}
	}

	// 测试XPC客户端池
	fmt.Println("\n6. Testing XPC Client Pool...")
	testClientPool(logger)

	// 测试Security Manager
	fmt.Println("\n=== Security Manager Test ===")
	testSecurityManager()
	fmt.Println("\n=== Security Manager Test Complete ===")

	// 测试Backup Manager
	fmt.Println("\n=== Backup Manager Test ===")
	testBackupManager()
	fmt.Println("\n=== Backup Manager Test Complete ===")

	fmt.Println("\nXPC Client Example completed successfully!")
}

// testSecurityManager 测试安全管理器
func testSecurityManager() {
	// 创建增强日志器
	logger := logger.NewEnhancedLogger(logger.LogLevelDebug, false)

	// 创建审计日志器
	auditLogger, err := helper.NewAuditLogger("/tmp/mhost-security-test.log", logger)
	if err != nil {
		fmt.Printf("Failed to create audit logger: %v\n", err)
		return
	}

	// 创建安全管理器
	securityMgr := helper.NewSecurityManager(auditLogger, logger)

	fmt.Println("\n1. Testing valid request...")
	testValidSecurityRequest(securityMgr)

	fmt.Println("\n2. Testing invalid requests...")
	testInvalidSecurityRequests(securityMgr)

	fmt.Println("\n3. Testing security stats...")
	testSecurityManagerStats(securityMgr)
}

// testValidSecurityRequest 测试有效安全请求
func testValidSecurityRequest(securityMgr helper.SecurityManager) {
	req := &helper.XPCRequest{
		ClientID:  "test-client-1",
		Operation: "get_status",
		Timestamp: time.Now(),
		Parameters: map[string]interface{}{},
	}

	err := securityMgr.ValidateRequest(req)
	if err != nil {
		fmt.Printf("Valid request failed: %v\n", err)
	} else {
		fmt.Println("✓ Valid request passed validation")
	}
}

// testInvalidSecurityRequests 测试无效安全请求
func testInvalidSecurityRequests(securityMgr helper.SecurityManager) {
	// 测试空客户端ID
	req1 := &helper.XPCRequest{
		ClientID:  "",
		Operation: "get_status",
		Timestamp: time.Now(),
		Parameters: map[string]interface{}{},
	}

	err := securityMgr.ValidateRequest(req1)
	if err != nil {
		fmt.Printf("✓ Empty client ID rejected: %v\n", err)
	} else {
		fmt.Println("✗ Empty client ID should be rejected")
	}

	// 测试无效操作
	req2 := &helper.XPCRequest{
		ClientID:  "test-client-2",
		Operation: "invalid_operation",
		Timestamp: time.Now(),
		Parameters: map[string]interface{}{},
	}

	err = securityMgr.ValidateRequest(req2)
	if err != nil {
		fmt.Printf("✓ Invalid operation rejected: %v\n", err)
	} else {
		fmt.Println("✗ Invalid operation should be rejected")
	}
}

// testSecurityManagerStats 测试安全管理器统计
func testSecurityManagerStats(securityMgr helper.SecurityManager) {
	stats := securityMgr.GetSecurityStats()
	fmt.Printf("Security stats: %+v\n", stats)

	if stats != nil {
		fmt.Println("✓ Security stats retrieved successfully")
	} else {
		fmt.Println("✗ Failed to retrieve security stats")
	}

	// 测试白名单功能
	clientID := "whitelist-test-client"
	securityMgr.AddToWhitelist(clientID)
	fmt.Printf("Added %s to whitelist\n", clientID)

	// 测试客户端哈希生成
	hash := securityMgr.GenerateClientHash("test-client-info")
	if len(hash) > 0 {
		fmt.Printf("✓ Client hash generated: %s\n", hash[:16]+"...")
	} else {
		fmt.Println("✗ Failed to generate client hash")
	}
}

// testClientPool 测试XPC客户端池
func testClientPool(logger helper.Logger) {
	// 创建客户端池
	pool := helper.NewXPCClientPool("com.mhost.helper", logger, 3)
	defer pool.Close()

	// 并发测试
	for i := 0; i < 5; i++ {
		go func(id int) {
			client, err := pool.GetClient()
			if err != nil {
				log.Printf("Pool GetClient %d failed: %v", id, err)
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			status, err := client.GetStatus(ctx)
			if err != nil {
				log.Printf("Pool client %d GetStatus failed: %v", id, err)
			} else {
				fmt.Printf("Pool client %d got status: %v\n", id, status["running"])
			}
		}(i)
	}

	// 等待并发测试完成
	time.Sleep(2 * time.Second)
	fmt.Println("Client pool test completed")
}

// testBackupManager 测试备份管理器功能
func testBackupManager() {
	// 创建增强测试logger
	logger := logger.NewEnhancedLogger(logger.LogLevelInfo, false)

	// 创建临时目录用于测试
	testDir := "/tmp/mhost-backup-test"
	os.RemoveAll(testDir) // 清理之前的测试
	os.MkdirAll(testDir, 0755)
	defer os.RemoveAll(testDir) // 测试完成后清理

	// 创建备份管理器
	backupMgr, err := helper.NewBackupManager(logger, testDir, 5)
	if err != nil {
		fmt.Printf("Failed to create backup manager: %v\n", err)
		return
	}

	// 创建测试文件
	testFile := filepath.Join(testDir, "test_hosts")
	testContent := "127.0.0.1 localhost\n192.168.1.1 router\n"
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		fmt.Printf("Failed to create test file: %v\n", err)
		return
	}

	// 测试创建备份
	fmt.Println("Testing backup creation...")
	backupInfo, err := backupMgr.CreateBackup(testFile, "test-backup", "Test backup", []string{"test"}, true)
	if err != nil {
		fmt.Printf("Failed to create backup: %v\n", err)
		return
	}
	fmt.Printf("✓ Backup created: ID=%s, Path=%s, Size=%d\n", backupInfo.ID, backupInfo.Path, backupInfo.Size)

	// 测试列出备份
	fmt.Println("\nTesting backup listing...")
	backups := backupMgr.ListBackups()
	fmt.Printf("✓ Found %d backups\n", len(backups))
	for _, backup := range backups {
		fmt.Printf("  - %s: %s (created: %s)\n", backup.ID, backup.Name, backup.CreatedAt.Format(time.RFC3339))
	}

	// 测试获取备份信息
	fmt.Println("\nTesting backup info retrieval...")
	retrievedInfo, err := backupMgr.GetBackup(backupInfo.ID)
	if err != nil {
		fmt.Printf("Failed to get backup info: %v\n", err)
		return
	}
	fmt.Printf("✓ Retrieved backup info: %s\n", retrievedInfo.Name)

	// 修改原文件
	modifiedContent := testContent + "10.0.0.1 gateway\n"
	err = os.WriteFile(testFile, []byte(modifiedContent), 0644)
	if err != nil {
		fmt.Printf("Failed to modify test file: %v\n", err)
		return
	}
	fmt.Println("✓ Original file modified")

	// 测试恢复备份
	fmt.Println("\nTesting backup restoration...")
	err = backupMgr.RestoreBackup(backupInfo.ID, testFile)
	if err != nil {
		fmt.Printf("Failed to restore backup: %v\n", err)
		return
	}
	fmt.Println("✓ Backup restored successfully")

	// 验证恢复的内容
	restoredContent, err := os.ReadFile(testFile)
	if err != nil {
		fmt.Printf("Failed to read restored file: %v\n", err)
		return
	}
	if string(restoredContent) == testContent {
		fmt.Println("✓ File content restored correctly")
	} else {
		fmt.Println("✗ File content mismatch after restoration")
	}

	// 测试备份统计
	fmt.Println("\nTesting backup statistics...")
	stats := backupMgr.GetBackupStats()
	fmt.Printf("✓ Backup stats: Total=%d, TotalSize=%d bytes\n", stats.TotalBackups, stats.TotalSize)

	// 测试验证备份
	fmt.Println("\nTesting backup validation...")
	err = backupMgr.ValidateBackup(backupInfo.ID)
	if err != nil {
		fmt.Printf("Failed to validate backup: %v\n", err)
		return
	}
	fmt.Println("✓ Backup validation passed")

	// 测试删除备份
	fmt.Println("\nTesting backup deletion...")
	err = backupMgr.DeleteBackup(backupInfo.ID)
	if err != nil {
		fmt.Printf("Failed to delete backup: %v\n", err)
		return
	}
	fmt.Println("✓ Backup deleted successfully")

	// 验证删除
	backups = backupMgr.ListBackups()
	fmt.Printf("✓ Backups after deletion: %d\n", len(backups))
}