package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/flyhigher139/mhost/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ConfigManagerTestSuite 配置管理器测试套件
type ConfigManagerTestSuite struct {
	suite.Suite
	manager    Manager
	tempDir    string
	configPath string
	backupDir  string
}

// SetupTest 设置测试环境
func (suite *ConfigManagerTestSuite) SetupTest() {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "config_test_*")
	require.NoError(suite.T(), err)
	suite.tempDir = tempDir

	// 设置配置文件路径
	suite.configPath = filepath.Join(tempDir, "config.json")
	suite.backupDir = filepath.Join(tempDir, "backups")

	// 创建配置管理器
	suite.manager = NewManager(suite.configPath, suite.backupDir)
}

// TearDownTest 清理测试环境
func (suite *ConfigManagerTestSuite) TearDownTest() {
	// 停止监听
	suite.manager.StopWatching()

	// 清理临时目录
	if suite.tempDir != "" {
		os.RemoveAll(suite.tempDir)
	}
}

// TestLoadConfigWithoutFile 测试加载不存在的配置文件
func (suite *ConfigManagerTestSuite) TestLoadConfigWithoutFile() {
	// 配置文件不存在时应该创建默认配置
	config, err := suite.manager.LoadConfig()
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), config)

	// 验证默认配置
	defaultConfig := models.DefaultAppConfig()
	assert.Equal(suite.T(), defaultConfig.Window.Width, config.Window.Width)
	assert.Equal(suite.T(), defaultConfig.Window.Height, config.Window.Height)

	// 配置文件应该被创建
	_, err = os.Stat(suite.configPath)
	assert.NoError(suite.T(), err)
}

// TestLoadConfigWithValidFile 测试加载有效的配置文件
func (suite *ConfigManagerTestSuite) TestLoadConfigWithValidFile() {
	// 创建测试配置
	testConfig := models.DefaultAppConfig()
	testConfig.Window.Width = 1200
	testConfig.Window.Height = 800
	testConfig.Backup.MaxBackups = 10

	// 保存配置
	err := suite.manager.SaveConfig(testConfig)
	require.NoError(suite.T(), err)

	// 重新加载配置
	loadedConfig, err := suite.manager.LoadConfig()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), testConfig.Window.Width, loadedConfig.Window.Width)
	assert.Equal(suite.T(), testConfig.Window.Height, loadedConfig.Window.Height)
	assert.Equal(suite.T(), testConfig.Backup.MaxBackups, loadedConfig.Backup.MaxBackups)
}

// TestLoadConfigWithInvalidFile 测试加载无效的配置文件
func (suite *ConfigManagerTestSuite) TestLoadConfigWithInvalidFile() {
	// 写入无效的JSON
	err := os.WriteFile(suite.configPath, []byte("invalid json"), 0644)
	require.NoError(suite.T(), err)

	// 加载配置应该失败
	_, err = suite.manager.LoadConfig()
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to parse config file")
}

// TestSaveConfig 测试保存配置
func (suite *ConfigManagerTestSuite) TestSaveConfig() {
	// 创建测试配置
	testConfig := models.DefaultAppConfig()
	testConfig.Window.Width = 1400
	testConfig.Window.Height = 900

	// 保存配置
	err := suite.manager.SaveConfig(testConfig)
	assert.NoError(suite.T(), err)

	// 验证文件内容
	data, err := os.ReadFile(suite.configPath)
	require.NoError(suite.T(), err)

	var savedConfig models.AppConfig
	err = json.Unmarshal(data, &savedConfig)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), testConfig.Window.Width, savedConfig.Window.Width)
	assert.Equal(suite.T(), testConfig.Window.Height, savedConfig.Window.Height)
}

// TestSaveConfigWithNil 测试保存nil配置
func (suite *ConfigManagerTestSuite) TestSaveConfigWithNil() {
	err := suite.manager.SaveConfig(nil)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), models.ErrInvalidConfig, err)
}

// TestGetConfig 测试获取配置
func (suite *ConfigManagerTestSuite) TestGetConfig() {
	// 第一次获取应该返回默认配置
	config := suite.manager.GetConfig()
	assert.NotNil(suite.T(), config)

	// 修改配置并保存
	config.Window.Width = 1600
	err := suite.manager.SaveConfig(config)
	require.NoError(suite.T(), err)

	// 再次获取应该返回修改后的配置
	updatedConfig := suite.manager.GetConfig()
	assert.Equal(suite.T(), 1600, updatedConfig.Window.Width)
}

// TestUpdateConfig 测试更新配置
func (suite *ConfigManagerTestSuite) TestUpdateConfig() {
	// 更新配置
	err := suite.manager.UpdateConfig(func(config *models.AppConfig) {
		config.Window.Width = 1800
		config.Window.Height = 1000
		config.Backup.Enabled = true
	})
	assert.NoError(suite.T(), err)

	// 验证更新结果
	config := suite.manager.GetConfig()
	assert.Equal(suite.T(), 1800, config.Window.Width)
	assert.Equal(suite.T(), 1000, config.Window.Height)
	assert.True(suite.T(), config.Backup.Enabled)
}

// TestResetToDefault 测试重置为默认配置
func (suite *ConfigManagerTestSuite) TestResetToDefault() {
	// 先修改配置
	err := suite.manager.UpdateConfig(func(config *models.AppConfig) {
		config.Window.Width = 2000
		config.Backup.MaxBackups = 50
	})
	require.NoError(suite.T(), err)

	// 重置为默认配置
	err = suite.manager.ResetToDefault()
	assert.NoError(suite.T(), err)

	// 验证配置已重置
	config := suite.manager.GetConfig()
	defaultConfig := models.DefaultAppConfig()
	assert.Equal(suite.T(), defaultConfig.Window.Width, config.Window.Width)
	assert.Equal(suite.T(), defaultConfig.Backup.MaxBackups, config.Backup.MaxBackups)
}

