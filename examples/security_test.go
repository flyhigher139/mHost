package main

import (
	"fmt"
	"log"
	"time"

	"github.com/flyhigher139/mhost/internal/helper"
	"github.com/flyhigher139/mhost/pkg/logger"
)

// SecurityTestMain 安全管理器测试示例
func SecurityTestMain() {
	fmt.Println("=== Security Manager Test ===")

	// 创建增强日志器
	logger := logger.NewEnhancedLogger(logger.LogLevelDebug, false)

	// 创建审计日志器
	auditLogger, err := helper.NewAuditLogger("/tmp/mhost-security-test.log", logger)
	if err != nil {
		log.Fatalf("Failed to create audit logger: %v", err)
	}

	// 创建安全管理器
	securityMgr := helper.NewSecurityManager(auditLogger, logger)

	fmt.Println("\n1. Testing valid request...")
	testValidRequest(securityMgr)

	fmt.Println("\n2. Testing invalid requests...")
	testInvalidRequests(securityMgr)

	fmt.Println("\n3. Testing rate limiting...")
	testRateLimiting(securityMgr)

	fmt.Println("\n4. Testing security stats...")
	testSecurityStats(securityMgr)

	fmt.Println("\n5. Testing whitelist functionality...")
	testWhitelist(securityMgr)

	fmt.Println("\n=== Security Manager Test Complete ===")
}

// testValidRequest 测试有效请求
func testValidRequest(securityMgr helper.SecurityManager) {
	req := &helper.XPCRequest{
		ClientID:  "test-client-1",
		Operation: "get_status",
		Timestamp: time.Now(),
		Parameters: map[string]interface{}{},
	}

	err := securityMgr.ValidateRequest(req)
	if err != nil {
		log.Printf("Valid request failed: %v", err)
	} else {
		fmt.Println("✓ Valid request passed validation")
	}
}

// testInvalidRequests 测试无效请求
func testInvalidRequests(securityMgr helper.SecurityManager) {
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

	// 测试过期时间戳
	req3 := &helper.XPCRequest{
		ClientID:  "test-client-3",
		Operation: "get_status",
		Timestamp: time.Now().Add(-10 * time.Minute), // 10分钟前
		Parameters: map[string]interface{}{},
	}

	err = securityMgr.ValidateRequest(req3)
	if err != nil {
		fmt.Printf("✓ Old timestamp rejected: %v\n", err)
	} else {
		fmt.Println("✗ Old timestamp should be rejected")
	}

	// 测试无效的hosts条目
	req4 := &helper.XPCRequest{
		ClientID:  "test-client-4",
		Operation: "write_hosts",
		Timestamp: time.Now(),
		Parameters: map[string]interface{}{
			"entries": []interface{}{
				map[string]interface{}{
					"ip":       "invalid-ip",
					"hostname": "example.com",
				},
			},
		},
	}

	err = securityMgr.ValidateRequest(req4)
	if err != nil {
		fmt.Printf("✓ Invalid IP rejected: %v\n", err)
	} else {
		fmt.Println("✗ Invalid IP should be rejected")
	}
}

// testRateLimiting 测试速率限制
func testRateLimiting(securityMgr helper.SecurityManager) {
	clientID := "rate-limit-test-client"
	successCount := 0
	failCount := 0

	// 发送大量请求测试速率限制
	for i := 0; i < 70; i++ {
		req := &helper.XPCRequest{
			ClientID:  clientID,
			Operation: "get_status",
			Timestamp: time.Now(),
			Parameters: map[string]interface{}{},
		}

		err := securityMgr.ValidateRequest(req)
		if err != nil {
			failCount++
		} else {
			successCount++
		}
	}

	fmt.Printf("Rate limiting test: %d successful, %d failed\n", successCount, failCount)
	if failCount > 0 {
		fmt.Println("✓ Rate limiting is working")
	} else {
		fmt.Println("✗ Rate limiting may not be working properly")
	}
}

// testSecurityStats 测试安全统计
func testSecurityStats(securityMgr helper.SecurityManager) {
	stats := securityMgr.GetSecurityStats()
	fmt.Printf("Security stats: %+v\n", stats)

	if stats != nil {
		fmt.Println("✓ Security stats retrieved successfully")
	} else {
		fmt.Println("✗ Failed to retrieve security stats")
	}
}

// testWhitelist 测试白名单功能
func testWhitelist(securityMgr helper.SecurityManager) {
	clientID := "whitelist-test-client"

	// 添加到白名单
	securityMgr.AddToWhitelist(clientID)
	fmt.Printf("Added %s to whitelist\n", clientID)

	// 测试白名单客户端是否可以发送大量请求
	successCount := 0
	for i := 0; i < 70; i++ {
		req := &helper.XPCRequest{
			ClientID:  clientID,
			Operation: "get_status",
			Timestamp: time.Now(),
			Parameters: map[string]interface{}{},
		}

		err := securityMgr.ValidateRequest(req)
		if err == nil {
			successCount++
		}
	}

	if successCount == 70 {
		fmt.Println("✓ Whitelist functionality working - all requests passed")
	} else {
		fmt.Printf("✗ Whitelist may not be working - only %d/70 requests passed\n", successCount)
	}

	// 从白名单移除
	securityMgr.RemoveFromWhitelist(clientID)
	fmt.Printf("Removed %s from whitelist\n", clientID)

	// 清空黑名单
	securityMgr.ClearBlacklist()
	fmt.Println("Cleared blacklist")

	// 测试客户端哈希生成
	hash := securityMgr.GenerateClientHash("test-client-info")
	if len(hash) > 0 {
		fmt.Printf("✓ Client hash generated: %s\n", hash[:16]+"...")
	} else {
		fmt.Println("✗ Failed to generate client hash")
	}
}