package models

import (
	"time"
)

// Profile 表示一个hosts配置文件
type Profile struct {
	ID          string       `json:"id"`          // 唯一标识符
	Name        string       `json:"name"`        // 配置文件名称
	Description string       `json:"description"` // 描述信息
	Entries     []*HostEntry `json:"entries"`     // hosts条目列表
	CreatedAt   time.Time    `json:"created_at"`  // 创建时间
	UpdatedAt   time.Time    `json:"updated_at"`  // 更新时间
	IsActive    bool         `json:"is_active"`   // 是否为当前激活的配置
	Tags        []string     `json:"tags"`        // 标签
}

// HostEntry hosts文件条目
type HostEntry struct {
	ID        string    `json:"id"`         // 唯一标识符
	IP        string    `json:"ip"`         // IP地址
	Hostname  string    `json:"hostname"`   // 主机名
	Comment   string    `json:"comment"`    // 注释
	Enabled   bool      `json:"enabled"`    // 是否启用
	CreatedAt time.Time `json:"created_at"` // 创建时间
	UpdatedAt time.Time `json:"updated_at"` // 更新时间
}

// ProfileSummary 用于列表显示的简化Profile信息
type ProfileSummary struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	EntryCount  int       `json:"entry_count"`
	IsActive    bool      `json:"is_active"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewProfile 创建一个新的Profile实例
func NewProfile(name, description string) *Profile {
	now := time.Now()
	return &Profile{
		ID:          generateID(),
		Name:        name,
		Description: description,
		Entries:     make([]*HostEntry, 0),
		CreatedAt:   now,
		UpdatedAt:   now,
		IsActive:    false,
		Tags:        make([]string, 0),
	}
}

// NewHostEntry 创建新的HostEntry
func NewHostEntry(ip, hostname, comment string) *HostEntry {
	now := time.Now()
	return &HostEntry{
		ID:        generateID(),
		IP:        ip,
		Hostname:  hostname,
		Comment:   comment,
		Enabled:   true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// AddEntry 向Profile添加一个hosts条目
func (p *Profile) AddEntry(entry *HostEntry) {
	p.Entries = append(p.Entries, entry)
	p.UpdateTimestamp()
}

// RemoveEntry 从Profile中移除指定ID的hosts条目
func (p *Profile) RemoveEntry(entryID string) bool {
	for i, entry := range p.Entries {
		if entry.ID == entryID {
			p.Entries = append(p.Entries[:i], p.Entries[i+1:]...)
			p.UpdateTimestamp()
			return true
		}
	}
	return false
}

// UpdateEntry 更新指定ID的hosts条目
func (p *Profile) UpdateEntry(entryID string, updatedEntry *HostEntry) bool {
	for i, entry := range p.Entries {
		if entry.ID == entryID {
			updatedEntry.ID = entryID // 保持原有ID
			p.Entries[i] = updatedEntry
			p.UpdateTimestamp()
			return true
		}
	}
	return false
}

// GetEntry 根据ID获取hosts条目
func (p *Profile) GetEntry(entryID string) (*HostEntry, bool) {
	for _, entry := range p.Entries {
		if entry.ID == entryID {
			return entry, true
		}
	}
	return nil, false
}

// UpdateTimestamp 更新Profile的时间戳
func (p *Profile) UpdateTimestamp() {
	p.UpdatedAt = time.Now()
}

// ToSummary 将Profile转换为ProfileSummary
func (p *Profile) ToSummary() ProfileSummary {
	return ProfileSummary{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		EntryCount:  len(p.Entries),
		IsActive:    p.IsActive,
		UpdatedAt:   p.UpdatedAt,
	}
}

// Clone 创建Profile的深拷贝
func (p *Profile) Clone() *Profile {
	cloned := *p
	cloned.Entries = make([]*HostEntry, len(p.Entries))
	for i, entry := range p.Entries {
		entryCopy := *entry
		cloned.Entries[i] = &entryCopy
	}
	cloned.Tags = make([]string, len(p.Tags))
	copy(cloned.Tags, p.Tags)
	return &cloned
}

// Validate 验证Profile数据的有效性
func (p *Profile) Validate() error {
	if p.Name == "" {
		return ErrInvalidProfileName
	}

	for _, entry := range p.Entries {
		if err := entry.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate 验证HostEntry数据的有效性
func (h *HostEntry) Validate() error {
	if h.IP == "" {
		return ErrInvalidIP
	}
	if h.Hostname == "" {
		return ErrInvalidHostname
	}
	return nil
}
