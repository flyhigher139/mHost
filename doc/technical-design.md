# mHost 技术设计文档

## 1. 系统架构

### 1.1 整体架构

#### 基础架构
```
┌─────────────────────────────────────────────────────────────┐
│                        UI Layer (Fyne)                     │
├─────────────────────────────────────────────────────────────┤
│                     Application Layer                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │   Profile   │  │    Host     │  │      Config         │ │
│  │  Manager    │  │   Manager   │  │     Manager         │ │
│  └─────────────┘  └─────────────┘  └─────────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│                      Core Layer                            │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │    File     │  │   Backup    │  │      Security       │ │
│  │   Handler   │  │   Manager   │  │     Manager         │ │
│  └─────────────┘  └─────────────┘  └─────────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│                     System Layer                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │   macOS     │  │    File     │  │      Logger         │ │
│  │   System    │  │   System    │  │                     │ │
│  └─────────────┘  └─────────────┘  └─────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

#### Helper Tool 架构（推荐实现）
```
┌─────────────────────────────────────────────────────────────┐
│                    mHost.app (用户权限)                     │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │                   UI Layer (Fyne)                      │ │
│  ├─────────────────────────────────────────────────────────┤ │
│  │                 Application Layer                      │ │
│  │ ┌─────────────┐ ┌─────────────┐ ┌─────────────────────┐ │ │
│  │ │   Profile   │ │    Host     │ │      Config         │ │ │
│  │ │  Manager    │ │   Manager   │ │     Manager         │ │ │
│  │ └─────────────┘ └─────────────┘ └─────────────────────┘ │ │
│  ├─────────────────────────────────────────────────────────┤ │
│  │                    XPC Client                          │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                                │
                            XPC 通信
                                │
┌─────────────────────────────────────────────────────────────┐
│                HostsHelper (特权 Helper Tool)               │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │                    XPC Server                          │ │
│  ├─────────────────────────────────────────────────────────┤ │
│  │                  Security Layer                        │ │
│  │ ┌─────────────┐ ┌─────────────┐ ┌─────────────────────┐ │ │
│  │ │   Request   │ │   Identity  │ │    Parameter        │ │ │
│  │ │ Validation  │ │ Verification│ │   Validation        │ │ │
│  │ └─────────────┘ └─────────────┘ └─────────────────────┘ │ │
│  ├─────────────────────────────────────────────────────────┤ │
│  │                 Operation Layer                        │ │
│  │ ┌─────────────┐ ┌─────────────┐ ┌─────────────────────┐ │ │
│  │ │    Hosts    │ │   Backup    │ │      Audit          │ │ │
│  │ │   Handler   │ │   Manager   │ │     Logger          │ │ │
│  │ └─────────────┘ └─────────────┘ └─────────────────────┘ │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 模块职责

#### UI Layer
- **主窗口**: 应用程序主界面
- **Profile 面板**: Profile 列表和管理
- **编辑面板**: Host 条目编辑
- **对话框**: 各种操作对话框

#### Application Layer
- **Profile Manager**: Profile 的 CRUD 操作
- **Host Manager**: Host 条目的管理和验证
- **Config Manager**: 应用配置管理
- **XPC Client**: 与 Helper Tool 通信的客户端，负责发送操作请求

#### Core Layer
- **File Handler**: 文件读写操作
- **Backup Manager**: 备份和恢复功能
- **Security Manager**: 权限管理和安全检查

#### System Layer
- **macOS System**: 系统调用和权限管理
- **File System**: 底层文件操作
- **Logger**: 日志记录

#### Helper Tool (HostsHelper)
- **XPC Server**: 接收来自主应用的请求，提供特权操作服务
- **Security Layer**: 
  - **Request Validation**: 验证请求的合法性和完整性
  - **Identity Verification**: 验证请求来源的身份
  - **Parameter Validation**: 验证请求参数的安全性
- **Operation Layer**:
  - **Hosts Handler**: 执行实际的 hosts 文件操作
  - **Backup Manager**: 管理 hosts 文件备份
  - **Audit Logger**: 记录所有特权操作的审计日志

## 2. 数据模型

### 2.1 Profile 数据结构

