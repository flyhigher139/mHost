package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/flyhigher139/mhost/pkg/models"
)

// Manager 定义配置管理器接口
type Manager interface {
	// LoadConfig 加载配置
	LoadConfig() (*models.AppConfig, error)

	// SaveConfig 保存配置
	SaveConfig(config *models.AppConfig) error

	// GetConfig 获取当前配置
	GetConfig() *models.AppConfig

	// UpdateConfig 更新配置
	UpdateConfig(updater func(*models.AppConfig)) error

	// ResetToDefault 重置为默认配置
	ResetToDefault() error

	// GetConfigPath 获取配置文件路径
	GetConfigPath() string

	// ValidateConfig 验证配置有效性
	ValidateConfig(config *models.AppConfig) error

	// BackupConfig 备份当前配置
	BackupConfig() error

	// RestoreConfig 从备份恢复配置
	RestoreConfig(backupPath string) error

	// WatchConfig 监听配置文件变化
	WatchConfig(callback func(*models.AppConfig)) error

	// StopWatching 停止监听配置文件
	StopWatching()
}

// ManagerImpl 配置管理器实现
type ManagerImpl struct {
	configPath    string
	backupDir     string
	currentConfig *models.AppConfig
	mu            sync.RWMutex
	watching      bool
	stopChan      chan struct{}
}

// NewManager 创建新的配置管理器
func NewManager(configPath, backupDir string) Manager {
	if configPath == "" {
		configPath = getDefaultConfigPath()
	}

	if backupDir == "" {
		backupDir = getDefaultBackupDir()
	}

	return &ManagerImpl{
		configPath: configPath,
		backupDir:  backupDir,
		stopChan:   make(chan struct{}),
	}
}

// getDefaultConfigPath 获取默认配置文件路径
func getDefaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "./config.json"
	}
	return filepath.Join(homeDir, ".mhost", "config.json")
}

// getDefaultBackupDir 获取默认备份目录
func getDefaultBackupDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "./backups"
	}
	return filepath.Join(homeDir, ".mhost", "backups")
}

// LoadConfig 加载配置
func (m *ManagerImpl) LoadConfig() (*models.AppConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	config, err := m.loadConfigInternal()
	if err != nil {
		return nil, err
	}
	m.currentConfig = config
	return config, nil
}

// loadConfigInternal 内部加载配置方法，不获取锁
func (m *ManagerImpl) loadConfigInternal() (*models.AppConfig, error) {
	// 检查配置文件是否存在
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		// 配置文件不存在，创建默认配置
		defaultConfig := models.DefaultAppConfig()
		if err := m.saveConfigInternal(defaultConfig); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return defaultConfig, nil
	}

	// 读取配置文件
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析JSON
	var config models.AppConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 验证配置
	if err := m.validateConfigInternal(&config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

// SaveConfig 保存配置
func (m *ManagerImpl) SaveConfig(config *models.AppConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if config == nil {
		return models.ErrInvalidConfig
	}

	// 验证配置
	if err := m.validateConfigInternal(config); err != nil {
		return err
	}

	// 保存配置
	if err := m.saveConfigInternal(config); err != nil {
		return err
	}

	m.currentConfig = config
	return nil
}

// saveConfigInternal 内部保存配置方法（不加锁）
func (m *ManagerImpl) saveConfigInternal(config *models.AppConfig) error {
	// 确保配置目录存在
	configDir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 序列化配置
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// 写入临时文件
	tempPath := m.configPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp config file: %w", err)
	}

	// 原子性替换
	if err := os.Rename(tempPath, m.configPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to replace config file: %w", err)
	}

	return nil
}

// GetConfig 获取当前配置
func (m *ManagerImpl) GetConfig() *models.AppConfig {
	m.mu.RLock()
	if m.currentConfig != nil {
		config := m.currentConfig.Clone()
		m.mu.RUnlock()
		return config
	}
	m.mu.RUnlock()

	// 需要加载配置，先释放读锁再获取写锁
	m.mu.Lock()
	defer m.mu.Unlock()

	// 双重检查，防止在获取写锁期间其他goroutine已经加载了配置
	if m.currentConfig != nil {
		return m.currentConfig.Clone()
	}

	// 尝试加载配置
	config, err := m.loadConfigInternal()
	if err != nil {
		// 返回默认配置
		return models.DefaultAppConfig()
	}
	m.currentConfig = config
	return config.Clone()
}

