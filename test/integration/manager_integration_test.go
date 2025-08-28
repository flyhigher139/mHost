package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/flyhigher139/mhost/internal/config"
	"github.com/flyhigher139/mhost/internal/host"
	"github.com/flyhigher139/mhost/internal/profile"
	"github.com/flyhigher139/mhost/pkg/models"
)

// ManagerIntegrationTestSuite 管理器集成测试套件
type ManagerIntegrationTestSuite struct {
	suite.Suite
	testDir        string
	configManager  config.Manager
	profileManager profile.Manager
	hostManager    host.Manager
}

// SetupSuite 设置测试套件
func (suite *ManagerIntegrationTestSuite) SetupSuite() {
	// 创建临时测试目录
	testDir, err := os.MkdirTemp("", "mhost_integration_test")
	require.NoError(suite.T(), err)
	suite.testDir = testDir
}

// SetupTest 每个测试前的设置
func (suite *ManagerIntegrationTestSuite) SetupTest() {
	// 为每个测试创建独立的子目录
	testSubDir, err := os.MkdirTemp(suite.testDir, "test_*")
	require.NoError(suite.T(), err)

	// 初始化管理器
	suite.configManager = config.NewManager(filepath.Join(testSubDir, "config"), "test_config.json")
	profileManager, err := profile.NewManager(filepath.Join(testSubDir, "profiles"))
	require.NoError(suite.T(), err)
	suite.profileManager = profileManager
	suite.hostManager = host.NewManager(filepath.Join(testSubDir, "hosts"), filepath.Join(testSubDir, "backup"))
}

// TearDownSuite 清理测试套件
func (suite *ManagerIntegrationTestSuite) TearDownSuite() {
	if suite.testDir != "" {
		os.RemoveAll(suite.testDir)
	}
}

// TestProfileCreationAndRetrieval 测试Profile创建和获取
func (suite *ManagerIntegrationTestSuite) TestProfileCreationAndRetrieval() {
	// 创建测试Profile
	profile, err := suite.profileManager.CreateProfile("测试Profile", "集成测试用Profile")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), profile)
	assert.Equal(suite.T(), "测试Profile", profile.Name)
	assert.Equal(suite.T(), "集成测试用Profile", profile.Description)

	// 获取Profile
	retrievedProfile, err := suite.profileManager.GetProfile(profile.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), profile.ID, retrievedProfile.ID)
	assert.Equal(suite.T(), profile.Name, retrievedProfile.Name)

	// 列出所有Profile
	profiles, err := suite.profileManager.ListProfiles()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), profiles, 1)
	assert.Equal(suite.T(), profile.ID, profiles[0].ID)
}

// TestProfileActivation 测试Profile激活
func (suite *ManagerIntegrationTestSuite) TestProfileActivation() {
	// 创建测试Profile
	profile, err := suite.profileManager.CreateProfile("激活测试Profile", "用于测试激活功能")
	assert.NoError(suite.T(), err)

	// 激活Profile
	err = suite.profileManager.ActivateProfile(profile.ID)
	assert.NoError(suite.T(), err)

	// 验证活动Profile
	activeProfile, err := suite.profileManager.GetActiveProfile()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), profile.ID, activeProfile.ID)
}

// TestConfigManagement 测试配置管理
func (suite *ManagerIntegrationTestSuite) TestConfigManagement() {
	// 创建测试配置
	config := models.DefaultAppConfig()
	config.Backup.BackupPath = filepath.Join(suite.testDir, "backup")
	config.Window.Width = 800
	config.Window.Height = 600

	// 保存配置
	err := suite.configManager.SaveConfig(config)
	assert.NoError(suite.T(), err)

	// 加载配置
	loadedConfig, err := suite.configManager.LoadConfig()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), config.UI.Language, loadedConfig.UI.Language)
	assert.Equal(suite.T(), config.Backup.BackupPath, loadedConfig.Backup.BackupPath)
	assert.Equal(suite.T(), config.Window.Width, loadedConfig.Window.Width)
}

// TestHostsFileOperations 测试Hosts文件操作
func (suite *ManagerIntegrationTestSuite) TestHostsFileOperations() {
	// 创建测试Hosts内容
	testContent := []string{
		"127.0.0.1\tlocalhost",
		"192.168.1.100\ttest.example.com",
	}

	// 写入Hosts文件
	err := suite.hostManager.WriteHostsFile(testContent)
	assert.NoError(suite.T(), err)

	// 读取Hosts文件
	readContent, err := suite.hostManager.ReadHostsFile()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), readContent, 2)

	// 验证内容
	contentStr := strings.Join(readContent, "\n")
	assert.Contains(suite.T(), contentStr, "localhost")
	assert.Contains(suite.T(), contentStr, "test.example.com")
}

// TestBackupOperations 测试备份操作
func (suite *ManagerIntegrationTestSuite) TestBackupOperations() {
	// 创建测试Hosts内容
	testContent := []string{
		"127.0.0.1\tlocalhost",
		"192.168.1.100\tbackup.test.com",
	}

	// 写入Hosts文件
	err := suite.hostManager.WriteHostsFile(testContent)
	assert.NoError(suite.T(), err)

	// 创建备份
	backup, err := suite.hostManager.BackupHostsFile()
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), backup)
	assert.NotEmpty(suite.T(), backup.ID)
	assert.NotEmpty(suite.T(), backup.FilePath)

	// 验证备份文件存在
	_, err = os.Stat(backup.FilePath)
	assert.NoError(suite.T(), err)
}