```go
// Profile 表示一个 host 配置文件
type Profile struct {
    ID          string    `json:"id"`          // 唯一标识符
    Name        string    `json:"name"`        // Profile 名称
    Description string    `json:"description"` // 描述信息
    CreatedAt   time.Time `json:"created_at"`  // 创建时间
    UpdatedAt   time.Time `json:"updated_at"`  // 更新时间
    IsActive    bool      `json:"is_active"`   // 是否为当前激活的 Profile
    Entries     []HostEntry `json:"entries"`   // Host 条目列表
}

// HostEntry 表示一个 host 条目
type HostEntry struct {
    ID       string `json:"id"`       // 唯一标识符
    IP       string `json:"ip"`       // IP 地址
    Hostname string `json:"hostname"` // 主机名/域名
    Comment  string `json:"comment"`  // 注释
    Enabled  bool   `json:"enabled"`  // 是否启用
}
```

### 2.2 应用配置结构

```go
// AppConfig 应用程序配置
type AppConfig struct {
    Version        string      `json:"version"`         // 配置版本
    LastProfile    string      `json:"last_profile"`    // 最后使用的 Profile ID
    WindowConfig   WindowConfig `json:"window_config"`   // 窗口配置
    BackupConfig   BackupConfig `json:"backup_config"`   // 备份配置
    LoggingConfig  LogConfig   `json:"logging_config"`  // 日志配置
}

// WindowConfig 窗口配置
type WindowConfig struct {
    Width    int  `json:"width"`     // 窗口宽度
    Height   int  `json:"height"`    // 窗口高度
    X        int  `json:"x"`         // 窗口 X 坐标
    Y        int  `json:"y"`         // 窗口 Y 坐标
    Maximized bool `json:"maximized"` // 是否最大化
}

// BackupConfig 备份配置
type BackupConfig struct {
    MaxBackups    int  `json:"max_backups"`     // 最大备份数量
    AutoBackup    bool `json:"auto_backup"`     // 是否自动备份
    BackupOnApply bool `json:"backup_on_apply"` // 应用时是否备份
}
```

## 3. 核心组件设计

### 3.1 Profile Manager

```go
// ProfileManager Profile 管理器接口
type ProfileManager interface {
    // 创建新的 Profile
    CreateProfile(name, description string) (*Profile, error)
    
    // 获取所有 Profile
    GetAllProfiles() ([]*Profile, error)
    
    // 根据 ID 获取 Profile
    GetProfile(id string) (*Profile, error)
    
    // 更新 Profile
    UpdateProfile(profile *Profile) error
    
    // 删除 Profile
    DeleteProfile(id string) error
    
    // 应用 Profile 到系统
    ApplyProfile(id string) error
    
    // 获取当前激活的 Profile
    GetActiveProfile() (*Profile, error)
}

// ProfileManagerImpl Profile 管理器实现
type ProfileManagerImpl struct {
    profiles    map[string]*Profile
    currentID   string
    configPath  string
    eventBus    *EventBus
}

func (pm *ProfileManagerImpl) CreateProfile(name, description string) (*Profile, error)
func (pm *ProfileManagerImpl) DeleteProfile(id string) error
func (pm *ProfileManagerImpl) SwitchProfile(id string) error
func (pm *ProfileManagerImpl) GetProfile(id string) *Profile
func (pm *ProfileManagerImpl) ListProfiles() []*Profile
func (pm *ProfileManagerImpl) ExportProfile(id, filePath string) error
func (pm *ProfileManagerImpl) ImportProfile(filePath string) (*Profile, error)
```

### 3.2 Host Manager

