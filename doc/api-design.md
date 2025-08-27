# mHost API 设计文档

## 1. 概述

本文档定义了 mHost 应用程序内部各模块之间的 API 接口规范。这些接口确保了模块间的松耦合和高内聚，便于测试、维护和扩展。

## 2. 核心接口定义

### 2.1 Profile Manager API

#### 接口定义

```go
package core

import (
    "context"
    "time"
)

// ProfileManager 定义 Profile 管理的核心接口
type ProfileManager interface {
    // CreateProfile 创建新的 Profile
    // 参数:
    //   ctx: 上下文
    //   req: 创建请求
    // 返回:
    //   创建的 Profile 和可能的错误
    CreateProfile(ctx context.Context, req *CreateProfileRequest) (*Profile, error)
    
    // GetProfile 根据 ID 获取 Profile
    GetProfile(ctx context.Context, id string) (*Profile, error)
    
    // ListProfiles 获取所有 Profile 列表
    ListProfiles(ctx context.Context, req *ListProfilesRequest) (*ListProfilesResponse, error)
    
    // UpdateProfile 更新 Profile
    UpdateProfile(ctx context.Context, req *UpdateProfileRequest) (*Profile, error)
    
    // DeleteProfile 删除 Profile
    DeleteProfile(ctx context.Context, id string) error
    
    // ApplyProfile 应用 Profile 到系统
    ApplyProfile(ctx context.Context, id string) error
    
    // GetActiveProfile 获取当前激活的 Profile
    GetActiveProfile(ctx context.Context) (*Profile, error)
    
    // ImportProfile 从文件导入 Profile
    ImportProfile(ctx context.Context, req *ImportProfileRequest) (*Profile, error)
    
    // ExportProfile 导出 Profile 到文件
    ExportProfile(ctx context.Context, req *ExportProfileRequest) error
}
```

#### 请求/响应结构

```go
// CreateProfileRequest 创建 Profile 请求
type CreateProfileRequest struct {
    Name        string `json:"name" validate:"required,min=1,max=50"`
    Description string `json:"description" validate:"max=200"`
    BaseProfile string `json:"base_profile,omitempty"` // 基于现有 Profile 创建
}

// UpdateProfileRequest 更新 Profile 请求
type UpdateProfileRequest struct {
    ID          string `json:"id" validate:"required"`
    Name        string `json:"name" validate:"required,min=1,max=50"`
    Description string `json:"description" validate:"max=200"`
    Entries     []HostEntry `json:"entries"`
}

// ListProfilesRequest 列表查询请求
type ListProfilesRequest struct {
    Search   string `json:"search,omitempty"`   // 搜索关键词
    SortBy   string `json:"sort_by,omitempty"`  // 排序字段
    SortDesc bool   `json:"sort_desc,omitempty"` // 是否降序
}

// ListProfilesResponse 列表查询响应
type ListProfilesResponse struct {
    Profiles []*Profile `json:"profiles"`
    Total    int        `json:"total"`
}

// ImportProfileRequest 导入 Profile 请求
type ImportProfileRequest struct {
    FilePath string `json:"file_path" validate:"required"`
    Name     string `json:"name,omitempty"` // 可选，覆盖文件中的名称
}

// ExportProfileRequest 导出 Profile 请求
type ExportProfileRequest struct {
    ProfileID string `json:"profile_id" validate:"required"`
    FilePath  string `json:"file_path" validate:"required"`
    Format    string `json:"format" validate:"oneof=json yaml hosts"` // 导出格式
}
```

### 2.2 Host Manager API

#### 接口定义

```go
// HostManager 定义 Host 管理的核心接口
type HostManager interface {
    // ReadSystemHosts 读取系统 hosts 文件
    ReadSystemHosts(ctx context.Context) (*HostsFile, error)
    
    // WriteSystemHosts 写入系统 hosts 文件
    WriteSystemHosts(ctx context.Context, req *WriteHostsRequest) error
    
    // ValidateEntry 验证单个 Host 条目
    ValidateEntry(ctx context.Context, entry *HostEntry) error
    
    // ValidateEntries 批量验证 Host 条目
    ValidateEntries(ctx context.Context, entries []HostEntry) (*ValidationResult, error)
    
    // ParseHostsContent 解析 hosts 文件内容
    ParseHostsContent(ctx context.Context, content string) ([]HostEntry, error)
    
    // GenerateHostsContent 生成 hosts 文件内容
    GenerateHostsContent(ctx context.Context, entries []HostEntry) (string, error)
    
    // MergeEntries 合并多个 Host 条目列表
    MergeEntries(ctx context.Context, req *MergeEntriesRequest) ([]HostEntry, error)
}
```

