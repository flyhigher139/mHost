package host

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/flyhigher139/mhost/pkg/models"
)

// HostManagerTestSuite Host Manager测试套件
type HostManagerTestSuite struct {
	suite.Suite
	manager       Manager
	tempDir       string
	hostsPath     string
	backupDir     string
	originalHosts string
}

// SetupSuite 设置测试套件
func (suite *HostManagerTestSuite) SetupSuite() {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "mhost_test_*")
	require.NoError(suite.T(), err)
	suite.tempDir = tempDir

	// 设置测试文件路径
	suite.hostsPath = filepath.Join(tempDir, "hosts")
	suite.backupDir = filepath.Join(tempDir, "backups")

	// 创建测试hosts文件
	suite.originalHosts = `127.0.0.1	localhost
::1		localhost
# Test comment
192.168.1.100	test.local	# Test entry`
	err = os.WriteFile(suite.hostsPath, []byte(suite.originalHosts), 0644)
	require.NoError(suite.T(), err)

	// 创建manager实例
	suite.manager = NewManager(suite.hostsPath, suite.backupDir)
}

// TearDownSuite 清理测试套件
func (suite *HostManagerTestSuite) TearDownSuite() {
	if suite.tempDir != "" {
		os.RemoveAll(suite.tempDir)
	}
}

// SetupTest 每个测试前的设置
func (suite *HostManagerTestSuite) SetupTest() {
	// 恢复原始hosts文件内容
	err := os.WriteFile(suite.hostsPath, []byte(suite.originalHosts), 0644)
	require.NoError(suite.T(), err)
}

// TestReadHostsFile 测试读取hosts文件
func (suite *HostManagerTestSuite) TestReadHostsFile() {
	lines, err := suite.manager.ReadHostsFile()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), lines, 4)
	assert.Equal(suite.T(), "127.0.0.1\tlocalhost", lines[0])
	assert.Equal(suite.T(), "::1\t\tlocalhost", lines[1])
	assert.Equal(suite.T(), "# Test comment", lines[2])
	assert.Equal(suite.T(), "192.168.1.100\ttest.local\t# Test entry", lines[3])
}

// TestWriteHostsFile 测试写入hosts文件
func (suite *HostManagerTestSuite) TestWriteHostsFile() {
	newLines := []string{
		"127.0.0.1\tlocalhost",
		"192.168.1.200\tnew.local",
	}

	err := suite.manager.WriteHostsFile(newLines)
	assert.NoError(suite.T(), err)

	// 验证文件内容
	lines, err := suite.manager.ReadHostsFile()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), newLines, lines)
}

// TestApplyProfile 测试应用Profile
func (suite *HostManagerTestSuite) TestApplyProfile() {
	// 创建测试Profile
	profile := &models.Profile{
		ID:          "test-profile",
		Name:        "Test Profile",
		Description: "Test description",
		Entries: []*models.HostEntry{
			{
				ID:       "entry1",
				IP:       "192.168.1.10",
				Hostname: "app.local",
				Comment:  "App server",
				Enabled:  true,
			},
			{
				ID:       "entry2",
				IP:       "192.168.1.20",
				Hostname: "db.local",
				Comment:  "Database server",
				Enabled:  false, // 禁用的条目不应该被添加
			},
		},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := suite.manager.ApplyProfile(profile)
	assert.NoError(suite.T(), err)

	// 验证hosts文件内容
	lines, err := suite.manager.ReadHostsFile()
	assert.NoError(suite.T(), err)

	// 检查是否包含mHost管理的section
	found := false
	for _, line := range lines {
		if strings.Contains(line, "# mHost managed section START") {
			found = true
			break
		}
	}
	assert.True(suite.T(), found, "应该包含mHost管理的section")

	// 检查是否包含启用的条目
	foundEntry := false
	for _, line := range lines {
		if strings.Contains(line, "192.168.1.10\tapp.local") {
			foundEntry = true
			break
		}
	}
	assert.True(suite.T(), foundEntry, "应该包含启用的host条目")

	// 检查是否不包含禁用的条目
	foundDisabled := false
	for _, line := range lines {
		if strings.Contains(line, "192.168.1.20\tdb.local") {
			foundDisabled = true
			break
		}
	}
	assert.False(suite.T(), foundDisabled, "不应该包含禁用的host条目")
}

// TestApplyEmptyProfile 测试应用空Profile
func (suite *HostManagerTestSuite) TestApplyEmptyProfile() {
	profile := &models.Profile{
		ID:          "empty-profile",
		Name:        "Empty Profile",
		Description: "Empty test profile",
		Entries:     []*models.HostEntry{},
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := suite.manager.ApplyProfile(profile)
	assert.NoError(suite.T(), err)

	// 验证hosts文件内容（应该只包含原始内容，没有mHost section）
	lines, err := suite.manager.ReadHostsFile()
	assert.NoError(suite.T(), err)

	// 检查不应该包含mHost管理的section
	found := false
	for _, line := range lines {
		if strings.Contains(line, "# mHost managed section") {
			found = true
			break
		}
	}
	assert.False(suite.T(), found, "空Profile不应该添加mHost管理的section")
}

// TestApplyNilProfile 测试应用nil Profile
func (suite *HostManagerTestSuite) TestApplyNilProfile() {
	err := suite.manager.ApplyProfile(nil)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), models.ErrInvalidProfile, err)
}