```go
// HostManager Host 管理器接口
type HostManager interface {
    // 读取系统 hosts 文件
    ReadSystemHosts() ([]HostEntry, error)
    
    // 写入 hosts 文件
    WriteSystemHosts(entries []HostEntry) error
    
    // 验证 Host 条目
    ValidateEntry(entry *HostEntry) error
    
    // 解析 hosts 文件内容
    ParseHostsContent(content string) ([]HostEntry, error)
    
    // 生成 hosts 文件内容
    GenerateHostsContent(entries []HostEntry) string
}

// HostManagerImpl Host 管理器实现
type HostManagerImpl struct {
    hostsPath   string
    backupMgr   *BackupManager
    xpcClient   *XPCClient
    eventBus    *EventBus
}

func (hm *HostManagerImpl) ReadHosts() ([]HostEntry, error)
func (hm *HostManagerImpl) WriteHosts(entries []HostEntry) error
func (hm *HostManagerImpl) ApplyProfile(profile *Profile) error
func (hm *HostManagerImpl) ValidateEntries(entries []HostEntry) error
func (hm *HostManagerImpl) BackupCurrent() error
func (hm *HostManagerImpl) RestoreBackup(backupID string) error
```

### 3.3 XPC Client

```go
// XPCClient XPC 客户端，负责与 Helper Tool 通信
type XPCClient struct {
    connection  *xpc.Connection
    serviceName string
    timeout     time.Duration
    logger      *Logger
}

// XPCRequest XPC 请求结构
type XPCRequest struct {
    Operation   string                 `json:"operation"`
    Parameters  map[string]interface{} `json:"parameters"`
    RequestID   string                 `json:"request_id"`
    Timestamp   time.Time              `json:"timestamp"`
}

// XPCResponse XPC 响应结构
type XPCResponse struct {
    Success     bool                   `json:"success"`
    Data        map[string]interface{} `json:"data"`
    Error       string                 `json:"error"`
    RequestID   string                 `json:"request_id"`
}

func (xc *XPCClient) Connect() error
func (xc *XPCClient) Disconnect() error
func (xc *XPCClient) SendRequest(req *XPCRequest) (*XPCResponse, error)
func (xc *XPCClient) WriteHosts(entries []HostEntry) error
func (xc *XPCClient) BackupHosts() (string, error)
func (xc *XPCClient) RestoreHosts(backupID string) error
func (xc *XPCClient) InstallHelper() error
func (xc *XPCClient) UninstallHelper() error
func (xc *XPCClient) CheckHelperStatus() (bool, error)
```

### 3.4 Security Manager

```go
// SecurityManager 安全管理器接口
type SecurityManager interface {
    // 检查当前用户权限
    CheckPermissions() error
    
    // 请求管理员权限
    RequestAdminPermission() error
    
    // 验证文件访问权限
    ValidateFileAccess(path string) error
    
    // 安全地执行需要权限的操作
    ExecuteWithPermission(operation func() error) error
}

// SecurityManagerImpl 安全管理器实现
type SecurityManagerImpl struct {
    helperInstalled bool
    xpcClient      *XPCClient
    fallbackAuth   *FallbackAuth
}

// FallbackAuth 回退认证机制
type FallbackAuth struct {
    authCache   map[string]time.Time
    cacheExpiry time.Duration
}

func (sm *SecurityManagerImpl) RequestPermission() error
func (sm *SecurityManagerImpl) HasPermission() bool
func (sm *SecurityManagerImpl) ValidateOperation(op Operation) error
func (sm *SecurityManagerImpl) UseHelperTool() bool
func (sm *SecurityManagerImpl) InstallHelperTool() error
func (sm *SecurityManagerImpl) UninstallHelperTool() error
func (sm *SecurityManagerImpl) CheckHelperStatus() (bool, error)
```

### 3.5 Helper Tool (HostsHelper)

```go
// HostsHelper 特权 Helper Tool 主结构
type HostsHelper struct {
    xpcServer    *XPCServer
    securityMgr  *HelperSecurityManager
    hostsHandler *HelperHostsHandler
    auditLogger  *AuditLogger
}

// XPCServer XPC 服务器
type XPCServer struct {
    listener    *xpc.Listener
    serviceName string
    logger      *Logger
}

// HelperSecurityManager Helper Tool 安全管理器
type HelperSecurityManager struct {
    allowedClients map[string]bool
    auditLogger    *AuditLogger
}

// HelperHostsHandler Helper Tool hosts 文件处理器
type HelperHostsHandler struct {
    hostsPath   string
    backupMgr   *HelperBackupManager
    validator   *HostsValidator
}

func (hh *HostsHelper) Start() error
func (hh *HostsHelper) Stop() error
func (hh *HostsHelper) HandleRequest(req *XPCRequest) *XPCResponse

func (hs *XPCServer) Listen() error
func (hs *XPCServer) HandleConnection(conn *xpc.Connection)

func (hsm *HelperSecurityManager) ValidateClient(clientID string) bool
func (hsm *HelperSecurityManager) ValidateRequest(req *XPCRequest) error
func (hsm *HelperSecurityManager) LogOperation(op string, params map[string]interface{})

func (hhh *HelperHostsHandler) WriteHosts(entries []HostEntry) error
func (hhh *HelperHostsHandler) BackupHosts() (string, error)
func (hhh *HelperHostsHandler) RestoreHosts(backupID string) error
func (hhh *HelperHostsHandler) ValidateHosts(entries []HostEntry) error
```

