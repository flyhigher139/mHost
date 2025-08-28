package helper

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/flyhigher139/mhost/pkg/errors"
	"github.com/flyhigher139/mhost/pkg/logger"
)

// HostsHelper Helper Tool主结构体
type HostsHelper struct {
	serviceName string
	logger      logger.Logger
	xpcServer   XPCServer
	securityMgr SecurityManager
	hostsHandler *HostsHandler
	auditLogger *AuditLogger
	backupMgr   BackupManager
	mu          sync.RWMutex
	running     bool
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewHostsHelper 创建新的HostsHelper实例
func NewHostsHelper(serviceName string, logger logger.Logger) (*HostsHelper, error) {
	if serviceName == "" {
		return nil, errors.NewValidationError(errors.ErrCodeValidationFailed, "service name cannot be empty", nil)
	}

	if logger == nil {
		return nil, errors.NewValidationError(errors.ErrCodeValidationFailed, "logger cannot be nil", nil)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// 创建审计日志器
	auditLogger, err := NewAuditLogger("/var/log/mhost-helper-audit.log", logger)
	if err != nil {
		logger.ErrorWithContext(nil, err, "Failed to create audit logger")
		return nil, errors.NewSystemError(errors.ErrCodeAuditLogFailed, "failed to create audit logger", err)
	}

	// 创建安全管理器
	securityMgr := NewSecurityManager(auditLogger, logger)

	// 创建hosts文件处理器
	hostsHandler, err := NewHostsHandler("/etc/hosts", logger)
	if err != nil {
		logger.ErrorWithContext(nil, err, "Failed to create hosts handler")
		return nil, errors.NewFileSystemError(errors.ErrCodeFileReadFailed, "failed to create hosts handler", err)
	}

	// 创建XPC服务器
	xpcServer, err := NewXPCServer(serviceName, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create XPC server: %w", err)
	}

	// 创建备份管理器
	backupMgr, err := NewBackupManager(logger, "/tmp/mhost-backups", 10)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup manager: %w", err)
	}

	return &HostsHelper{
		serviceName:  serviceName,
		logger:       logger,
		xpcServer:    xpcServer,
		securityMgr:  securityMgr,
		hostsHandler: hostsHandler,
		auditLogger:  auditLogger,
		backupMgr:    backupMgr,
		running:      false,
		ctx:          ctx,
		cancel:       cancel,
	}, nil
}

// Start 启动Helper Tool
func (h *HostsHelper) Start() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.running {
		return fmt.Errorf("HostsHelper is already running")
	}

	h.logger.Info("Starting HostsHelper", "service", h.serviceName)

	// 启动XPC服务器
	if err := h.xpcServer.Start(h.ctx, h.handleXPCRequest); err != nil {
		return fmt.Errorf("failed to start XPC server: %w", err)
	}

	h.running = true
	h.logger.Info("HostsHelper started successfully")

	return nil
}

// Stop 停止Helper Tool
func (h *HostsHelper) Stop() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.running {
		return nil
	}

	h.logger.Info("Stopping HostsHelper")

	// 取消上下文
	h.cancel()

	// 停止XPC服务器
	if err := h.xpcServer.Stop(); err != nil {
		h.logger.Error("Error stopping XPC server", "error", err)
	}

	// 关闭审计日志器
	if err := h.auditLogger.Close(); err != nil {
		h.logger.Error("Error closing audit logger", "error", err)
	}

	h.running = false
	h.logger.Info("HostsHelper stopped successfully")

	return nil
}

// IsRunning 检查Helper Tool是否正在运行
func (h *HostsHelper) IsRunning() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.running
}

// handleXPCRequest 处理XPC请求
func (h *HostsHelper) handleXPCRequest(req *XPCRequest) *XPCResponse {
	start := time.Now()

	// 记录请求开始
	h.logger.Debug("Handling XPC request", "operation", req.Operation, "client", req.ClientID)

	// 安全验证
	if err := h.securityMgr.ValidateRequest(req); err != nil {
		h.logger.Error("Security validation failed", "error", err, "client", req.ClientID)
		h.auditLogger.LogFailedOperation(req.Operation, req.ClientID, err.Error())
		return &XPCResponse{
			Success: false,
			Error:   fmt.Sprintf("Security validation failed: %v", err),
		}
	}

	// 处理具体操作
	var response *XPCResponse
	switch req.Operation {
	case "write_hosts":
		response = h.handleWriteHosts(req)
	case "backup_hosts":
		response = h.handleBackupHosts(req)
	case "restore_hosts":
		response = h.handleRestoreHosts(req)
	case "validate_hosts":
		response = h.handleValidateHosts(req)
	case "get_status":
		response = h.handleGetStatus(req)
	default:
		response = &XPCResponse{
			Success: false,
			Error:   fmt.Sprintf("Unknown operation: %s", req.Operation),
		}
	}

	// 记录操作结果
	duration := time.Since(start)
	if response.Success {
		h.logger.Info("XPC request completed", "operation", req.Operation, "duration", duration)
		h.auditLogger.LogSuccessfulOperation(req.Operation, req.ClientID, req.Parameters)
	} else {
		h.logger.Error("XPC request failed", "operation", req.Operation, "error", response.Error, "duration", duration)
		h.auditLogger.LogFailedOperation(req.Operation, req.ClientID, response.Error)
	}

	return response
}