// TestProfileWithHostEntries 测试带有Host条目的Profile
func (suite *ManagerIntegrationTestSuite) TestProfileWithHostEntries() {
	// 创建Profile
	profile, err := suite.profileManager.CreateProfile("Host条目测试", "包含Host条目的Profile")
	assert.NoError(suite.T(), err)

	// 添加Host条目
	profile.Entries = []*models.HostEntry{
		{
			ID:       "entry-1",
			IP:       "192.168.1.100",
			Hostname: "test.local",
			Comment:  "测试条目",
			Enabled:  true,
		},
		{
			ID:       "entry-2",
			IP:       "192.168.1.101",
			Hostname: "api.test.local",
			Comment:  "API测试条目",
			Enabled:  false,
		},
	}

	// 更新Profile
	err = suite.profileManager.UpdateProfile(profile)
	assert.NoError(suite.T(), err)

	// 验证Profile已更新
	updatedProfile, err := suite.profileManager.GetProfile(profile.ID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), updatedProfile.Entries, 2)
	assert.Equal(suite.T(), "test.local", updatedProfile.Entries[0].Hostname)
	assert.Equal(suite.T(), "api.test.local", updatedProfile.Entries[1].Hostname)
	assert.True(suite.T(), updatedProfile.Entries[0].Enabled)
	assert.False(suite.T(), updatedProfile.Entries[1].Enabled)
}

// TestFullWorkflow 测试完整工作流程
func (suite *ManagerIntegrationTestSuite) TestFullWorkflow() {
	// 1. 创建配置
	config := models.DefaultAppConfig()
	config.Backup.BackupPath = filepath.Join(suite.testDir, "workflow_backup")
	err := suite.configManager.SaveConfig(config)
	assert.NoError(suite.T(), err)

	// 2. 创建开发环境Profile
	devProfile, err := suite.profileManager.CreateProfile("开发环境", "开发环境配置")
	assert.NoError(suite.T(), err)

	devProfile.Entries = []*models.HostEntry{
		{ID: "dev-1", IP: "192.168.1.100", Hostname: "dev-api.local", Enabled: true},
		{ID: "dev-2", IP: "192.168.1.101", Hostname: "dev-web.local", Enabled: true},
	}
	err = suite.profileManager.UpdateProfile(devProfile)
	assert.NoError(suite.T(), err)

	// 3. 创建生产环境Profile
	prodProfile, err := suite.profileManager.CreateProfile("生产环境", "生产环境配置")
	assert.NoError(suite.T(), err)

	prodProfile.Entries = []*models.HostEntry{
		{ID: "prod-1", IP: "10.0.1.100", Hostname: "api.example.com", Enabled: true},
		{ID: "prod-2", IP: "10.0.1.101", Hostname: "web.example.com", Enabled: true},
	}
	err = suite.profileManager.UpdateProfile(prodProfile)
	assert.NoError(suite.T(), err)

	// 4. 验证Profile列表
	profiles, err := suite.profileManager.ListProfiles()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), profiles, 2)

	// 5. 激活开发环境
	err = suite.profileManager.ActivateProfile(devProfile.ID)
	assert.NoError(suite.T(), err)

	activeProfile, err := suite.profileManager.GetActiveProfile()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), devProfile.ID, activeProfile.ID)

	// 6. 创建初始hosts文件并应用Profile
	initialHosts := []string{"127.0.0.1\tlocalhost"}
	err = suite.hostManager.WriteHostsFile(initialHosts)
	assert.NoError(suite.T(), err)

	err = suite.hostManager.ApplyProfile(activeProfile)
	assert.NoError(suite.T(), err)

	// 7. 验证Hosts文件内容
	hostsContent, err := suite.hostManager.ReadHostsFile()
	assert.NoError(suite.T(), err)
	contentStr := strings.Join(hostsContent, "\n")
	assert.Contains(suite.T(), contentStr, "dev-api.local")
	assert.Contains(suite.T(), contentStr, "dev-web.local")

	// 8. 创建备份
	backup, err := suite.hostManager.BackupHostsFile()
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), backup)

	// 9. 切换到生产环境
	err = suite.profileManager.ActivateProfile(prodProfile.ID)
	assert.NoError(suite.T(), err)

	activeProfile, err = suite.profileManager.GetActiveProfile()
	assert.NoError(suite.T(), err)
	err = suite.hostManager.ApplyProfile(activeProfile)
	assert.NoError(suite.T(), err)

	// 10. 验证生产环境配置已应用
	hostsContent, err = suite.hostManager.ReadHostsFile()
	assert.NoError(suite.T(), err)
	contentStr = strings.Join(hostsContent, "\n")
	assert.Contains(suite.T(), contentStr, "api.example.com")
	assert.Contains(suite.T(), contentStr, "web.example.com")
	assert.NotContains(suite.T(), contentStr, "dev-api.local")
}

// TestManagerIntegrationSuite 运行集成测试套件
func TestManagerIntegrationSuite(t *testing.T) {
	suite.Run(t, new(ManagerIntegrationTestSuite))
}