## 4. UI 组件设计

### 4.1 主窗口结构

```go
// MainWindow 主窗口结构
type MainWindow struct {
    window       fyne.Window
    profileList  *ProfileListWidget
    editPanel    *EditPanelWidget
    statusBar    *StatusBarWidget
    menuBar      *fyne.MainMenu
    
    profileManager ProfileManager
    hostManager    HostManager
    configManager  ConfigManager
}

// 初始化主窗口
func NewMainWindow(app fyne.App) *MainWindow {
    // 实现窗口初始化逻辑
}
```

### 4.2 Profile 列表组件

```go
// ProfileListWidget Profile 列表组件
type ProfileListWidget struct {
    container    *container.VBox
    profileList  *widget.List
    addButton    *widget.Button
    searchEntry  *widget.Entry
    
    profiles     []*Profile
    selectedProfile *Profile
    onSelectionChanged func(*Profile)
}
```

### 4.3 编辑面板组件

```go
// EditPanelWidget 编辑面板组件
type EditPanelWidget struct {
    container     *container.VBox
    entriesTable  *widget.Table
    addButton     *widget.Button
    deleteButton  *widget.Button
    applyButton   *widget.Button
    
    currentProfile *Profile
    entries        []HostEntry
    onEntriesChanged func([]HostEntry)
}
```

## 5. 文件系统设计

### 5.1 目录结构

```
~/Library/Application Support/mHost/
├── config.json              # 应用配置文件
├── profiles/                # Profile 存储目录
│   ├── profile1.json
│   ├── profile2.json
│   └── ...
├── backups/                 # 备份文件目录
│   ├── hosts_backup_20231201_120000.txt
│   ├── hosts_backup_20231201_130000.txt
│   └── ...
└── logs/                    # 日志文件目录
    ├── app.log
    ├── error.log
    └── operation.log
```

### 5.2 文件操作接口

```go
// FileHandler 文件处理器接口
type FileHandler interface {
    // 读取文件内容
    ReadFile(path string) ([]byte, error)
    
    // 写入文件内容
    WriteFile(path string, data []byte) error
    
    // 检查文件是否存在
    FileExists(path string) bool
    
    // 创建目录
    CreateDir(path string) error
    
    // 复制文件
    CopyFile(src, dst string) error
    
    // 删除文件
    DeleteFile(path string) error
}
```

## 6. 错误处理策略

### 6.1 错误类型定义

```go
// 自定义错误类型
type ErrorType int

const (
    ErrorTypePermission ErrorType = iota
    ErrorTypeFileNotFound
    ErrorTypeInvalidFormat
    ErrorTypeNetworkError
    ErrorTypeSystemError
)

// AppError 应用程序错误
type AppError struct {
    Type    ErrorType
    Message string
    Cause   error
}

func (e *AppError) Error() string {
    return e.Message
}
```

### 6.2 错误处理流程

1. **捕获错误**: 在各个层级捕获可能的错误
2. **错误分类**: 根据错误类型进行分类处理
3. **用户反馈**: 向用户显示友好的错误信息
4. **日志记录**: 记录详细的错误信息用于调试
5. **恢复机制**: 在可能的情况下自动恢复

## 7. 安全性设计

### 7.1 权限管理

#### 基础权限管理
- **最小权限原则**: 只在需要时请求管理员权限
- **权限检查**: 在每次操作前检查必要权限
- **安全提示**: 向用户明确说明权限用途

