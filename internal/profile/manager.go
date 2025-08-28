package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/flyhigher139/mhost/pkg/models"
)

// Manager Profile管理器接口
type Manager interface {
	// 创建新的Profile
	CreateProfile(name, description string) (*models.Profile, error)

	// 获取Profile列表
	ListProfiles() ([]*models.ProfileSummary, error)

	// 根据ID获取Profile
	GetProfile(id string) (*models.Profile, error)

	// 更新Profile
	UpdateProfile(profile *models.Profile) error

	// 删除Profile
	DeleteProfile(id string) error

	// 激活Profile
	ActivateProfile(id string) error

	// 获取当前激活的Profile
	GetActiveProfile() (*models.Profile, error)

	// 导入Profile
	ImportProfile(filePath string) (*models.Profile, error)

	// 导出Profile
	ExportProfile(id, filePath string) error

	// 复制Profile
	CloneProfile(id, newName string) (*models.Profile, error)

	// 搜索Profile
	SearchProfiles(query string) ([]*models.ProfileSummary, error)
}

// ManagerImpl Profile管理器实现
type ManagerImpl struct {
	mu          sync.RWMutex
	profiles    map[string]*models.Profile
	activeID    string
	dataDir     string
	profileFile string
}

// NewManager 创建新的Profile管理器
func NewManager(dataDir string) (*ManagerImpl, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	manager := &ManagerImpl{
		profiles:    make(map[string]*models.Profile),
		dataDir:     dataDir,
		profileFile: filepath.Join(dataDir, "profiles.json"),
	}

	// 加载现有的Profile数据
	if err := manager.loadProfiles(); err != nil {
		return nil, fmt.Errorf("failed to load profiles: %w", err)
	}

	return manager, nil
}

// CreateProfile 创建新的Profile
func (m *ManagerImpl) CreateProfile(name, description string) (*models.Profile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查名称是否已存在
	for _, profile := range m.profiles {
		if profile.Name == name {
			return nil, models.ErrProfileExists
		}
	}

	profile := models.NewProfile(name, description)
	m.profiles[profile.ID] = profile

	// 如果这是第一个Profile，自动激活
	if len(m.profiles) == 1 {
		profile.IsActive = true
		m.activeID = profile.ID
	}

	if err := m.saveProfiles(); err != nil {
		return nil, fmt.Errorf("failed to save profile: %w", err)
	}

	return profile, nil
}

// ListProfiles 获取Profile列表
func (m *ManagerImpl) ListProfiles() ([]*models.ProfileSummary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	summaries := make([]*models.ProfileSummary, 0, len(m.profiles))
	for _, profile := range m.profiles {
		summary := profile.ToSummary()
		summaries = append(summaries, &summary)
	}

	// 按更新时间排序
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].UpdatedAt.After(summaries[j].UpdatedAt)
	})

	return summaries, nil
}

// GetProfile 根据ID获取Profile
func (m *ManagerImpl) GetProfile(id string) (*models.Profile, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	profile, exists := m.profiles[id]
	if !exists {
		return nil, models.ErrProfileNotFound
	}

	return profile.Clone(), nil
}

// UpdateProfile 更新Profile
func (m *ManagerImpl) UpdateProfile(profile *models.Profile) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.profiles[profile.ID]; !exists {
		return models.ErrProfileNotFound
	}

	// 验证Profile数据
	if err := profile.Validate(); err != nil {
		return err
	}

	// 检查名称冲突（排除自己）
	for id, existingProfile := range m.profiles {
		if id != profile.ID && existingProfile.Name == profile.Name {
			return models.ErrProfileExists
		}
	}

	profile.UpdateTimestamp()
	m.profiles[profile.ID] = profile

	return m.saveProfiles()
}

// DeleteProfile 删除Profile
func (m *ManagerImpl) DeleteProfile(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	profile, exists := m.profiles[id]
	if !exists {
		return models.ErrProfileNotFound
	}

	// 不能删除激活的Profile
	if profile.IsActive {
		return models.ErrActiveProfile
	}

	delete(m.profiles, id)
	return m.saveProfiles()
}

// ActivateProfile 激活Profile
func (m *ManagerImpl) ActivateProfile(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	profile, exists := m.profiles[id]
	if !exists {
		return models.ErrProfileNotFound
	}

	// 取消当前激活的Profile
	if m.activeID != "" {
		if currentActive, exists := m.profiles[m.activeID]; exists {
			currentActive.IsActive = false
		}
	}

	// 激活新的Profile
	profile.IsActive = true
	m.activeID = id

	return m.saveProfiles()
}

