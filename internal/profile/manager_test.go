package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/flyhigher139/mhost/pkg/models"
)

// ProfileManagerTestSuite Profile Manager测试套件
type ProfileManagerTestSuite struct {
	suite.Suite
	manager *ManagerImpl
	tempDir string
}

// SetupTest 设置测试环境
func (suite *ProfileManagerTestSuite) SetupTest() {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "mhost_test_*")
	require.NoError(suite.T(), err)
	suite.tempDir = tempDir

	// 创建管理器实例
	manager, err := NewManager(tempDir)
	require.NoError(suite.T(), err)
	suite.manager = manager
}

// TearDownTest 清理测试环境
func (suite *ProfileManagerTestSuite) TearDownTest() {
	if suite.tempDir != "" {
		os.RemoveAll(suite.tempDir)
	}
}

// TestCreateProfile 测试创建Profile
func (suite *ProfileManagerTestSuite) TestCreateProfile() {
	// 创建第一个Profile
	profile1, err := suite.manager.CreateProfile("Test Profile 1", "Test Description 1")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), profile1)
	assert.Equal(suite.T(), "Test Profile 1", profile1.Name)
	assert.Equal(suite.T(), "Test Description 1", profile1.Description)
	assert.True(suite.T(), profile1.IsActive) // 第一个Profile应该自动激活
	assert.NotEmpty(suite.T(), profile1.ID)

	// 创建第二个Profile
	profile2, err := suite.manager.CreateProfile("Test Profile 2", "Test Description 2")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), profile2)
	assert.False(suite.T(), profile2.IsActive) // 第二个Profile不应该自动激活

	// 尝试创建重名Profile
	_, err = suite.manager.CreateProfile("Test Profile 1", "Duplicate")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), models.ErrProfileExists, err)
}

// TestListProfiles 测试获取Profile列表
func (suite *ProfileManagerTestSuite) TestListProfiles() {
	// 初始状态应该为空
	profiles, err := suite.manager.ListProfiles()
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), profiles)

	// 创建几个Profile
	_, err = suite.manager.CreateProfile("Profile 1", "Description 1")
	assert.NoError(suite.T(), err)

	time.Sleep(time.Millisecond) // 确保时间戳不同
	_, err = suite.manager.CreateProfile("Profile 2", "Description 2")
	assert.NoError(suite.T(), err)

	// 获取列表
	profiles, err = suite.manager.ListProfiles()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), profiles, 2)

	// 验证排序（按更新时间倒序）
	assert.True(suite.T(), profiles[0].UpdatedAt.After(profiles[1].UpdatedAt) ||
		profiles[0].UpdatedAt.Equal(profiles[1].UpdatedAt))
}

// TestGetProfile 测试获取单个Profile
func (suite *ProfileManagerTestSuite) TestGetProfile() {
	// 获取不存在的Profile
	_, err := suite.manager.GetProfile("nonexistent")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), models.ErrProfileNotFound, err)

	// 创建Profile
	created, err := suite.manager.CreateProfile("Test Profile", "Test Description")
	assert.NoError(suite.T(), err)

	// 获取Profile
	retrieved, err := suite.manager.GetProfile(created.ID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), retrieved)
	assert.Equal(suite.T(), created.ID, retrieved.ID)
	assert.Equal(suite.T(), created.Name, retrieved.Name)
	assert.Equal(suite.T(), created.Description, retrieved.Description)
}

// TestUpdateProfile 测试更新Profile
func (suite *ProfileManagerTestSuite) TestUpdateProfile() {
	// 创建Profile
	profile, err := suite.manager.CreateProfile("Original Name", "Original Description")
	assert.NoError(suite.T(), err)

	originalTime := profile.UpdatedAt
	time.Sleep(time.Millisecond)

	// 更新Profile
	profile.Name = "Updated Name"
	profile.Description = "Updated Description"
	err = suite.manager.UpdateProfile(profile)
	assert.NoError(suite.T(), err)

	// 验证更新
	updated, err := suite.manager.GetProfile(profile.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Name", updated.Name)
	assert.Equal(suite.T(), "Updated Description", updated.Description)
	assert.True(suite.T(), updated.UpdatedAt.After(originalTime))

	// 尝试更新不存在的Profile
	nonexistent := &models.Profile{ID: "nonexistent"}
	err = suite.manager.UpdateProfile(nonexistent)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), models.ErrProfileNotFound, err)
}