// TestBackupHostsFile 测试备份hosts文件
func (suite *HostManagerTestSuite) TestBackupHostsFile() {
	backup, err := suite.manager.BackupHostsFile()
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), backup)
	assert.NotEmpty(suite.T(), backup.ID)
	assert.Equal(suite.T(), models.BackupTypeManual, backup.Type)
	assert.Equal(suite.T(), suite.hostsPath, backup.OriginalPath)
	assert.Greater(suite.T(), backup.Size, int64(0))
	assert.True(suite.T(), backup.CreatedAt.After(time.Now().Add(-time.Minute)))

	// 验证备份文件存在
	_, err = os.Stat(backup.FilePath)
	assert.NoError(suite.T(), err)

	// 验证备份文件内容
	backupContent, err := os.ReadFile(backup.FilePath)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), suite.originalHosts, string(backupContent))
}

// TestRestoreFromBackup 测试从备份恢复
func (suite *HostManagerTestSuite) TestRestoreFromBackup() {
	// 先创建备份
	backup, err := suite.manager.BackupHostsFile()
	require.NoError(suite.T(), err)

	// 修改hosts文件
	newContent := "127.0.0.1\tmodified.local"
	err = os.WriteFile(suite.hostsPath, []byte(newContent), 0644)
	require.NoError(suite.T(), err)

	// 从备份恢复
	err = suite.manager.RestoreFromBackup(backup)
	assert.NoError(suite.T(), err)

	// 验证文件内容已恢复
	restoredContent, err := os.ReadFile(suite.hostsPath)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), suite.originalHosts, string(restoredContent))
}

// TestRestoreFromInvalidBackup 测试从无效备份恢复
func (suite *HostManagerTestSuite) TestRestoreFromInvalidBackup() {
	// 测试nil备份
	err := suite.manager.RestoreFromBackup(nil)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), models.ErrInvalidBackup, err)

	// 测试不存在的备份文件
	backup := &models.Backup{
		ID:           "invalid-backup",
		Type:         models.BackupTypeManual,
		FilePath:     "/nonexistent/path",
		OriginalPath: suite.hostsPath,
		Size:         100,
		CreatedAt:    time.Now(),
	}

	err = suite.manager.RestoreFromBackup(backup)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), models.ErrBackupNotFound, err)
}

// TestGetHostsFilePath 测试获取hosts文件路径
func (suite *HostManagerTestSuite) TestGetHostsFilePath() {
	path := suite.manager.GetHostsFilePath()
	assert.Equal(suite.T(), suite.hostsPath, path)
}

// TestValidateHostsFile 测试验证hosts文件
func (suite *HostManagerTestSuite) TestValidateHostsFile() {
	// 测试有效的hosts文件
	err := suite.manager.ValidateHostsFile()
	assert.NoError(suite.T(), err)

	// 测试无效的hosts文件
	invalidContent := `invalid.ip.address	test.local
127.0.0.1	invalid..hostname`
	err = os.WriteFile(suite.hostsPath, []byte(invalidContent), 0644)
	require.NoError(suite.T(), err)

	err = suite.manager.ValidateHostsFile()
	assert.Error(suite.T(), err)
}

// TestParseHostsFile 测试解析hosts文件
func (suite *HostManagerTestSuite) TestParseHostsFile() {
	entries, err := suite.manager.ParseHostsFile()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), entries, 3) // localhost (IPv4), localhost (IPv6), test.local

	// 验证第一个条目
	assert.Equal(suite.T(), "127.0.0.1", entries[0].IP)
	assert.Equal(suite.T(), "localhost", entries[0].Hostname)
	assert.Equal(suite.T(), "", entries[0].Comment)
	assert.True(suite.T(), entries[0].Enabled)

	// 验证最后一个条目（带注释）
	lastEntry := entries[len(entries)-1]
	assert.Equal(suite.T(), "192.168.1.100", lastEntry.IP)
	assert.Equal(suite.T(), "test.local", lastEntry.Hostname)
	assert.Equal(suite.T(), "Test entry", lastEntry.Comment)
	assert.True(suite.T(), lastEntry.Enabled)
}

// TestGetManagedSection 测试获取管理的section
func (suite *HostManagerTestSuite) TestGetManagedSection() {
	// 初始状态应该没有管理的section
	managedLines, err := suite.manager.GetManagedSection()
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), managedLines)

	// 应用Profile后应该有管理的section
	profile := &models.Profile{
		ID:   "test-profile",
		Name: "Test Profile",
		Entries: []*models.HostEntry{
			{
				ID:       "entry1",
				IP:       "192.168.1.10",
				Hostname: "app.local",
				Enabled:  true,
			},
		},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = suite.manager.ApplyProfile(profile)
	require.NoError(suite.T(), err)

	managedLines, err = suite.manager.GetManagedSection()
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), managedLines)

	// 检查是否包含预期的条目
	found := false
	for _, line := range managedLines {
		if strings.Contains(line, "192.168.1.10\tapp.local") {
			found = true
			break
		}
	}
	assert.True(suite.T(), found)
}