// UpdateConfig 更新配置
func (m *ManagerImpl) UpdateConfig(updater func(*models.AppConfig)) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 获取当前配置的副本
	var config *models.AppConfig
	if m.currentConfig != nil {
		config = m.currentConfig.Clone()
	} else {
		config = models.DefaultAppConfig()
	}

	// 应用更新
	updater(config)

	// 验证更新后的配置
	if err := m.validateConfigInternal(config); err != nil {
		return err
	}

	// 保存配置
	if err := m.saveConfigInternal(config); err != nil {
		return err
	}

	m.currentConfig = config
	return nil
}

// ResetToDefault 重置为默认配置
func (m *ManagerImpl) ResetToDefault() error {
	defaultConfig := models.DefaultAppConfig()
	return m.SaveConfig(defaultConfig)
}

// GetConfigPath 获取配置文件路径
func (m *ManagerImpl) GetConfigPath() string {
	return m.configPath
}

// ValidateConfig 验证配置有效性
func (m *ManagerImpl) ValidateConfig(config *models.AppConfig) error {
	return m.validateConfigInternal(config)
}

// validateConfigInternal 内部验证配置方法
func (m *ManagerImpl) validateConfigInternal(config *models.AppConfig) error {
	if config == nil {
		return models.ErrInvalidConfig
	}

	// 使用模型的验证方法
	return config.Validate()
}

// BackupConfig 备份当前配置
func (m *ManagerImpl) BackupConfig() error {
	m.mu.RLock()
	configPath := m.configPath
	backupDir := m.backupDir
	m.mu.RUnlock()

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return models.ErrConfigNotFound
	}

	// 确保备份目录存在
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// 生成备份文件名
	backupFileName := fmt.Sprintf("config_backup_%d.json", getCurrentTimestamp())
	backupPath := filepath.Join(backupDir, backupFileName)

	// 读取当前配置
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// 写入备份文件
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	return nil
}

// RestoreConfig 从备份恢复配置
func (m *ManagerImpl) RestoreConfig(backupPath string) error {
	// 检查备份文件是否存在
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return models.ErrFileNotFound
	}

	// 读取备份文件
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	// 解析配置
	var config models.AppConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse backup config: %w", err)
	}

	// 验证配置
	if err := m.ValidateConfig(&config); err != nil {
		return fmt.Errorf("invalid backup config: %w", err)
	}

	// 保存配置
	return m.SaveConfig(&config)
}

// WatchConfig 监听配置文件变化
func (m *ManagerImpl) WatchConfig(callback func(*models.AppConfig)) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.watching {
		return fmt.Errorf("already watching config file")
	}

	m.watching = true
	m.stopChan = make(chan struct{})

	// 启动监听goroutine
	go m.watchConfigFile(callback)

	return nil
}

// StopWatching 停止监听配置文件
func (m *ManagerImpl) StopWatching() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.watching {
		close(m.stopChan)
		m.watching = false
	}
}

// watchConfigFile 监听配置文件变化的内部方法
func (m *ManagerImpl) watchConfigFile(callback func(*models.AppConfig)) {
	// 简单的轮询实现（在实际项目中可以使用fsnotify等库）
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	var lastModTime time.Time
	if stat, err := os.Stat(m.configPath); err == nil {
		lastModTime = stat.ModTime()
	}

	for {
		select {
		case <-m.stopChan:
			return
		case <-ticker.C:
			if stat, err := os.Stat(m.configPath); err == nil {
				if stat.ModTime().After(lastModTime) {
					lastModTime = stat.ModTime()

					// 重新加载配置
					if config, err := m.LoadConfig(); err == nil {
						callback(config)
					}
				}
			}
		}
	}
}

// getCurrentTimestamp 获取当前时间戳
func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}