// TestDeleteProfile 测试删除Profile
func (suite *ProfileManagerTestSuite) TestDeleteProfile() {
	// 创建两个Profile
	profile1, err := suite.manager.CreateProfile("Profile 1", "Description 1")
	assert.NoError(suite.T(), err)

	profile2, err := suite.manager.CreateProfile("Profile 2", "Description 2")
	assert.NoError(suite.T(), err)

	// 尝试删除激活的Profile（应该失败）
	err = suite.manager.DeleteProfile(profile1.ID)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), models.ErrActiveProfile, err)

	// 删除非激活的Profile
	err = suite.manager.DeleteProfile(profile2.ID)
	assert.NoError(suite.T(), err)

	// 验证删除
	_, err = suite.manager.GetProfile(profile2.ID)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), models.ErrProfileNotFound, err)

	// 尝试删除不存在的Profile
	err = suite.manager.DeleteProfile("nonexistent")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), models.ErrProfileNotFound, err)
}

// TestActivateProfile 测试激活Profile
func (suite *ProfileManagerTestSuite) TestActivateProfile() {
	// 创建两个Profile
	profile1, err := suite.manager.CreateProfile("Profile 1", "Description 1")
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), profile1.IsActive)

	profile2, err := suite.manager.CreateProfile("Profile 2", "Description 2")
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), profile2.IsActive)

	// 激活第二个Profile
	err = suite.manager.ActivateProfile(profile2.ID)
	assert.NoError(suite.T(), err)

	// 验证激活状态
	active, err := suite.manager.GetActiveProfile()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), profile2.ID, active.ID)

	// 验证第一个Profile不再激活
	updated1, err := suite.manager.GetProfile(profile1.ID)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), updated1.IsActive)

	// 尝试激活不存在的Profile
	err = suite.manager.ActivateProfile("nonexistent")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), models.ErrProfileNotFound, err)
}

// TestGetActiveProfile 测试获取激活的Profile
func (suite *ProfileManagerTestSuite) TestGetActiveProfile() {
	// 没有Profile时
	_, err := suite.manager.GetActiveProfile()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), models.ErrProfileNotFound, err)

	// 创建Profile
	profile, err := suite.manager.CreateProfile("Test Profile", "Test Description")
	assert.NoError(suite.T(), err)

	// 获取激活的Profile
	active, err := suite.manager.GetActiveProfile()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), profile.ID, active.ID)
}

// TestCloneProfile 测试复制Profile
func (suite *ProfileManagerTestSuite) TestCloneProfile() {
	// 创建原始Profile
	original, err := suite.manager.CreateProfile("Original Profile", "Original Description")
	assert.NoError(suite.T(), err)

	// 添加一些host条目
	entry := models.NewHostEntry("127.0.0.1", "localhost", "Local host")
	original.AddEntry(entry)
	err = suite.manager.UpdateProfile(original)
	assert.NoError(suite.T(), err)

	// 复制Profile
	cloned, err := suite.manager.CloneProfile(original.ID, "Cloned Profile")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), cloned)
	assert.NotEqual(suite.T(), original.ID, cloned.ID)
	assert.Equal(suite.T(), "Cloned Profile", cloned.Name)
	assert.Equal(suite.T(), original.Description, cloned.Description)
	assert.Len(suite.T(), cloned.Entries, len(original.Entries))
	assert.False(suite.T(), cloned.IsActive)

	// 尝试用已存在的名称复制
	_, err = suite.manager.CloneProfile(original.ID, "Original Profile")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), models.ErrProfileExists, err)

	// 尝试复制不存在的Profile
	_, err = suite.manager.CloneProfile("nonexistent", "New Name")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), models.ErrProfileNotFound, err)
}

