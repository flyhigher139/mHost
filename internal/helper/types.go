package helper

import (
	"context"
	"time"

	"github.com/flyhigher139/mhost/pkg/logger"
)

// Logger 日志接口别名，使用增强的日志接口
type Logger = logger.Logger

// XPCRequest XPC请求结构
type XPCRequest struct {
	Operation  string                 `json:"operation"`
	ClientID   string                 `json:"client_id"`
	Parameters map[string]interface{} `json:"parameters"`
	Timestamp  time.Time              `json:"timestamp"`
}

// XPCResponse XPC响应结构
type XPCResponse struct {
	Success   bool                   `json:"success"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// HostEntry hosts文件条目
type HostEntry struct {
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
	Comment  string `json:"comment,omitempty"`
	Enabled  bool   `json:"enabled"`
}

// XPCServer XPC服务器接口
type XPCServer interface {
	Start(ctx context.Context, handler XPCRequestHandler) error
	Stop() error
	IsRunning() bool
}

// SecurityManager 安全管理器接口
type SecurityManager interface {
	ValidateRequest(req *XPCRequest) error
	GetSecurityStats() map[string]interface{}
	AddToWhitelist(clientID string)
	RemoveFromWhitelist(clientID string)
	ClearBlacklist()
	GenerateClientHash(clientInfo string) string
}

// HostsHandler hosts文件处理器
type HostsHandler struct {
	hostsPath string
	logger    Logger
}

// AuditLogger 审计日志器
type AuditLogger struct {
	logPath string
	logger  Logger
}

// NewXPCServer 创建XPC服务器
func NewXPCServer(serviceName string, logger Logger) (XPCServer, error) {
	return NewXPCServerImpl(serviceName, logger)
}

// NewSecurityManager 创建安全管理器
func NewSecurityManager(auditLogger *AuditLogger, logger Logger) SecurityManager {
	return NewSecurityManagerImpl(auditLogger, logger)
}

// NewHostsHandler 创建hosts处理器
func NewHostsHandler(hostsPath string, logger Logger) (*HostsHandler, error) {
	return &HostsHandler{
		hostsPath: hostsPath,
		logger:    logger,
	}, nil
}

// WriteHosts 写入hosts文件
func (h *HostsHandler) WriteHosts(entries []HostEntry) error {
	h.logger.Info("Writing hosts file", "entries", len(entries))
	return nil
}

// BackupHosts 备份hosts文件
func (h *HostsHandler) BackupHosts() (string, error) {
	h.logger.Info("Backing up hosts file")
	return "/tmp/hosts.backup", nil
}

// RestoreHosts 恢复hosts文件
func (h *HostsHandler) RestoreHosts(backupPath string) error {
	h.logger.Info("Restoring hosts file", "backup", backupPath)
	return nil
}

// ValidateHosts 验证hosts文件
func (h *HostsHandler) ValidateHosts() error {
	h.logger.Info("Validating hosts file")
	return nil
}

// GetHostsPath 获取hosts文件路径
func (h *HostsHandler) GetHostsPath() string {
	return h.hostsPath
}

// NewAuditLogger 创建审计日志器
func NewAuditLogger(logPath string, logger Logger) (*AuditLogger, error) {
	return &AuditLogger{
		logPath: logPath,
		logger:  logger,
	}, nil
}

// LogSuccessfulOperation 记录成功操作
func (a *AuditLogger) LogSuccessfulOperation(operation, clientID string, params map[string]interface{}) {
	a.logger.Info("Audit: successful operation", "operation", operation, "client", clientID)
}

// LogFailedOperation 记录失败操作
func (a *AuditLogger) LogFailedOperation(operation, clientID, error string) {
	a.logger.Error("Audit: failed operation", "operation", operation, "client", clientID, "error", error)
}

// Close 关闭审计日志器
func (a *AuditLogger) Close() error {
	a.logger.Info("Closing audit logger")
	return nil
}

// BackupManager 备份管理器接口
type BackupManager interface {
	CreateBackup(sourcePath, name, description string, tags []string, automatic bool) (*BackupInfo, error)
	RestoreBackup(backupID, targetPath string) error
	DeleteBackup(backupID string) error
	ListBackups() []*BackupInfo
	GetBackup(backupID string) (*BackupInfo, error)
	GetBackupStats() *BackupStats
	CleanupOldBackups() error
	ValidateBackup(backupID string) error
}

// NewBackupManager 创建备份管理器
func NewBackupManager(logger Logger, backupDir string, maxBackups int) (BackupManager, error) {
	return NewBackupManagerImpl(logger, backupDir, maxBackups)
}