// GetActiveProfile 获取当前激活的Profile
func (m *ManagerImpl) GetActiveProfile() (*models.Profile, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.activeID == "" {
		return nil, models.ErrProfileNotFound
	}

	profile, exists := m.profiles[m.activeID]
	if !exists {
		return nil, models.ErrProfileNotFound
	}

	return profile.Clone(), nil
}

// ImportProfile 导入Profile
func (m *ManagerImpl) ImportProfile(filePath string) (*models.Profile, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var profile models.Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile: %w", err)
	}

	// 验证Profile数据
	if err := profile.Validate(); err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 生成新的ID和时间戳
	profile.ID = models.NewProfile("", "").ID // 临时生成ID
	profile.ID = fmt.Sprintf("%d-%x", time.Now().UnixNano(), []byte{1, 2, 3, 4})
	now := time.Now()
	profile.CreatedAt = now
	profile.UpdatedAt = now
	profile.IsActive = false

	// 检查名称冲突，如果存在则添加后缀
	originalName := profile.Name
	counter := 1
	for {
		nameExists := false
		for _, existingProfile := range m.profiles {
			if existingProfile.Name == profile.Name {
				nameExists = true
				break
			}
		}
		if !nameExists {
			break
		}
		profile.Name = fmt.Sprintf("%s (%d)", originalName, counter)
		counter++
	}

	m.profiles[profile.ID] = &profile

	if err := m.saveProfiles(); err != nil {
		return nil, fmt.Errorf("failed to save imported profile: %w", err)
	}

	return &profile, nil
}

// ExportProfile 导出Profile
func (m *ManagerImpl) ExportProfile(id, filePath string) error {
	m.mu.RLock()
	profile, exists := m.profiles[id]
	m.mu.RUnlock()

	if !exists {
		return models.ErrProfileNotFound
	}

	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// CloneProfile 复制Profile
func (m *ManagerImpl) CloneProfile(id, newName string) (*models.Profile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	original, exists := m.profiles[id]
	if !exists {
		return nil, models.ErrProfileNotFound
	}

	// 检查新名称是否已存在
	for _, profile := range m.profiles {
		if profile.Name == newName {
			return nil, models.ErrProfileExists
		}
	}

	cloned := original.Clone()
	cloned.ID = fmt.Sprintf("%d-%x", time.Now().UnixNano(), []byte{1, 2, 3, 4})
	cloned.Name = newName
	now := time.Now()
	cloned.CreatedAt = now
	cloned.UpdatedAt = now
	cloned.IsActive = false

	m.profiles[cloned.ID] = cloned

	if err := m.saveProfiles(); err != nil {
		return nil, fmt.Errorf("failed to save cloned profile: %w", err)
	}

	return cloned, nil
}

// SearchProfiles 搜索Profile
func (m *ManagerImpl) SearchProfiles(query string) ([]*models.ProfileSummary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []*models.ProfileSummary
	for _, profile := range m.profiles {
		// 简单的字符串匹配搜索
		if containsIgnoreCase(profile.Name, query) ||
			containsIgnoreCase(profile.Description, query) {
			summary := profile.ToSummary()
			results = append(results, &summary)
		}
	}

	// 按更新时间排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].UpdatedAt.After(results[j].UpdatedAt)
	})

	return results, nil
}

// loadProfiles 从文件加载Profile数据
func (m *ManagerImpl) loadProfiles() error {
	if _, err := os.Stat(m.profileFile); os.IsNotExist(err) {
		return nil // 文件不存在，返回空数据
	}

	data, err := os.ReadFile(m.profileFile)
	if err != nil {
		return err
	}

	var profileData struct {
		Profiles map[string]*models.Profile `json:"profiles"`
		ActiveID string                     `json:"active_id"`
	}

	if err := json.Unmarshal(data, &profileData); err != nil {
		return err
	}

	m.profiles = profileData.Profiles
	m.activeID = profileData.ActiveID

	if m.profiles == nil {
		m.profiles = make(map[string]*models.Profile)
	}

	return nil
}

// saveProfiles 保存Profile数据到文件
func (m *ManagerImpl) saveProfiles() error {
	profileData := struct {
		Profiles map[string]*models.Profile `json:"profiles"`
		ActiveID string                     `json:"active_id"`
	}{
		Profiles: m.profiles,
		ActiveID: m.activeID,
	}

	data, err := json.MarshalIndent(profileData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.profileFile, data, 0644)
}

// containsIgnoreCase 不区分大小写的字符串包含检查
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		len(substr) > 0 &&
		contains(strings.ToLower(s), strings.ToLower(substr))
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(len(substr) == 0 || indexOfSubstring(s, substr) >= 0)
}

// indexOfSubstring 查找子字符串的索引
func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