// TestSearchProfiles 测试搜索Profile
func (suite *ProfileManagerTestSuite) TestSearchProfiles() {
	// 创建测试Profile
	_, err := suite.manager.CreateProfile("Web Development", "Profiles for web development")
	assert.NoError(suite.T(), err)

	_, err = suite.manager.CreateProfile("Mobile Testing", "Profiles for mobile app testing")
	assert.NoError(suite.T(), err)

	_, err = suite.manager.CreateProfile("Production", "Production environment hosts")
	assert.NoError(suite.T(), err)

	// 搜索测试
	results, err := suite.manager.SearchProfiles("dev")
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), results, 1)
	assert.Equal(suite.T(), "Web Development", results[0].Name)

	results, err = suite.manager.SearchProfiles("test")
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), results, 1)
	assert.Equal(suite.T(), "Mobile Testing", results[0].Name)

	results, err = suite.manager.SearchProfiles("prod")
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), results, 1)
	assert.Equal(suite.T(), "Production", results[0].Name)

	// 搜索不存在的内容
	results, err = suite.manager.SearchProfiles("nonexistent")
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), results)
}

// TestPersistence 测试数据持久化
func (suite *ProfileManagerTestSuite) TestPersistence() {
	// 创建Profile
	profile, err := suite.manager.CreateProfile("Persistent Profile", "Test persistence")
	assert.NoError(suite.T(), err)

	// 创建新的管理器实例（模拟重启）
	newManager, err := NewManager(suite.tempDir)
	assert.NoError(suite.T(), err)

	// 验证数据是否持久化
	loaded, err := newManager.GetProfile(profile.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), profile.Name, loaded.Name)
	assert.Equal(suite.T(), profile.Description, loaded.Description)

	// 验证激活状态
	active, err := newManager.GetActiveProfile()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), profile.ID, active.ID)
}

// TestExportImportProfile 测试导出导入Profile
func (suite *ProfileManagerTestSuite) TestExportImportProfile() {
	// 创建Profile
	original, err := suite.manager.CreateProfile("Export Test", "Test export/import")
	assert.NoError(suite.T(), err)

	// 添加host条目
	entry := models.NewHostEntry("192.168.1.1", "router.local", "Local router")
	original.AddEntry(entry)
	err = suite.manager.UpdateProfile(original)
	assert.NoError(suite.T(), err)

	// 导出Profile
	exportPath := filepath.Join(suite.tempDir, "exported_profile.json")
	err = suite.manager.ExportProfile(original.ID, exportPath)
	assert.NoError(suite.T(), err)

	// 验证导出文件存在
	_, err = os.Stat(exportPath)
	assert.NoError(suite.T(), err)

	// 导入Profile
	imported, err := suite.manager.ImportProfile(exportPath)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), imported)
	assert.NotEqual(suite.T(), original.ID, imported.ID)     // ID应该不同
	assert.Contains(suite.T(), imported.Name, "Export Test") // 名称可能有后缀
	assert.Equal(suite.T(), original.Description, imported.Description)
	assert.Len(suite.T(), imported.Entries, len(original.Entries))
	assert.False(suite.T(), imported.IsActive)

	// 尝试导出不存在的Profile
	err = suite.manager.ExportProfile("nonexistent", exportPath)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), models.ErrProfileNotFound, err)
}

// 运行测试套件
func TestProfileManagerSuite(t *testing.T) {
	suite.Run(t, new(ProfileManagerTestSuite))
}

// 基准测试
func BenchmarkCreateProfile(b *testing.B) {
	tempDir, _ := os.MkdirTemp("", "mhost_bench_*")
	defer os.RemoveAll(tempDir)

	manager, _ := NewManager(tempDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.CreateProfile(fmt.Sprintf("Profile %d", i), "Benchmark test")
	}
}

func BenchmarkListProfiles(b *testing.B) {
	tempDir, _ := os.MkdirTemp("", "mhost_bench_*")
	defer os.RemoveAll(tempDir)

	manager, _ := NewManager(tempDir)

	// 创建一些Profile
	for i := 0; i < 100; i++ {
		manager.CreateProfile(fmt.Sprintf("Profile %d", i), "Benchmark test")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.ListProfiles()
	}
}