#### Helper Tool 权限管理（推荐）
- **一次性授权**: 首次安装时获取管理员权限，后续操作无需重复授权
- **特权分离**: 主应用运行在用户权限，Helper Tool 运行在特权模式
- **安全通信**: 通过 XPC 进行安全的进程间通信
- **身份验证**: Helper Tool 验证请求来源的合法性

### 7.2 Helper Tool 安全机制

#### 安装和验证
```go
type HelperInstaller struct {
    bundleID     string
    helperPath   string
    launchdPlist string
}

// 安装 Helper Tool
func (hi *HelperInstaller) Install() error {
    // 1. 验证 Helper Tool 签名
    if err := hi.verifySignature(); err != nil {
        return err
    }
    
    // 2. 复制到系统目录
    if err := hi.copyToSystemDirectory(); err != nil {
        return err
    }
    
    // 3. 注册 launchd 服务
    if err := hi.registerLaunchdService(); err != nil {
        return err
    }
    
    return nil
}
```

#### 请求验证
```go
type SecurityValidator struct {
    allowedOperations map[string]bool
    clientWhitelist   map[string]bool
}

// 验证请求安全性
func (sv *SecurityValidator) ValidateRequest(req *XPCRequest, clientID string) error {
    // 1. 验证客户端身份
    if !sv.clientWhitelist[clientID] {
        return errors.New("unauthorized client")
    }
    
    // 2. 验证操作权限
    if !sv.allowedOperations[req.Operation] {
        return errors.New("unauthorized operation")
    }
    
    // 3. 验证参数安全性
    if err := sv.validateParameters(req.Parameters); err != nil {
        return err
    }
    
    return nil
}
```

#### 审计日志
```go
type AuditLogger struct {
    logPath string
    logger  *log.Logger
}

// 记录操作审计
func (al *AuditLogger) LogOperation(operation string, clientID string, params map[string]interface{}) {
    entry := AuditEntry{
        Timestamp: time.Now(),
        Operation: operation,
        ClientID:  clientID,
        Parameters: params,
        Success:   true,
    }
    
    al.writeLog(entry)
}
```

### 7.3 数据验证

```go
// Validator 数据验证器
type Validator struct{}

// 验证 IP 地址格式
func (v *Validator) ValidateIP(ip string) error {
    if net.ParseIP(ip) == nil {
        return errors.New("invalid IP address format")
    }
    return nil
}

// 验证域名格式
func (v *Validator) ValidateHostname(hostname string) error {
    // 实现域名格式验证逻辑
    return nil
}

// 验证 Profile 名称
func (v *Validator) ValidateProfileName(name string) error {
    // 实现 Profile 名称验证逻辑
    return nil
}
```

### 7.4 备份策略

- **自动备份**: 在修改 hosts 文件前自动创建备份
- **版本管理**: 保留多个版本的备份文件
- **完整性检查**: 验证备份文件的完整性
- **快速恢复**: 提供一键恢复功能
- **备份加密**: 对敏感备份数据进行加密存储

## 8. 性能优化

### 8.1 内存管理

- **延迟加载**: 只在需要时加载 Profile 数据
- **缓存机制**: 缓存常用的 Profile 和配置
- **资源释放**: 及时释放不再使用的资源

### 8.2 文件操作优化

- **批量操作**: 减少文件 I/O 次数
- **异步处理**: 对于耗时操作使用异步处理
- **错误重试**: 实现智能的重试机制

## 9. 测试策略

### 9.1 单元测试

#### 基础组件测试
```go
// Profile 管理测试
func TestProfileManager_CreateProfile(t *testing.T) {
    pm := NewProfileManager("test_config")
    profile, err := pm.CreateProfile("Test Profile", "Test Description")
    
    assert.NoError(t, err)
    assert.NotNil(t, profile)
    assert.Equal(t, "Test Profile", profile.Name)
}

// Host 条目验证测试
func TestHostEntry_Validate(t *testing.T) {
    entry := HostEntry{
        IP:       "192.168.1.1",
        Hostname: "test.local",
        Comment:  "Test entry",
        Enabled:  true,
    }
    
    err := entry.Validate()
    assert.NoError(t, err)
}
```