#### 请求/响应结构

```go
// HostsFile 表示 hosts 文件
type HostsFile struct {
    Path         string      `json:"path"`
    Content      string      `json:"content"`
    Entries      []HostEntry `json:"entries"`
    LastModified time.Time   `json:"last_modified"`
    Size         int64       `json:"size"`
}

// WriteHostsRequest 写入 hosts 文件请求
type WriteHostsRequest struct {
    Entries    []HostEntry `json:"entries" validate:"required"`
    BackupPath string      `json:"backup_path,omitempty"` // 备份文件路径
    DryRun     bool        `json:"dry_run,omitempty"`     // 是否为试运行
}

// ValidationResult 验证结果
type ValidationResult struct {
    Valid   bool                    `json:"valid"`
    Errors  []ValidationError       `json:"errors,omitempty"`
    Warnings []ValidationWarning    `json:"warnings,omitempty"`
    Summary *ValidationSummary      `json:"summary"`
}

// ValidationError 验证错误
type ValidationError struct {
    EntryIndex int    `json:"entry_index"`
    Field      string `json:"field"`
    Message    string `json:"message"`
    Code       string `json:"code"`
}

// ValidationWarning 验证警告
type ValidationWarning struct {
    EntryIndex int    `json:"entry_index"`
    Field      string `json:"field"`
    Message    string `json:"message"`
    Code       string `json:"code"`
}

// ValidationSummary 验证摘要
type ValidationSummary struct {
    TotalEntries    int `json:"total_entries"`
    ValidEntries    int `json:"valid_entries"`
    InvalidEntries  int `json:"invalid_entries"`
    DuplicateEntries int `json:"duplicate_entries"`
}

// MergeEntriesRequest 合并条目请求
type MergeEntriesRequest struct {
    BasEntries    []HostEntry `json:"base_entries"`
    OverrideEntries []HostEntry `json:"override_entries"`
    Strategy      string      `json:"strategy" validate:"oneof=replace merge skip"`
}
```

### 2.3 Config Manager API

#### 接口定义

```go
// ConfigManager 定义配置管理的核心接口
type ConfigManager interface {
    // LoadConfig 加载应用配置
    LoadConfig(ctx context.Context) (*AppConfig, error)
    
    // SaveConfig 保存应用配置
    SaveConfig(ctx context.Context, config *AppConfig) error
    
    // GetConfigValue 获取配置值
    GetConfigValue(ctx context.Context, key string) (interface{}, error)
    
    // SetConfigValue 设置配置值
    SetConfigValue(ctx context.Context, key string, value interface{}) error
    
    // ResetConfig 重置配置到默认值
    ResetConfig(ctx context.Context) error
    
    // ValidateConfig 验证配置
    ValidateConfig(ctx context.Context, config *AppConfig) error
    
    // GetConfigPath 获取配置文件路径
    GetConfigPath() string
    
    // WatchConfig 监听配置变化
    WatchConfig(ctx context.Context) (<-chan *ConfigChangeEvent, error)
}
```

#### 配置结构

```go
// AppConfig 应用程序配置
type AppConfig struct {
    Version       string        `json:"version"`
    LastProfile   string        `json:"last_profile"`
    Window        WindowConfig  `json:"window"`
    Backup        BackupConfig  `json:"backup"`
    Logging       LogConfig     `json:"logging"`
    Security      SecurityConfig `json:"security"`
    UI            UIConfig      `json:"ui"`
}

// WindowConfig 窗口配置
type WindowConfig struct {
    Width     int  `json:"width" validate:"min=800"`
    Height    int  `json:"height" validate:"min=600"`
    X         int  `json:"x"`
    Y         int  `json:"y"`
    Maximized bool `json:"maximized"`
    Resizable bool `json:"resizable"`
}

// BackupConfig 备份配置
type BackupConfig struct {
    Enabled       bool   `json:"enabled"`
    MaxBackups    int    `json:"max_backups" validate:"min=1,max=50"`
    AutoBackup    bool   `json:"auto_backup"`
    BackupOnApply bool   `json:"backup_on_apply"`
    BackupPath    string `json:"backup_path"`
}

// LogConfig 日志配置
type LogConfig struct {
    Level      string `json:"level" validate:"oneof=debug info warn error"`
    MaxSize    int    `json:"max_size" validate:"min=1"`    // MB
    MaxBackups int    `json:"max_backups" validate:"min=1"`
    MaxAge     int    `json:"max_age" validate:"min=1"`     // days
    Compress   bool   `json:"compress"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
    RequireConfirmation bool `json:"require_confirmation"`
    AutoElevate        bool `json:"auto_elevate"`
    ValidateOnApply    bool `json:"validate_on_apply"`
}

