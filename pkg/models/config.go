package models

import "time"

// AppConfig 应用程序配置
type AppConfig struct {
	Window   WindowConfig   `json:"window"`   // 窗口配置
	Backup   BackupConfig   `json:"backup"`   // 备份配置
	Log      LogConfig      `json:"log"`      // 日志配置
	Security SecurityConfig `json:"security"` // 安全配置
	UI       UIConfig       `json:"ui"`       // UI配置
}

// WindowConfig 窗口配置
type WindowConfig struct {
	Width     int  `json:"width"`     // 窗口宽度
	Height    int  `json:"height"`    // 窗口高度
	X         int  `json:"x"`         // 窗口X坐标
	Y         int  `json:"y"`         // 窗口Y坐标
	Maximized bool `json:"maximized"` // 是否最大化
}

// BackupConfig 备份配置
type BackupConfig struct {
	Enabled       bool          `json:"enabled"`        // 是否启用自动备份
	Interval      time.Duration `json:"interval"`       // 备份间隔
	MaxBackups    int           `json:"max_backups"`    // 最大备份数量
	BackupPath    string        `json:"backup_path"`    // 备份路径
	Compression   bool          `json:"compression"`    // 是否压缩备份
	RetentionDays int           `json:"retention_days"` // 备份保留天数
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `json:"level"`       // 日志级别 (debug, info, warn, error)
	FilePath   string `json:"file_path"`   // 日志文件路径
	MaxSize    int    `json:"max_size"`    // 最大文件大小(MB)
	MaxBackups int    `json:"max_backups"` // 最大备份文件数
	MaxAge     int    `json:"max_age"`     // 最大保留天数
	Compress   bool   `json:"compress"`    // 是否压缩旧日志
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	RequireConfirmation bool     `json:"require_confirmation"` // 是否需要确认危险操作
	AllowedIPs          []string `json:"allowed_ips"`          // 允许的IP地址范围
	BlockedHosts        []string `json:"blocked_hosts"`        // 禁止的主机名
	AuditLog            bool     `json:"audit_log"`            // 是否启用审计日志
	BackupBeforeChange  bool     `json:"backup_before_change"` // 修改前是否自动备份
}

// UIConfig UI配置
type UIConfig struct {
	Theme            string `json:"theme"`              // 主题 (light, dark, auto)
	Language         string `json:"language"`           // 语言
	ShowLineNumbers  bool   `json:"show_line_numbers"`  // 是否显示行号
	FontSize         int    `json:"font_size"`          // 字体大小
	AutoSave         bool   `json:"auto_save"`          // 是否自动保存
	AutoSaveInterval int    `json:"auto_save_interval"` // 自动保存间隔(秒)
}

// DefaultAppConfig 返回默认的应用程序配置
func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		Window: WindowConfig{
			Width:     800,
			Height:    600,
			X:         -1, // -1表示居中
			Y:         -1, // -1表示居中
			Maximized: false,
		},
		Backup: BackupConfig{
			Enabled:       true,
			Interval:      24 * time.Hour, // 每天备份一次
			MaxBackups:    10,
			BackupPath:    "", // 空字符串表示使用默认路径
			Compression:   true,
			RetentionDays: 30,
		},
		Log: LogConfig{
			Level:      "info",
			FilePath:   "", // 空字符串表示使用默认路径
			MaxSize:    10, // 10MB
			MaxBackups: 5,
			MaxAge:     30, // 30天
			Compress:   true,
		},
		Security: SecurityConfig{
			RequireConfirmation: true,
			AllowedIPs:          []string{},
			BlockedHosts:        []string{},
			AuditLog:            true,
			BackupBeforeChange:  true,
		},
		UI: UIConfig{
			Theme:            "auto",
			Language:         "zh-CN",
			ShowLineNumbers:  true,
			FontSize:         12,
			AutoSave:         true,
			AutoSaveInterval: 30, // 30秒
		},
	}
}

// Validate 验证配置的有效性
func (c *AppConfig) Validate() error {
	if c.Window.Width <= 0 || c.Window.Height <= 0 {
		return ErrInvalidConfig
	}

	if c.Backup.MaxBackups < 0 || c.Backup.RetentionDays < 0 {
		return ErrInvalidConfig
	}

	if c.Log.MaxSize <= 0 || c.Log.MaxBackups < 0 || c.Log.MaxAge < 0 {
		return ErrInvalidConfig
	}

	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[c.Log.Level] {
		return ErrInvalidConfig
	}

	validThemes := map[string]bool{
		"light": true,
		"dark":  true,
		"auto":  true,
	}
	if !validThemes[c.UI.Theme] {
		return ErrInvalidConfig
	}

	if c.UI.FontSize <= 0 || c.UI.AutoSaveInterval <= 0 {
		return ErrInvalidConfig
	}

	return nil
}

// Clone 创建配置的深拷贝
func (c *AppConfig) Clone() *AppConfig {
	cloned := *c

	// 深拷贝切片
	cloned.Security.AllowedIPs = make([]string, len(c.Security.AllowedIPs))
	copy(cloned.Security.AllowedIPs, c.Security.AllowedIPs)

	cloned.Security.BlockedHosts = make([]string, len(c.Security.BlockedHosts))
	copy(cloned.Security.BlockedHosts, c.Security.BlockedHosts)

	return &cloned
}