#### XPC 通信测试
```go
// XPC Client 测试
func TestXPCClient_SendRequest(t *testing.T) {
    client := NewXPCClient("com.test.helper")
    
    req := &XPCRequest{
        Operation:  "test_operation",
        Parameters: map[string]interface{}{"key": "value"},
        RequestID:  "test-123",
        Timestamp:  time.Now(),
    }
    
    // 模拟 XPC 响应
    mockResponse := &XPCResponse{
        Success:   true,
        Data:      map[string]interface{}{"result": "ok"},
        RequestID: "test-123",
    }
    
    // 测试请求发送
    response, err := client.SendRequest(req)
    assert.NoError(t, err)
    assert.Equal(t, mockResponse.Success, response.Success)
}

// Helper Tool 安全验证测试
func TestHelperSecurityManager_ValidateRequest(t *testing.T) {
    hsm := NewHelperSecurityManager()
    
    // 测试有效请求
    validReq := &XPCRequest{
        Operation:  "write_hosts",
        Parameters: map[string]interface{}{"entries": []HostEntry{}},
        RequestID:  "valid-123",
    }
    
    err := hsm.ValidateRequest(validReq, "authorized-client")
    assert.NoError(t, err)
    
    // 测试无效操作
    invalidReq := &XPCRequest{
        Operation: "malicious_operation",
        RequestID: "invalid-123",
    }
    
    err = hsm.ValidateRequest(invalidReq, "authorized-client")
    assert.Error(t, err)
}
```

### 9.2 集成测试

#### 完整流程测试
```go
// 基础流程测试
func TestCompleteWorkflow(t *testing.T) {
    // 1. 创建 Profile
    pm := NewProfileManager("test_config")
    profile, err := pm.CreateProfile("Integration Test", "")
    require.NoError(t, err)
    
    // 2. 添加 Host 条目
    entry := HostEntry{
        IP:       "127.0.0.1",
        Hostname: "test.local",
        Enabled:  true,
    }
    profile.AddEntry(entry)
    
    // 3. 应用 Profile
    hm := NewHostManager("/tmp/test_hosts")
    err = hm.ApplyProfile(profile)
    require.NoError(t, err)
    
    // 4. 验证结果
    content, err := ioutil.ReadFile("/tmp/test_hosts")
    require.NoError(t, err)
    assert.Contains(t, string(content), "127.0.0.1 test.local")
}

// Helper Tool 集成测试
func TestHelperToolIntegration(t *testing.T) {
    // 1. 启动测试 Helper Tool
    helper := NewTestHelperTool()
    err := helper.Start()
    require.NoError(t, err)
    defer helper.Stop()
    
    // 2. 创建 XPC 客户端
    client := NewXPCClient("com.test.helper")
    err = client.Connect()
    require.NoError(t, err)
    defer client.Disconnect()
    
    // 3. 测试 hosts 文件写入
    entries := []HostEntry{
        {IP: "127.0.0.1", Hostname: "test.local", Enabled: true},
    }
    
    err = client.WriteHosts(entries)
    assert.NoError(t, err)
    
    // 4. 验证写入结果
    content, err := ioutil.ReadFile("/tmp/test_hosts")
    require.NoError(t, err)
    assert.Contains(t, string(content), "127.0.0.1 test.local")
}
```

### 9.3 安全测试

```go
// 权限验证测试
func TestSecurityValidation(t *testing.T) {
    hsm := NewHelperSecurityManager()
    
    // 测试未授权客户端
    req := &XPCRequest{
        Operation: "write_hosts",
        RequestID: "test-123",
    }
    
    err := hsm.ValidateRequest(req, "unauthorized-client")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "unauthorized client")
    
    // 测试参数注入攻击
    maliciousReq := &XPCRequest{
        Operation: "write_hosts",
        Parameters: map[string]interface{}{
            "entries": "../../../etc/passwd",
        },
        RequestID: "malicious-123",
    }
    
    err = hsm.ValidateRequest(maliciousReq, "authorized-client")
    assert.Error(t, err)
}

// 审计日志测试
func TestAuditLogging(t *testing.T) {
    logger := NewAuditLogger("/tmp/test_audit.log")
    
    // 记录操作
    logger.LogOperation("write_hosts", "test-client", map[string]interface{}{
        "entries_count": 5,
    })
    
    // 验证日志记录
    content, err := ioutil.ReadFile("/tmp/test_audit.log")
    require.NoError(t, err)
    assert.Contains(t, string(content), "write_hosts")
    assert.Contains(t, string(content), "test-client")
}
```