// TestUpdateManagedSection 测试更新管理的section
func (suite *HostManagerTestSuite) TestUpdateManagedSection() {
	entries := []*models.HostEntry{
		{
			ID:       "entry1",
			IP:       "192.168.1.10",
			Hostname: "app.local",
			Comment:  "App server",
			Enabled:  true,
		},
		{
			ID:       "entry2",
			IP:       "192.168.1.20",
			Hostname: "db.local",
			Comment:  "Database server",
			Enabled:  true,
		},
	}

	err := suite.manager.UpdateManagedSection(entries)
	assert.NoError(suite.T(), err)

	// 验证更新后的内容
	managedLines, err := suite.manager.GetManagedSection()
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), managedLines)

	// 检查是否包含两个条目
	foundApp := false
	foundDB := false
	for _, line := range managedLines {
		if strings.Contains(line, "192.168.1.10\tapp.local") {
			foundApp = true
		}
		if strings.Contains(line, "192.168.1.20\tdb.local") {
			foundDB = true
		}
	}
	assert.True(suite.T(), foundApp)
	assert.True(suite.T(), foundDB)
}

// TestUpdateManagedSectionEmpty 测试更新空的管理section
func (suite *HostManagerTestSuite) TestUpdateManagedSectionEmpty() {
	// 先添加一些条目
	entries := []*models.HostEntry{
		{
			ID:       "entry1",
			IP:       "192.168.1.10",
			Hostname: "app.local",
			Enabled:  true,
		},
	}

	err := suite.manager.UpdateManagedSection(entries)
	require.NoError(suite.T(), err)

	// 然后清空
	err = suite.manager.UpdateManagedSection([]*models.HostEntry{})
	assert.NoError(suite.T(), err)

	// 验证管理的section已被移除
	managedLines, err := suite.manager.GetManagedSection()
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), managedLines)

	// 验证原始内容仍然存在
	lines, err := suite.manager.ReadHostsFile()
	assert.NoError(suite.T(), err)
	found := false
	for _, line := range lines {
		if strings.Contains(line, "127.0.0.1\tlocalhost") {
			found = true
			break
		}
	}
	assert.True(suite.T(), found, "原始hosts内容应该保留")
}

// TestHostManagerSuite 运行Host Manager测试套件
func TestHostManagerSuite(t *testing.T) {
	suite.Run(t, new(HostManagerTestSuite))
}

// BenchmarkApplyProfile 性能测试：应用Profile
func BenchmarkApplyProfile(b *testing.B) {
	// 创建临时目录和文件
	tempDir, err := os.MkdirTemp("", "mhost_bench_*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	hostsPath := filepath.Join(tempDir, "hosts")
	backupDir := filepath.Join(tempDir, "backups")

	// 创建测试hosts文件
	originalHosts := `127.0.0.1	localhost
::1		localhost`
	err = os.WriteFile(hostsPath, []byte(originalHosts), 0644)
	if err != nil {
		b.Fatal(err)
	}

	manager := NewManager(hostsPath, backupDir)

	// 创建测试Profile
	profile := &models.Profile{
		ID:   "bench-profile",
		Name: "Benchmark Profile",
		Entries: []*models.HostEntry{
			{
				ID:       "entry1",
				IP:       "192.168.1.10",
				Hostname: "app.local",
				Enabled:  true,
			},
			{
				ID:       "entry2",
				IP:       "192.168.1.20",
				Hostname: "db.local",
				Enabled:  true,
			},
		},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := manager.ApplyProfile(profile)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseHostsFile 性能测试：解析hosts文件
func BenchmarkParseHostsFile(b *testing.B) {
	// 创建临时目录和文件
	tempDir, err := os.MkdirTemp("", "mhost_bench_*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	hostsPath := filepath.Join(tempDir, "hosts")
	backupDir := filepath.Join(tempDir, "backups")

	// 创建包含多个条目的hosts文件
	var hostsContent strings.Builder
	hostsContent.WriteString("127.0.0.1\tlocalhost\n")
	hostsContent.WriteString("::1\t\tlocalhost\n")
	for i := 0; i < 100; i++ {
		hostsContent.WriteString(fmt.Sprintf("192.168.1.%d\ttest%d.local\t# Test entry %d\n", i+1, i+1, i+1))
	}

	err = os.WriteFile(hostsPath, []byte(hostsContent.String()), 0644)
	if err != nil {
		b.Fatal(err)
	}

	manager := NewManager(hostsPath, backupDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.ParseHostsFile()
		if err != nil {
			b.Fatal(err)
		}
	}
}
