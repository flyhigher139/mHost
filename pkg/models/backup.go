package models

import (
	"fmt"
	"time"
)

// BackupType 备份类型
type BackupType string

const (
	BackupTypeManual    BackupType = "manual"    // 手动备份
	BackupTypeAutomatic BackupType = "automatic" // 自动备份
	BackupTypeSystem    BackupType = "system"    // 系统备份
)

// Backup 备份信息结构体
type Backup struct {
	ID           string           `json:"id"`
	Type         BackupType       `json:"type"`
	FilePath     string           `json:"file_path"`
	OriginalPath string           `json:"original_path"`
	Size         int64            `json:"size"`
	CreatedAt    time.Time        `json:"created_at"`
	ExpiresAt    *time.Time       `json:"expires_at,omitempty"`
	Metadata     BackupMetadata   `json:"metadata"`
	Validation   BackupValidation `json:"validation"`
}

// BackupMetadata 备份元数据
type BackupMetadata struct {
	Version     string            `json:"version"`     // 备份版本
	Description string            `json:"description"` // 备份描述
	Checksum    string            `json:"checksum"`    // 文件校验和
	Compressed  bool              `json:"compressed"`  // 是否压缩
	Encrypted   bool              `json:"encrypted"`   // 是否加密
	ProfileID   string            `json:"profile_id"`  // 关联的Profile ID
	Tags        []string          `json:"tags"`        // 标签
	CustomData  map[string]string `json:"custom_data"` // 自定义数据
}

// BackupValidation 备份验证结果
type BackupValidation struct {
	IsValid       bool     `json:"is_valid"`       // 是否有效
	Errors        []string `json:"errors"`         // 错误列表
	Warnings      []string `json:"warnings"`       // 警告列表
	ChecksumMatch bool     `json:"checksum_match"` // 校验和是否匹配
	FileExists    bool     `json:"file_exists"`    // 文件是否存在
	CanRestore    bool     `json:"can_restore"`    // 是否可以恢复
}

// BackupSummary 备份摘要信息
type BackupSummary struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Type        BackupType `json:"type"`
	Size        int64      `json:"size"`
	CreatedAt   time.Time  `json:"created_at"`
	IsValid     bool       `json:"is_valid"`
	Description string     `json:"description"`
}

// NewBackup 创建新的备份记录
func NewBackup(backupType BackupType, filePath, originalPath string, size int64) *Backup {
	return &Backup{
		ID:           generateID(),
		Type:         backupType,
		FilePath:     filePath,
		OriginalPath: originalPath,
		Size:         size,
		CreatedAt:    time.Now(),
		Metadata:     BackupMetadata{},
		Validation:   BackupValidation{},
	}
}

// ToSummary 将Backup转换为BackupSummary
func (b *Backup) ToSummary() BackupSummary {
	return BackupSummary{
		ID:          b.ID,
		Name:        b.Metadata.Description,
		Type:        b.Type,
		Size:        b.Size,
		CreatedAt:   b.CreatedAt,
		Description: b.Metadata.Description,
		IsValid:     b.Validation.IsValid,
	}
}

// Validate 验证备份数据的有效性
func (b *Backup) Validate() error {
	if b.FilePath == "" {
		return ErrFileNotFound
	}
	if b.Type != BackupTypeManual && b.Type != BackupTypeAutomatic && b.Type != BackupTypeSystem {
		return ErrInvalidConfig
	}
	return nil
}

// Clone 创建备份信息的深拷贝
func (b *Backup) Clone() *Backup {
	cloned := *b

	// 深拷贝元数据中的map
	cloned.Metadata.CustomData = make(map[string]string)
	for k, v := range b.Metadata.CustomData {
		cloned.Metadata.CustomData[k] = v
	}

	// 深拷贝Tags切片
	cloned.Metadata.Tags = make([]string, len(b.Metadata.Tags))
	copy(cloned.Metadata.Tags, b.Metadata.Tags)

	return &cloned
}

// IsExpired 检查备份是否已过期
func (b *Backup) IsExpired(retentionDays int) bool {
	if retentionDays <= 0 {
		return false
	}
	expiryDate := b.CreatedAt.AddDate(0, 0, retentionDays)
	return time.Now().After(expiryDate)
}

// GetSizeString 获取格式化的文件大小字符串
func (b *Backup) GetSizeString() string {
	if b.Size == 0 {
		return "Unknown"
	}

	const unit = 1024
	if b.Size < unit {
		return fmt.Sprintf("%d B", b.Size)
	}

	div, exp := int64(unit), 0
	for n := b.Size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(b.Size)/float64(div), "KMGTPE"[exp])
}