### 9.4 UI 测试

```go
// Fyne UI 测试
func TestMainWindow(t *testing.T) {
    app := test.NewApp()
    window := app.NewWindow("Test")
    
    // 创建主界面
    mainWindow := NewMainWindow(window)
    mainWindow.Show()
    
    // 测试界面元素
    assert.NotNil(t, mainWindow.profileList)
    assert.NotNil(t, mainWindow.editPanel)
}

// Helper Tool 状态显示测试
func TestHelperToolStatusUI(t *testing.T) {
    app := test.NewApp()
    window := app.NewWindow("Test")
    
    // 创建设置界面
    settingsDialog := NewSettingsDialog(window)
    
    // 测试 Helper Tool 状态显示
    settingsDialog.UpdateHelperStatus(true)
    assert.True(t, settingsDialog.helperStatusLabel.Text == "已安装")
    
    settingsDialog.UpdateHelperStatus(false)
    assert.True(t, settingsDialog.helperStatusLabel.Text == "未安装")
}
```

### 9.5 性能测试

```go
// XPC 通信性能测试
func BenchmarkXPCCommunication(b *testing.B) {
    client := NewXPCClient("com.test.helper")
    client.Connect()
    defer client.Disconnect()
    
    req := &XPCRequest{
        Operation:  "ping",
        Parameters: map[string]interface{}{},
        RequestID:  "bench-test",
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := client.SendRequest(req)
        if err != nil {
            b.Fatal(err)
        }
    }
}

// 大量 hosts 条目处理性能测试
func BenchmarkLargeHostsFile(b *testing.B) {
    // 生成大量 hosts 条目
    entries := make([]HostEntry, 10000)
    for i := 0; i < 10000; i++ {
        entries[i] = HostEntry{
            IP:       fmt.Sprintf("192.168.%d.%d", i/256, i%256),
            Hostname: fmt.Sprintf("host%d.local", i),
            Enabled:  true,
        }
    }
    
    client := NewXPCClient("com.test.helper")
    client.Connect()
    defer client.Disconnect()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        err := client.WriteHosts(entries)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### 9.6 系统测试

- **兼容性测试**: 测试不同 macOS 版本的兼容性
- **Helper Tool 安装测试**: 测试 Helper Tool 的安装和卸载流程
- **权限升级测试**: 测试权限请求和管理功能
- **安全沙箱测试**: 测试应用在沙箱环境下的行为

## 10. 部署和分发

### 10.1 构建流程

#### 基础构建
```bash
# 1. 依赖检查
go mod verify

# 2. 代码检查
go vet ./...
golint ./...

# 3. 运行测试
go test ./...

# 4. 构建主应用
go build -ldflags "-X main.version=$(git describe --tags)" -o mHost ./cmd/mhost

# 5. 构建 Helper Tool
go build -ldflags "-X main.version=$(git describe --tags)" -o HostsHelper ./cmd/helper
```

#### 应用打包
```bash
# 创建应用包结构
mkdir -p mHost.app/Contents/{MacOS,Resources,Library/LaunchServices}

# 复制主应用
cp mHost mHost.app/Contents/MacOS/

# 复制 Helper Tool
cp HostsHelper mHost.app/Contents/Library/LaunchServices/
cp com.yourcompany.mhost.helper.plist mHost.app/Contents/Library/LaunchServices/

# 复制资源文件
cp Info.plist mHost.app/Contents/
cp assets/icon.icns mHost.app/Contents/Resources/