// UIConfig UI 配置
type UIConfig struct {
    Theme         string `json:"theme" validate:"oneof=light dark auto"`
    Language      string `json:"language" validate:"oneof=en zh-CN"`
    ShowLineNumbers bool `json:"show_line_numbers"`
    AutoSave      bool `json:"auto_save"`
    ConfirmDelete bool `json:"confirm_delete"`
}

// ConfigChangeEvent 配置变化事件
type ConfigChangeEvent struct {
    Key      string      `json:"key"`
    OldValue interface{} `json:"old_value"`
    NewValue interface{} `json:"new_value"`
    Timestamp time.Time  `json:"timestamp"`
}
```

### 2.4 Security Manager API

#### 接口定义

```go
// SecurityManager 定义安全管理的核心接口
type SecurityManager interface {
    // CheckPermissions 检查当前权限
    CheckPermissions(ctx context.Context) (*PermissionStatus, error)
    
    // RequestPermission 请求权限
    RequestPermission(ctx context.Context, req *PermissionRequest) error
    
    // ValidateFileAccess 验证文件访问权限
    ValidateFileAccess(ctx context.Context, path string, mode AccessMode) error
    
    // ExecuteWithPermission 以特定权限执行操作
    ExecuteWithPermission(ctx context.Context, req *ExecuteRequest) error
    
    // AuditOperation 审计操作
    AuditOperation(ctx context.Context, operation *AuditOperation) error
    
    // GetSecurityStatus 获取安全状态
    GetSecurityStatus(ctx context.Context) (*SecurityStatus, error)
}
```

#### 安全相关结构

```go
// PermissionStatus 权限状态
type PermissionStatus struct {
    HasAdminRights   bool      `json:"has_admin_rights"`
    CanWriteHosts    bool      `json:"can_write_hosts"`
    LastChecked      time.Time `json:"last_checked"`
    PermissionSource string    `json:"permission_source"` // sudo, admin, etc.
}

// PermissionRequest 权限请求
type PermissionRequest struct {
    Type        PermissionType `json:"type"`
    Reason      string         `json:"reason"`
    Duration    time.Duration  `json:"duration,omitempty"`
    Interactive bool           `json:"interactive"`
}

// PermissionType 权限类型
type PermissionType string

const (
    PermissionTypeAdmin     PermissionType = "admin"
    PermissionTypeFileWrite PermissionType = "file_write"
    PermissionTypeSystemAccess PermissionType = "system_access"
)

// AccessMode 访问模式
type AccessMode string

const (
    AccessModeRead  AccessMode = "read"
    AccessModeWrite AccessMode = "write"
    AccessModeExecute AccessMode = "execute"
)

// ExecuteRequest 执行请求
type ExecuteRequest struct {
    Operation   func() error   `json:"-"`
    Description string         `json:"description"`
    RequiredPermissions []PermissionType `json:"required_permissions"`
    Timeout     time.Duration  `json:"timeout"`
}