// TestGetConfigPath 测试获取配置文件路径
func (suite *ConfigManagerTestSuite) TestGetConfigPath() {
	path := suite.manager.GetConfigPath()
	assert.Equal(suite.T(), suite.configPath, path)
}

// TestValidateConfig 测试验证配置
func (suite *ConfigManagerTestSuite) TestValidateConfig() {
	// 测试有效配置
	validConfig := models.DefaultAppConfig()
	err := suite.manager.ValidateConfig(validConfig)
	assert.NoError(suite.T(), err)

	// 测试nil配置
	err = suite.manager.ValidateConfig(nil)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), models.ErrInvalidConfig, err)
}

// TestBackupConfig 测试备份配置
func (suite *ConfigManagerTestSuite) TestBackupConfig() {
	// 先创建配置文件
	testConfig := models.DefaultAppConfig()
	err := suite.manager.SaveConfig(testConfig)
	require.NoError(suite.T(), err)

	// 备份配置
	err = suite.manager.BackupConfig()
	assert.NoError(suite.T(), err)

	// 验证备份文件存在
	backupFiles, err := filepath.Glob(filepath.Join(suite.backupDir, "config_backup_*.json"))
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), backupFiles, 1)

	// 验证备份文件内容
	backupData, err := os.ReadFile(backupFiles[0])
	require.NoError(suite.T(), err)

	var backupConfig models.AppConfig
	err = json.Unmarshal(backupData, &backupConfig)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), testConfig.Window.Width, backupConfig.Window.Width)
}

// TestBackupConfigWithoutFile 测试备份不存在的配置文件
func (suite *ConfigManagerTestSuite) TestBackupConfigWithoutFile() {
	// 配置文件不存在时备份应该失败
	err := suite.manager.BackupConfig()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), models.ErrConfigNotFound, err)
}

// TestRestoreConfig 测试恢复配置
func (suite *ConfigManagerTestSuite) TestRestoreConfig() {
	// 创建原始配置
	originalConfig := models.DefaultAppConfig()
	originalConfig.Window.Width = 1500
	err := suite.manager.SaveConfig(originalConfig)
	require.NoError(suite.T(), err)

	// 备份配置
	err = suite.manager.BackupConfig()
	require.NoError(suite.T(), err)

	// 修改配置
	modifiedConfig := models.DefaultAppConfig()
	modifiedConfig.Window.Width = 2000
	err = suite.manager.SaveConfig(modifiedConfig)
	require.NoError(suite.T(), err)

	// 找到备份文件
	backupFiles, err := filepath.Glob(filepath.Join(suite.backupDir, "config_backup_*.json"))
	require.NoError(suite.T(), err)
	require.Len(suite.T(), backupFiles, 1)

	// 恢复配置
	err = suite.manager.RestoreConfig(backupFiles[0])
	assert.NoError(suite.T(), err)

	// 验证配置已恢复
	restoredConfig := suite.manager.GetConfig()
	assert.Equal(suite.T(), originalConfig.Window.Width, restoredConfig.Window.Width)
}

// TestRestoreConfigWithInvalidFile 测试从无效文件恢复配置
func (suite *ConfigManagerTestSuite) TestRestoreConfigWithInvalidFile() {
	// 测试不存在的文件
	err := suite.manager.RestoreConfig("/nonexistent/file.json")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), models.ErrFileNotFound, err)

	// 创建无效的备份文件
	invalidBackupPath := filepath.Join(suite.tempDir, "invalid_backup.json")
	err = os.WriteFile(invalidBackupPath, []byte("invalid json"), 0644)
	require.NoError(suite.T(), err)

	// 恢复应该失败
	err = suite.manager.RestoreConfig(invalidBackupPath)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to parse backup config")
}

// TestWatchConfig 测试监听配置文件变化
func (suite *ConfigManagerTestSuite) TestWatchConfig() {
	// 创建初始配置
	initialConfig := models.DefaultAppConfig()
	err := suite.manager.SaveConfig(initialConfig)
	require.NoError(suite.T(), err)

	// 设置回调函数
	callbackCalled := false
	callback := func(config *models.AppConfig) {
		callbackCalled = true
	}

	// 开始监听
	err = suite.manager.WatchConfig(callback)
	assert.NoError(suite.T(), err)

	// 立即停止监听以避免长时间运行
	suite.manager.StopWatching()

	// 验证监听功能可以正常启动和停止
	assert.False(suite.T(), callbackCalled) // 由于立即停止，回调不应被调用
}

// TestStopWatching 测试停止监听
func (suite *ConfigManagerTestSuite) TestStopWatching() {
	// 开始监听
	err := suite.manager.WatchConfig(func(*models.AppConfig) {})
	require.NoError(suite.T(), err)

	// 停止监听（应该不会出错）
	suite.manager.StopWatching()

	// 再次停止监听（应该不会出错）
	suite.manager.StopWatching()
}

// TestWatchConfigAlreadyWatching 测试重复监听
func (suite *ConfigManagerTestSuite) TestWatchConfigAlreadyWatching() {
	// 开始第一次监听
	err := suite.manager.WatchConfig(func(*models.AppConfig) {})
	require.NoError(suite.T(), err)

	// 尝试再次监听应该失败
	err = suite.manager.WatchConfig(func(*models.AppConfig) {})
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "already watching")

	// 清理
	suite.manager.StopWatching()
}

// TestSuite 运行测试套件
func TestConfigManagerSuite(t *testing.T) {
	suite.Run(t, new(ConfigManagerTestSuite))
}