// handleWriteHosts 处理写入hosts文件请求
func (h *HostsHelper) handleWriteHosts(req *XPCRequest) *XPCResponse {
	entries, ok := req.Parameters["entries"]
	if !ok {
		return &XPCResponse{
			Success: false,
			Error:   "missing entries parameter",
		}
	}

	// 类型断言和转换
	entriesData, ok := entries.([]interface{})
	if !ok {
		return &XPCResponse{
			Success: false,
			Error:   "invalid entries format",
		}
	}

	// 转换为HostEntry结构
	hostEntries, err := h.convertToHostEntries(entriesData)
	if err != nil {
		return &XPCResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to convert entries: %v", err),
		}
	}

	// 写入hosts文件
	if err := h.hostsHandler.WriteHosts(hostEntries); err != nil {
		return &XPCResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to write hosts file: %v", err),
		}
	}

	return &XPCResponse{
		Success: true,
		Data:    map[string]interface{}{"entries_written": len(hostEntries)},
	}
}

// handleBackupHosts 处理备份hosts文件请求
func (h *HostsHelper) handleBackupHosts(req *XPCRequest) *XPCResponse {
	h.logger.Info("Handling backup hosts request")

	// 获取备份参数
	name := "hosts-backup"
	description := "Automatic hosts file backup"
	if nameParam, ok := req.Parameters["name"].(string); ok && nameParam != "" {
		name = nameParam
	}
	if descParam, ok := req.Parameters["description"].(string); ok && descParam != "" {
		description = descParam
	}

	// 创建备份
	backupInfo, err := h.backupMgr.CreateBackup("/etc/hosts", name, description, []string{"hosts"}, true)
	if err != nil {
		h.logger.Error("Failed to create backup", "error", err)
		return &XPCResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create backup: %v", err),
		}
	}

	return &XPCResponse{
		Success: true,
		Data: map[string]interface{}{
			"backup_id":   backupInfo.ID,
			"backup_path": backupInfo.Path,
			"created_at":  backupInfo.CreatedAt,
			"size":        backupInfo.Size,
		},
	}
}

// handleRestoreHosts 处理恢复hosts文件请求
func (h *HostsHelper) handleRestoreHosts(req *XPCRequest) *XPCResponse {
	h.logger.Info("Handling restore hosts request")

	// 获取备份ID
	backupID, ok := req.Parameters["backup_id"].(string)
	if !ok {
		// 兼容旧的backup_path参数
		backupPath, pathOk := req.Parameters["backup_path"].(string)
		if !pathOk {
			return &XPCResponse{
				Success: false,
				Error:   "backup_id or backup_path parameter is required",
			}
		}
		// 如果提供的是路径，尝试从路径中提取ID
		backupID = filepath.Base(strings.TrimSuffix(backupPath, ".backup"))
	}

	// 获取目标路径（可选）
	targetPath := "/etc/hosts" // 默认恢复到原位置
	if target, ok := req.Parameters["target_path"].(string); ok && target != "" {
		targetPath = target
	}

	// 恢复备份
	err := h.backupMgr.RestoreBackup(backupID, targetPath)
	if err != nil {
		h.logger.Error("Failed to restore backup", "backup_id", backupID, "error", err)
		return &XPCResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to restore backup: %v", err),
		}
	}

	return &XPCResponse{
		Success: true,
		Data: map[string]interface{}{
			"backup_id":    backupID,
			"target_path":  targetPath,
			"restored_at":  time.Now(),
		},
	}
}

// handleValidateHosts 处理验证hosts文件请求
func (h *HostsHelper) handleValidateHosts(req *XPCRequest) *XPCResponse {
	if err := h.hostsHandler.ValidateHosts(); err != nil {
		return &XPCResponse{
			Success: false,
			Error:   fmt.Sprintf("hosts file validation failed: %v", err),
		}
	}

	return &XPCResponse{
		Success: true,
		Data:    map[string]interface{}{"status": "valid"},
	}
}

// handleGetStatus 处理获取状态请求
func (h *HostsHelper) handleGetStatus(req *XPCRequest) *XPCResponse {
	status := map[string]interface{}{
		"running":     h.IsRunning(),
		"service":     h.serviceName,
		"uptime":      time.Since(time.Now()).String(), // 这里应该记录启动时间
		"hosts_path":  h.hostsHandler.GetHostsPath(),
	}

	return &XPCResponse{
		Success: true,
		Data:    status,
	}
}

// convertToHostEntries 转换接口数据为HostEntry结构
func (h *HostsHelper) convertToHostEntries(data []interface{}) ([]HostEntry, error) {
	var entries []HostEntry

	for i, item := range data {
		entryMap, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid entry format at index %d", i)
		}

		ip, ok := entryMap["ip"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid ip at index %d", i)
		}

		hostname, ok := entryMap["hostname"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid hostname at index %d", i)
		}

		comment, _ := entryMap["comment"].(string)
		enabled, _ := entryMap["enabled"].(bool)

		entries = append(entries, HostEntry{
			IP:       ip,
			Hostname: hostname,
			Comment:  comment,
			Enabled:  enabled,
		})
	}

	return entries, nil
}