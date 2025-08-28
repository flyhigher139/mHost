package helper

import (
	"crypto/sha256"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/flyhigher139/mhost/pkg/errors"
	"github.com/flyhigher139/mhost/pkg/logger"
)

// SecurityManagerImpl 安全管理器实现
type SecurityManagerImpl struct {
	auditLogger *AuditLogger
	logger      logger.Logger
	config      *SecurityConfig
	mu          sync.RWMutex
	blacklist   map[string]time.Time // IP黑名单
	whitelist   map[string]bool      // IP白名单
	rateLimit   map[string]*RateLimiter
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	MaxRequestsPerMinute int           `json:"max_requests_per_minute"`
	BlacklistDuration    time.Duration `json:"blacklist_duration"`
	RequireAuth          bool          `json:"require_auth"`
	AllowedOperations    []string      `json:"allowed_operations"`
	TrustedClients       []string      `json:"trusted_clients"`
	MaxHostEntries       int           `json:"max_host_entries"`
	ValidateHostnames    bool          `json:"validate_hostnames"`
	ValidateIPs          bool          `json:"validate_ips"`
}

// RateLimiter 速率限制器
type RateLimiter struct {
	requests  []time.Time
	maxReqs   int
	window    time.Duration
	mu        sync.Mutex
}

// SecurityViolation 安全违规记录
type SecurityViolation struct {
	ClientID    string    `json:"client_id"`
	Violation   string    `json:"violation"`
	Operation   string    `json:"operation"`
	Timestamp   time.Time `json:"timestamp"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
}

// NewSecurityManagerImpl 创建安全管理器实现
func NewSecurityManagerImpl(auditLogger *AuditLogger, logger logger.Logger) *SecurityManagerImpl {
	config := &SecurityConfig{
		MaxRequestsPerMinute: 60,
		BlacklistDuration:    15 * time.Minute,
		RequireAuth:          true,
		AllowedOperations: []string{
			"write_hosts",
			"backup_hosts",
			"restore_hosts",
			"validate_hosts",
			"get_status",
		},
		TrustedClients:    []string{},
		MaxHostEntries:    1000,
		ValidateHostnames: true,
		ValidateIPs:       true,
	}

	return &SecurityManagerImpl{
		auditLogger: auditLogger,
		logger:      logger,
		config:      config,
		blacklist:   make(map[string]time.Time),
		whitelist:   make(map[string]bool),
		rateLimit:   make(map[string]*RateLimiter),
	}
}

// ValidateRequest 验证XPC请求
func (s *SecurityManagerImpl) ValidateRequest(req *XPCRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 基本验证
	if err := s.validateBasicRequest(req); err != nil {
		s.logSecurityViolation(req.ClientID, "basic_validation", req.Operation, "high", err.Error())
		return fmt.Errorf("basic validation failed: %w", err)
	}

	// 检查黑名单
	if s.isBlacklisted(req.ClientID) {
		s.logSecurityViolation(req.ClientID, "blacklisted", req.Operation, "high", "Client is blacklisted")
		s.logger.Warn("Client is blacklisted", "client_id", req.ClientID)
		return errors.NewPermissionError(errors.ErrCodeClientBlacklisted, "client is blacklisted")
	}

	// 速率限制检查
	if !s.checkRateLimit(req.ClientID) {
		s.logSecurityViolation(req.ClientID, "rate_limit", req.Operation, "medium", "Rate limit exceeded")
		s.addToBlacklist(req.ClientID)
		s.logger.Warn("Rate limit exceeded", "client_id", req.ClientID)
		return errors.NewPermissionError(errors.ErrCodeRateLimitExceeded, "rate limit exceeded")
	}

	// 操作权限检查
	if !s.isOperationAllowed(req.Operation) {
		s.logSecurityViolation(req.ClientID, "unauthorized_operation", req.Operation, "high", "Operation not allowed")
		s.logger.Warn("Operation not allowed", "operation", req.Operation, "client_id", req.ClientID)
		return errors.NewPermissionError(errors.ErrCodeOperationNotAllowed, fmt.Sprintf("operation not allowed: %s", req.Operation))
	}

	// 参数验证
	if err := s.validateParameters(req); err != nil {
		s.logSecurityViolation(req.ClientID, "parameter_validation", req.Operation, "medium", err.Error())
		return fmt.Errorf("parameter validation failed: %w", err)
	}

	s.logger.Debug("Request validation passed", "client", req.ClientID, "operation", req.Operation)
	return nil
}

// validateBasicRequest 基本请求验证
func (s *SecurityManagerImpl) validateBasicRequest(req *XPCRequest) error {
	if req == nil {
		return fmt.Errorf("request is nil")
	}

	if req.ClientID == "" {
		return fmt.Errorf("client ID is empty")
	}

	if req.Operation == "" {
		return fmt.Errorf("operation is empty")
	}

	if req.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is zero")
	}

	// 检查请求时间戳是否合理（不能太旧或太新）
	now := time.Now()
	if req.Timestamp.Before(now.Add(-5*time.Minute)) {
		return fmt.Errorf("request timestamp too old")
	}
	if req.Timestamp.After(now.Add(1*time.Minute)) {
		return fmt.Errorf("request timestamp too new")
	}

	return nil
}

// isBlacklisted 检查客户端是否在黑名单中
func (s *SecurityManagerImpl) isBlacklisted(clientID string) bool {
	if expiry, exists := s.blacklist[clientID]; exists {
		if time.Now().Before(expiry) {
			return true
		}
		// 过期的黑名单条目，删除
		delete(s.blacklist, clientID)
	}
	return false
}

// addToBlacklist 添加客户端到黑名单
func (s *SecurityManagerImpl) addToBlacklist(clientID string) {
	expiry := time.Now().Add(s.config.BlacklistDuration)
	s.blacklist[clientID] = expiry
	s.logger.Warn("Client added to blacklist", "client", clientID, "expiry", expiry)
}

// checkRateLimit 检查速率限制
func (s *SecurityManagerImpl) checkRateLimit(clientID string) bool {
	// 如果客户端在白名单中，跳过速率限制
	if s.whitelist[clientID] {
		return true
	}

	limiter, exists := s.rateLimit[clientID]
	if !exists {
		limiter = &RateLimiter{
			requests: make([]time.Time, 0),
			maxReqs:  s.config.MaxRequestsPerMinute,
			window:   time.Minute,
		}
		s.rateLimit[clientID] = limiter
	}

	return limiter.Allow()
}

// isOperationAllowed 检查操作是否被允许
func (s *SecurityManagerImpl) isOperationAllowed(operation string) bool {
	for _, allowed := range s.config.AllowedOperations {
		if operation == allowed {
			return true
		}
	}
	return false
}

// validateParameters 验证请求参数
func (s *SecurityManagerImpl) validateParameters(req *XPCRequest) error {
	switch req.Operation {
	case "write_hosts":
		return s.validateWriteHostsParams(req.Parameters)
	case "restore_hosts":
		return s.validateRestoreHostsParams(req.Parameters)
	case "backup_hosts", "validate_hosts", "get_status":
		// 这些操作不需要特殊参数验证
		return nil
	default:
		return fmt.Errorf("unknown operation: %s", req.Operation)
	}
}

// validateWriteHostsParams 验证写入hosts参数
func (s *SecurityManagerImpl) validateWriteHostsParams(params map[string]interface{}) error {
	entries, ok := params["entries"]
	if !ok {
		return fmt.Errorf("missing entries parameter")
	}

	entriesSlice, ok := entries.([]interface{})
	if !ok {
		return fmt.Errorf("entries must be an array")
	}

	if len(entriesSlice) > s.config.MaxHostEntries {
		return fmt.Errorf("too many host entries: %d (max: %d)", len(entriesSlice), s.config.MaxHostEntries)
	}

	// 验证每个条目
	for i, entry := range entriesSlice {
		entryMap, ok := entry.(map[string]interface{})
		if !ok {
			return fmt.Errorf("entry %d is not a valid object", i)
		}

		if err := s.validateHostEntry(entryMap); err != nil {
			return fmt.Errorf("entry %d validation failed: %w", i, err)
		}
	}

	return nil
}

// validateRestoreHostsParams 验证恢复hosts参数
func (s *SecurityManagerImpl) validateRestoreHostsParams(params map[string]interface{}) error {
	backupPath, ok := params["backup_path"]
	if !ok {
		return fmt.Errorf("missing backup_path parameter")
	}

	backupPathStr, ok := backupPath.(string)
	if !ok {
		return fmt.Errorf("backup_path must be a string")
	}

	// 验证路径安全性
	if err := s.validateFilePath(backupPathStr); err != nil {
		return fmt.Errorf("invalid backup path: %w", err)
	}

	return nil
}

// validateHostEntry 验证单个host条目
func (s *SecurityManagerImpl) validateHostEntry(entry map[string]interface{}) error {
	ip, ok := entry["ip"].(string)
	if !ok || ip == "" {
		return fmt.Errorf("missing or invalid ip")
	}

	hostname, ok := entry["hostname"].(string)
	if !ok || hostname == "" {
		return fmt.Errorf("missing or invalid hostname")
	}

	// 验证IP地址
	if s.config.ValidateIPs {
		if err := s.validateIPAddress(ip); err != nil {
			return fmt.Errorf("invalid IP address: %w", err)
		}
	}

	// 验证主机名
	if s.config.ValidateHostnames {
		if err := s.validateHostname(hostname); err != nil {
			return fmt.Errorf("invalid hostname: %w", err)
		}
	}

	// 验证注释（如果存在）
	if comment, ok := entry["comment"].(string); ok {
		if len(comment) > 200 {
			return fmt.Errorf("comment too long (max 200 characters)")
		}
	}

	return nil
}

// validateIPAddress 验证IP地址
func (s *SecurityManagerImpl) validateIPAddress(ip string) error {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return fmt.Errorf("invalid IP format")
	}

	// 检查是否为危险的IP地址
	if s.isDangerousIP(parsedIP) {
		return fmt.Errorf("dangerous IP address not allowed")
	}

	return nil
}

// validateHostname 验证主机名
func (s *SecurityManagerImpl) validateHostname(hostname string) error {
	// 基本长度检查
	if len(hostname) > 253 {
		return fmt.Errorf("hostname too long (max 253 characters)")
	}

	// 正则表达式验证
	hostnameRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	if !hostnameRegex.MatchString(hostname) {
		return fmt.Errorf("invalid hostname format")
	}

	// 检查是否为危险的主机名
	if s.isDangerousHostname(hostname) {
		return fmt.Errorf("dangerous hostname not allowed")
	}

	return nil
}

// validateFilePath 验证文件路径
func (s *SecurityManagerImpl) validateFilePath(path string) error {
	// 检查路径遍历攻击
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal not allowed")
	}

	// 检查绝对路径
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("only absolute paths allowed")
	}

	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist")
	}

	return nil
}

// isDangerousIP 检查是否为危险IP
func (s *SecurityManagerImpl) isDangerousIP(ip net.IP) bool {
	// 检查是否为广播地址或多播地址
	if ip.IsMulticast() || ip.IsUnspecified() {
		return true
	}

	// 可以添加更多危险IP检查逻辑
	return false
}

// isDangerousHostname 检查是否为危险主机名
func (s *SecurityManagerImpl) isDangerousHostname(hostname string) bool {
	dangerousPatterns := []string{
		"localhost",
		"127.0.0.1",
		"0.0.0.0",
		"*.local",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(strings.ToLower(hostname), strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}

// logSecurityViolation 记录安全违规
func (s *SecurityManagerImpl) logSecurityViolation(clientID, violation, operation, severity, description string) {
	s.logger.Error("Security violation detected",
		"client", clientID,
		"violation", violation,
		"operation", operation,
		"severity", severity,
		"description", description)

	// 记录到审计日志
	s.auditLogger.LogFailedOperation(operation, clientID, fmt.Sprintf("%s: %s", violation, description))
}

// Allow 速率限制器允许请求
func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-r.window)

	// 清理过期的请求记录
	var validRequests []time.Time
	for _, reqTime := range r.requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}
	r.requests = validRequests

	// 检查是否超过限制
	if len(r.requests) >= r.maxReqs {
		return false
	}

	// 添加当前请求
	r.requests = append(r.requests, now)
	return true
}

// GetSecurityStats 获取安全统计信息
func (s *SecurityManagerImpl) GetSecurityStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"blacklisted_clients": len(s.blacklist),
		"whitelisted_clients": len(s.whitelist),
		"rate_limited_clients": len(s.rateLimit),
		"config": s.config,
	}
}

// AddToWhitelist 添加客户端到白名单
func (s *SecurityManagerImpl) AddToWhitelist(clientID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.whitelist[clientID] = true
	s.logger.Info("Client added to whitelist", "client", clientID)
}

// RemoveFromWhitelist 从白名单移除客户端
func (s *SecurityManagerImpl) RemoveFromWhitelist(clientID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.whitelist, clientID)
	s.logger.Info("Client removed from whitelist", "client", clientID)
}

// ClearBlacklist 清空黑名单
func (s *SecurityManagerImpl) ClearBlacklist() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.blacklist = make(map[string]time.Time)
	s.logger.Info("Blacklist cleared")
}

// GenerateClientHash 生成客户端哈希
func (s *SecurityManagerImpl) GenerateClientHash(clientInfo string) string {
	hash := sha256.Sum256([]byte(clientInfo + time.Now().String()))
	return fmt.Sprintf("%x", hash)
}