// AuditOperation 审计操作
type AuditOperation struct {
    Type        string                 `json:"type"`
    Description string                 `json:"description"`
    User        string                 `json:"user"`
    Timestamp   time.Time              `json:"timestamp"`
    Success     bool                   `json:"success"`
    Error       string                 `json:"error,omitempty"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SecurityStatus 安全状态
type SecurityStatus struct {
    OverallStatus   string             `json:"overall_status"` // secure, warning, error
    Permissions     *PermissionStatus  `json:"permissions"`
    LastAudit       time.Time          `json:"last_audit"`
    SecurityIssues  []SecurityIssue    `json:"security_issues,omitempty"`
}

// SecurityIssue 安全问题
type SecurityIssue struct {
    Type        string `json:"type"`
    Severity    string `json:"severity"` // low, medium, high, critical
    Description string `json:"description"`
    Recommendation string `json:"recommendation"`
}
```

### 2.5 Backup Manager API

#### 接口定义

```go
// BackupManager 定义备份管理的核心接口
type BackupManager interface {
    // CreateBackup 创建备份
    CreateBackup(ctx context.Context, req *CreateBackupRequest) (*Backup, error)
    
    // ListBackups 列出所有备份
    ListBackups(ctx context.Context, req *ListBackupsRequest) (*ListBackupsResponse, error)
    
    // RestoreBackup 恢复备份
    RestoreBackup(ctx context.Context, req *RestoreBackupRequest) error
    
    // DeleteBackup 删除备份
    DeleteBackup(ctx context.Context, backupID string) error
    
    // ValidateBackup 验证备份完整性
    ValidateBackup(ctx context.Context, backupID string) (*BackupValidation, error)
    
    // CleanupBackups 清理旧备份
    CleanupBackups(ctx context.Context) (*CleanupResult, error)
    
    // GetBackupInfo 获取备份信息
    GetBackupInfo(ctx context.Context, backupID string) (*Backup, error)
}
```

#### 备份相关结构

```go
// Backup 备份信息
type Backup struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    FilePath    string    `json:"file_path"`
    SourcePath  string    `json:"source_path"`
    Size        int64     `json:"size"`
    Checksum    string    `json:"checksum"`
    CreatedAt   time.Time `json:"created_at"`
    Type        BackupType `json:"type"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// BackupType 备份类型
type BackupType string

const (
    BackupTypeManual    BackupType = "manual"
    BackupTypeAutomatic BackupType = "automatic"
    BackupTypeScheduled BackupType = "scheduled"
)

// CreateBackupRequest 创建备份请求
type CreateBackupRequest struct {
    Name        string    `json:"name,omitempty"`
    Description string    `json:"description,omitempty"`
    SourcePath  string    `json:"source_path" validate:"required"`
    Type        BackupType `json:"type"`
}

// ListBackupsRequest 列出备份请求
type ListBackupsRequest struct {
    Type     BackupType `json:"type,omitempty"`
    SortBy   string     `json:"sort_by,omitempty"`
    SortDesc bool       `json:"sort_desc,omitempty"`
    Limit    int        `json:"limit,omitempty"`
    Offset   int        `json:"offset,omitempty"`
}

// ListBackupsResponse 列出备份响应
type ListBackupsResponse struct {
    Backups []*Backup `json:"backups"`
    Total   int       `json:"total"`
}

// RestoreBackupRequest 恢复备份请求
type RestoreBackupRequest struct {
    BackupID   string `json:"backup_id" validate:"required"`
    TargetPath string `json:"target_path,omitempty"` // 可选，默认为原路径
    Force      bool   `json:"force,omitempty"`       // 是否强制覆盖
}

// BackupValidation 备份验证结果
type BackupValidation struct {
    Valid       bool     `json:"valid"`
    Checksum    string   `json:"checksum"`
    ExpectedChecksum string `json:"expected_checksum"`
    Errors      []string `json:"errors,omitempty"`
}

// CleanupResult 清理结果
type CleanupResult struct {
    DeletedCount int      `json:"deleted_count"`
    DeletedBackups []string `json:"deleted_backups"`
    FreedSpace   int64    `json:"freed_space"` // bytes
}
```

## 3. 事件系统 API

### 3.1 事件管理器

```go
// EventManager 事件管理器接口
type EventManager interface {
    // Subscribe 订阅事件
    Subscribe(ctx context.Context, eventType EventType, handler EventHandler) error
    
    // Unsubscribe 取消订阅
    Unsubscribe(ctx context.Context, eventType EventType, handler EventHandler) error
    
    // Publish 发布事件
    Publish(ctx context.Context, event *Event) error
    
    // PublishAsync 异步发布事件
    PublishAsync(ctx context.Context, event *Event) error
}

// EventType 事件类型
type EventType string

const (
    EventTypeProfileCreated  EventType = "profile.created"
    EventTypeProfileUpdated  EventType = "profile.updated"
    EventTypeProfileDeleted  EventType = "profile.deleted"
    EventTypeProfileApplied  EventType = "profile.applied"
    EventTypeHostsChanged    EventType = "hosts.changed"
    EventTypeBackupCreated   EventType = "backup.created"
    EventTypeBackupRestored  EventType = "backup.restored"
    EventTypeConfigChanged   EventType = "config.changed"
    EventTypeSecurityAlert   EventType = "security.alert"
)

// Event 事件结构
type Event struct {
    ID        string                 `json:"id"`
    Type      EventType              `json:"type"`
    Source    string                 `json:"source"`
    Timestamp time.Time              `json:"timestamp"`
    Data      map[string]interface{} `json:"data"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// EventHandler 事件处理器
type EventHandler func(ctx context.Context, event *Event) error
```

## 4. 错误处理 API

### 4.1 错误定义

```go
// AppError 应用程序错误接口
type AppError interface {
    error
    Code() string
    Type() ErrorType
    Details() map[string]interface{}
    Cause() error
}

// ErrorType 错误类型
type ErrorType string

const (
    ErrorTypeValidation   ErrorType = "validation"
    ErrorTypePermission   ErrorType = "permission"
    ErrorTypeFileSystem   ErrorType = "filesystem"
    ErrorTypeNetwork      ErrorType = "network"
    ErrorTypeSystem       ErrorType = "system"
    ErrorTypeInternal     ErrorType = "internal"
)

// 具体错误实现
type appError struct {
    code    string
    errType ErrorType
    message string
    details map[string]interface{}
    cause   error
}

func (e *appError) Error() string {
    return e.message
}

func (e *appError) Code() string {
    return e.code
}

func (e *appError) Type() ErrorType {
    return e.errType
}

func (e *appError) Details() map[string]interface{} {
    return e.details
}

func (e *appError) Cause() error {
    return e.cause
}

// 错误构造函数
func NewValidationError(code, message string, details map[string]interface{}) AppError {
    return &appError{
        code:    code,
        errType: ErrorTypeValidation,
        message: message,
        details: details,
    }
}

func NewPermissionError(code, message string) AppError {
    return &appError{
        code:    code,
        errType: ErrorTypePermission,
        message: message,
    }
}

// 预定义错误代码
const (
    // 基础错误代码
    ErrCodeInvalidIP       = "INVALID_IP"
    ErrCodeInvalidHostname = "INVALID_HOSTNAME"
    ErrCodeProfileNotFound = "PROFILE_NOT_FOUND"
    ErrCodePermissionDenied = "PERMISSION_DENIED"
    ErrCodeFileNotFound    = "FILE_NOT_FOUND"
    ErrCodeBackupFailed    = "BACKUP_FAILED"
    
    // XPC 相关错误代码
    ErrCodeXPCConnectionFailed    = "XPC_CONNECTION_FAILED"
    ErrCodeXPCRequestTimeout      = "XPC_REQUEST_TIMEOUT"
    ErrCodeXPCInvalidRequest      = "XPC_INVALID_REQUEST"
    ErrCodeXPCInvalidResponse     = "XPC_INVALID_RESPONSE"
    ErrCodeXPCAuthenticationFailed = "XPC_AUTHENTICATION_FAILED"
    ErrCodeXPCServiceUnavailable  = "XPC_SERVICE_UNAVAILABLE"
    
    // Helper Tool 相关错误代码
    ErrCodeHelperNotInstalled     = "HELPER_NOT_INSTALLED"
    ErrCodeHelperInstallFailed    = "HELPER_INSTALL_FAILED"
    ErrCodeHelperUninstallFailed  = "HELPER_UNINSTALL_FAILED"
    ErrCodeHelperVersionMismatch  = "HELPER_VERSION_MISMATCH"
    ErrCodeHelperHealthCheckFailed = "HELPER_HEALTH_CHECK_FAILED"
    ErrCodeHelperRestartFailed    = "HELPER_RESTART_FAILED"
    
    // 权限相关错误代码
    ErrCodeInsufficientPrivileges = "INSUFFICIENT_PRIVILEGES"
    ErrCodeSignatureVerificationFailed = "SIGNATURE_VERIFICATION_FAILED"
    ErrCodeCertificateInvalid     = "CERTIFICATE_INVALID"
    ErrCodeAuditLogFailed         = "AUDIT_LOG_FAILED"
)
```

## 5. XPC 通信 API

### 5.1 XPC Client 接口

```go
// XPCClient XPC 客户端接口
type XPCClient interface {
    // Connect 连接到 Helper Tool
    Connect(ctx context.Context) error
    
    // Disconnect 断开连接
    Disconnect(ctx context.Context) error
    
    // IsConnected 检查连接状态
    IsConnected() bool
    
    // WriteHosts 写入 hosts 文件
    WriteHosts(ctx context.Context, req *WriteHostsRequest) (*WriteHostsResponse, error)
    
    // ReadHosts 读取 hosts 文件
    ReadHosts(ctx context.Context, req *ReadHostsRequest) (*ReadHostsResponse, error)
    
    // CreateBackup 创建备份
    CreateBackup(ctx context.Context, req *CreateBackupRequest) (*CreateBackupResponse, error)
    
    // RestoreBackup 恢复备份
    RestoreBackup(ctx context.Context, req *RestoreBackupRequest) (*RestoreBackupResponse, error)
    
    // GetHelperStatus 获取 Helper Tool 状态
    GetHelperStatus(ctx context.Context) (*HelperStatusResponse, error)
}

// WriteHostsRequest 写入 hosts 请求
type WriteHostsRequest struct {
    Content   string            `json:"content"`
    ProfileID string            `json:"profile_id"`
    Metadata  map[string]string `json:"metadata,omitempty"`
}

// WriteHostsResponse 写入 hosts 响应
type WriteHostsResponse struct {
    Success   bool              `json:"success"`
    BackupID  string            `json:"backup_id,omitempty"`
    Timestamp time.Time         `json:"timestamp"`
    Metadata  map[string]string `json:"metadata,omitempty"`
}

// ReadHostsRequest 读取 hosts 请求
type ReadHostsRequest struct {
    IncludeBackup bool              `json:"include_backup,omitempty"`
    Metadata      map[string]string `json:"metadata,omitempty"`
}

// ReadHostsResponse 读取 hosts 响应
type ReadHostsResponse struct {
    Content   string            `json:"content"`
    Timestamp time.Time         `json:"timestamp"`
    Checksum  string            `json:"checksum"`
    Metadata  map[string]string `json:"metadata,omitempty"`
}

// CreateBackupResponse 创建备份响应
type CreateBackupResponse struct {
    BackupID  string            `json:"backup_id"`
    Success   bool              `json:"success"`
    Timestamp time.Time         `json:"timestamp"`
    Metadata  map[string]string `json:"metadata,omitempty"`
}

// RestoreBackupResponse 恢复备份响应
type RestoreBackupResponse struct {
    Success   bool              `json:"success"`
    Timestamp time.Time         `json:"timestamp"`
    Metadata  map[string]string `json:"metadata,omitempty"`
}

// HelperStatusResponse Helper Tool 状态响应
type HelperStatusResponse struct {
    Version     string            `json:"version"`
    Status      string            `json:"status"`
    Uptime      time.Duration     `json:"uptime"`
    LastRequest time.Time         `json:"last_request"`
    Metadata    map[string]string `json:"metadata,omitempty"`
}
```

### 5.2 Helper Tool 服务接口

```go
// HelperService Helper Tool 服务接口
type HelperService interface {
    // Start 启动服务
    Start(ctx context.Context) error
    
    // Stop 停止服务
    Stop(ctx context.Context) error
    
    // HandleRequest 处理请求
    HandleRequest(ctx context.Context, req *XPCRequest) (*XPCResponse, error)
    
    // ValidateRequest 验证请求
    ValidateRequest(ctx context.Context, req *XPCRequest) error
    
    // AuditRequest 审计请求
    AuditRequest(ctx context.Context, req *XPCRequest, resp *XPCResponse) error
}

// XPCRequest XPC 请求结构
type XPCRequest struct {
    ID        string                 `json:"id"`
    Method    string                 `json:"method"`
    Params    map[string]interface{} `json:"params"`
    Timestamp time.Time              `json:"timestamp"`
    ClientID  string                 `json:"client_id"`
    Signature string                 `json:"signature,omitempty"`
}

// XPCResponse XPC 响应结构
type XPCResponse struct {
    ID        string                 `json:"id"`
    Success   bool                   `json:"success"`
    Result    map[string]interface{} `json:"result,omitempty"`
    Error     *XPCError              `json:"error,omitempty"`
    Timestamp time.Time              `json:"timestamp"`
}

// XPCError XPC 错误结构
type XPCError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}
```

### 5.3 Helper Tool 管理接口

```go
// HelperManager Helper Tool 管理器接口
type HelperManager interface {
    // Install 安装 Helper Tool
    Install(ctx context.Context) error
    
    // Uninstall 卸载 Helper Tool
    Uninstall(ctx context.Context) error
    
    // IsInstalled 检查是否已安装
    IsInstalled(ctx context.Context) (bool, error)
    
    // GetVersion 获取版本信息
    GetVersion(ctx context.Context) (string, error)
    
    // CheckHealth 健康检查
    CheckHealth(ctx context.Context) (*HealthStatus, error)
    
    // Restart 重启 Helper Tool
    Restart(ctx context.Context) error
}

// HealthStatus 健康状态
type HealthStatus struct {
    Status      string            `json:"status"`
    Version     string            `json:"version"`
    Uptime      time.Duration     `json:"uptime"`
    LastError   string            `json:"last_error,omitempty"`
    Performance *PerformanceStats `json:"performance,omitempty"`
}

// PerformanceStats 性能统计
type PerformanceStats struct {
    RequestCount    int64         `json:"request_count"`
    AverageLatency  time.Duration `json:"average_latency"`
    ErrorRate       float64       `json:"error_rate"`
    MemoryUsage     int64         `json:"memory_usage"`
    CPUUsage        float64       `json:"cpu_usage"`
}
```

## 6. 日志记录 API

### 6.1 日志接口

```go
// Logger 日志记录器接口
type Logger interface {
    // Debug 调试日志
    Debug(ctx context.Context, msg string, fields ...Field)
    
    // Info 信息日志
    Info(ctx context.Context, msg string, fields ...Field)
    
    // Warn 警告日志
    Warn(ctx context.Context, msg string, fields ...Field)
    
    // Error 错误日志
    Error(ctx context.Context, msg string, err error, fields ...Field)
    
    // WithFields 添加字段
    WithFields(fields ...Field) Logger
    
    // WithContext 添加上下文
    WithContext(ctx context.Context) Logger
}

// Field 日志字段
type Field struct {
    Key   string
    Value interface{}
}

// 字段构造函数
func String(key, value string) Field {
    return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
    return Field{Key: key, Value: value}
}

func Error(err error) Field {
    return Field{Key: "error", Value: err}
}

func Duration(key string, value time.Duration) Field {
    return Field{Key: key, Value: value}
}
```

## 7. 使用示例

### 7.1 Profile 管理示例

```go
func ExampleProfileUsage() {
    ctx := context.Background()
    
    // 创建 Profile Manager
    profileManager := NewProfileManager()
    
    // 创建新 Profile
    req := &CreateProfileRequest{
        Name:        "Development",
        Description: "Development environment hosts",
    }
    
    profile, err := profileManager.CreateProfile(ctx, req)
    if err != nil {
        log.Error(ctx, "Failed to create profile", err)
        return
    }
    
    // 添加 Host 条目
    profile.Entries = append(profile.Entries, HostEntry{
        ID:       "entry1",
        IP:       "127.0.0.1",
        Hostname: "api.dev.local",
        Comment:  "Development API server",
        Enabled:  true,
    })
    
    // 更新 Profile
    updateReq := &UpdateProfileRequest{
        ID:      profile.ID,
        Name:    profile.Name,
        Description: profile.Description,
        Entries: profile.Entries,
    }
    
    _, err = profileManager.UpdateProfile(ctx, updateReq)
    if err != nil {
        log.Error(ctx, "Failed to update profile", err)
        return
    }
    
    // 应用 Profile
    err = profileManager.ApplyProfile(ctx, profile.ID)
    if err != nil {
        log.Error(ctx, "Failed to apply profile", err)
        return
    }
    
    log.Info(ctx, "Profile applied successfully", String("profile_id", profile.ID))
}
```

### 7.2 XPC 通信示例

```go
func ExampleXPCUsage() {
    ctx := context.Background()
    
    // 创建 XPC 客户端
    xpcClient := NewXPCClient()
    
    // 连接到 Helper Tool
    err := xpcClient.Connect(ctx)
    if err != nil {
        log.Error(ctx, "Failed to connect to Helper Tool", err)
        return
    }
    defer xpcClient.Disconnect(ctx)
    
    // 读取当前 hosts 文件
    readReq := &ReadHostsRequest{
        IncludeBackup: true,
    }
    
    readResp, err := xpcClient.ReadHosts(ctx, readReq)
    if err != nil {
        log.Error(ctx, "Failed to read hosts file", err)
        return
    }
    
    log.Info(ctx, "Current hosts content", 
        String("checksum", readResp.Checksum),
        String("timestamp", readResp.Timestamp.String()))
    
    // 写入新的 hosts 内容
    writeReq := &WriteHostsRequest{
        Content:   "127.0.0.1 localhost\n127.0.0.1 api.dev.local",
        ProfileID: "dev-profile",
        Metadata: map[string]string{
            "source": "mHost",
            "version": "1.0.0",
        },
    }
    
    writeResp, err := xpcClient.WriteHosts(ctx, writeReq)
    if err != nil {
        log.Error(ctx, "Failed to write hosts file", err)
        return
    }
    
    log.Info(ctx, "Hosts file updated successfully",
        String("backup_id", writeResp.BackupID),
        String("timestamp", writeResp.Timestamp.String()))
}
```

### 7.3 Helper Tool 管理示例

```go
func ExampleHelperManagement() {
    ctx := context.Background()
    
    // 创建 Helper Manager
    helperManager := NewHelperManager()
    
    // 检查 Helper Tool 是否已安装
    installed, err := helperManager.IsInstalled(ctx)
    if err != nil {
        log.Error(ctx, "Failed to check Helper Tool status", err)
        return
    }
    
    if !installed {
        // 安装 Helper Tool
        err = helperManager.Install(ctx)
        if err != nil {
            log.Error(ctx, "Failed to install Helper Tool", err)
            return
        }
        log.Info(ctx, "Helper Tool installed successfully")
    }
    
    // 健康检查
    health, err := helperManager.CheckHealth(ctx)
    if err != nil {
        log.Error(ctx, "Helper Tool health check failed", err)
        return
    }
    
    log.Info(ctx, "Helper Tool status",
        String("status", health.Status),
        String("version", health.Version),
        Duration("uptime", health.Uptime))
    
    // 如果状态异常，尝试重启
    if health.Status != "healthy" {
        err = helperManager.Restart(ctx)
        if err != nil {
            log.Error(ctx, "Failed to restart Helper Tool", err)
            return
        }
        log.Info(ctx, "Helper Tool restarted successfully")
    }
}
```

### 7.4 错误处理示例

```go
func ExampleErrorHandling() {
    ctx := context.Background()
    
    // 验证 Host 条目
    entry := &HostEntry{
        IP:       "invalid-ip",
        Hostname: "example.com",
    }
    
    hostManager := NewHostManager()
    err := hostManager.ValidateEntry(ctx, entry)
    
    if err != nil {
        // 检查错误类型
        if appErr, ok := err.(AppError); ok {
            switch appErr.Type() {
            case ErrorTypeValidation:
                log.Warn(ctx, "Validation error", 
                    String("code", appErr.Code()),
                    String("field", appErr.Details()["field"].(string)))
                
            case ErrorTypePermission:
                log.Error(ctx, "Permission denied", err)
                
            default:
                log.Error(ctx, "Unknown error", err)
            }
        }
    }
}
```

## 8. 测试接口

### 8.1 Mock 接口

```go
// MockProfileManager Profile Manager 的 Mock 实现
type MockProfileManager struct {
    profiles map[string]*Profile
    active   string
}

func NewMockProfileManager() *MockProfileManager {
    return &MockProfileManager{
        profiles: make(map[string]*Profile),
    }
}

func (m *MockProfileManager) CreateProfile(ctx context.Context, req *CreateProfileRequest) (*Profile, error) {
    profile := &Profile{
        ID:          generateID(),
        Name:        req.Name,
        Description: req.Description,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
        Entries:     []HostEntry{},
    }
    
    m.profiles[profile.ID] = profile
    return profile, nil
}

// 其他方法的 Mock 实现...
```

### 8.2 XPC Mock 接口

```go
// MockXPCClient XPC 客户端的 Mock 实现
type MockXPCClient struct {
    connected bool
    responses map[string]interface{}
    errors    map[string]error
}

func NewMockXPCClient() *MockXPCClient {
    return &MockXPCClient{
        responses: make(map[string]interface{}),
        errors:    make(map[string]error),
    }
}

func (m *MockXPCClient) Connect(ctx context.Context) error {
    if err, exists := m.errors["Connect"]; exists {
        return err
    }
    m.connected = true
    return nil
}

func (m *MockXPCClient) WriteHosts(ctx context.Context, req *WriteHostsRequest) (*WriteHostsResponse, error) {
    if err, exists := m.errors["WriteHosts"]; exists {
        return nil, err
    }
    
    if resp, exists := m.responses["WriteHosts"]; exists {
        return resp.(*WriteHostsResponse), nil
    }
    
    return &WriteHostsResponse{
        Success:   true,
        BackupID:  "mock-backup-id",
        Timestamp: time.Now(),
    }, nil
}

// SetError 设置 Mock 错误
func (m *MockXPCClient) SetError(method string, err error) {
    m.errors[method] = err
}

// SetResponse 设置 Mock 响应
func (m *MockXPCClient) SetResponse(method string, response interface{}) {
    m.responses[method] = response
}
```

### 8.3 测试工具

```go
// TestHelper 测试辅助工具
type TestHelper struct {
    tempDir string
    cleanup []func()
}

func NewTestHelper() *TestHelper {
    tempDir, _ := os.MkdirTemp("", "mhost-test-*")
    return &TestHelper{
        tempDir: tempDir,
        cleanup: []func(){},
    }
}

func (h *TestHelper) CreateTempFile(content string) string {
    file, _ := os.CreateTemp(h.tempDir, "test-*")
    file.WriteString(content)
    file.Close()
    
    h.cleanup = append(h.cleanup, func() {
        os.Remove(file.Name())
    })
    
    return file.Name()
}

func (h *TestHelper) Cleanup() {
    for _, fn := range h.cleanup {
        fn()
    }
    os.RemoveAll(h.tempDir)
}
```

这个 API 设计文档为 mHost 项目提供了完整的接口规范，确保了各模块间的清晰边界和一致的交互方式，为后续的开发、测试和维护奠定了坚实的基础。