# 设置权限
chmod 755 mHost.app/Contents/MacOS/mHost
chmod 755 mHost.app/Contents/Library/LaunchServices/HostsHelper
```

#### 完整应用结构
```
mHost.app/
├── Contents/
│   ├── Info.plist
│   ├── MacOS/
│   │   └── mHost                    # 主应用
│   ├── Resources/
│   │   ├── icon.icns
│   │   └── profiles/                # 默认 Profile
│   └── Library/
│       └── LaunchServices/
│           ├── HostsHelper          # Helper Tool 可执行文件
│           └── com.yourcompany.mhost.helper.plist
```

### 10.2 代码签名

#### Helper Tool 权限配置
```xml
<!-- helper.entitlements -->
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>com.apple.security.app-sandbox</key>
    <false/>
    <key>com.apple.security.files.user-selected.read-write</key>
    <true/>
    <key>com.apple.security.files.downloads.read-write</key>
    <true/>
</dict>
</plist>
```

#### 签名流程
```bash
# 1. 签名 Helper Tool（需要特殊权限）
codesign --force --verify --verbose \
  --sign "Developer ID Application: Your Name" \
  --entitlements helper.entitlements \
  mHost.app/Contents/Library/LaunchServices/HostsHelper

# 2. 签名主应用
codesign --force --verify --verbose \
  --sign "Developer ID Application: Your Name" \
  --entitlements app.entitlements \
  mHost.app

# 3. 验证签名
codesign --verify --deep --strict --verbose=2 mHost.app
spctl -a -t exec -vv mHost.app
```

- **开发者证书**: 使用有效的 Apple 开发者证书
- **公证流程**: 通过 Apple 公证流程
- **安全检查**: 确保应用通过 macOS 安全检查

### 10.3 公证流程

```bash
# 创建 DMG
hdiutil create -volname "mHost" -srcfolder mHost.app -ov -format UDZO mHost.dmg

# 上传公证
xcrun altool --notarize-app \
  --primary-bundle-id "com.yourcompany.mhost" \
  --username "your-apple-id" \
  --password "@keychain:AC_PASSWORD" \
  --file mHost.dmg

# 检查公证状态
xcrun altool --notarization-info <RequestUUID> \
  --username "your-apple-id" \
  --password "@keychain:AC_PASSWORD"

# 装订公证票据
xcrun stapler staple mHost.dmg
```

### 10.4 Helper Tool 部署策略

#### 首次安装流程
```go
type HelperInstaller struct {
    appBundle    string
    helperPath   string
    serviceName  string
}

// 安装 Helper Tool
func (hi *HelperInstaller) InstallOnFirstRun() error {
    // 1. 检查是否已安装
    if hi.isHelperInstalled() {
        return nil
    }
    
    // 2. 请求管理员权限
    if err := hi.requestAdminPermission(); err != nil {
        return err
    }
    
    // 3. 复制 Helper Tool 到系统目录
    if err := hi.copyHelperToSystem(); err != nil {
        return err
    }
    
    // 4. 注册 launchd 服务
    if err := hi.registerService(); err != nil {
        return err
    }
    
    // 5. 启动服务
    return hi.startService()
}
```

#### 卸载流程
```go
// 卸载 Helper Tool
func (hi *HelperInstaller) Uninstall() error {
    // 1. 停止服务
    if err := hi.stopService(); err != nil {
        return err
    }
    
    // 2. 注销 launchd 服务
    if err := hi.unregisterService(); err != nil {
        return err
    }
    
    // 3. 删除 Helper Tool 文件
    return hi.removeHelperFiles()
}
```

### 10.5 分发方式

- **直接下载**: 提供 DMG 文件下载
- **Homebrew**: 支持通过 Homebrew 安装
- **Mac App Store**: 考虑上架 Mac App Store

## 11. 维护和更新

### 11.1 版本管理

- **语义化版本**: 使用语义化版本号
- **变更日志**: 维护详细的变更日志
- **向后兼容**: 确保配置文件的向后兼容性

### 11.2 自动更新

- **更新检查**: 定期检查新版本
- **增量更新**: 支持增量更新机制
- **回滚机制**: 提供更新失败时的回滚功能

这个技术设计文档为 mHost 项目提供了详细的技术实现指导，涵盖了系统架构、数据模型、核心组件、安全性、性能优化等各个方面，为后续的开发工作奠定了坚实的基础。