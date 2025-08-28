package helper

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/flyhigher139/mhost/pkg/errors"
	"github.com/flyhigher139/mhost/pkg/logger"
)

// BackupManagerImpl 备份管理器实现
type BackupManagerImpl struct {
	logger      logger.Logger
	backupDir   string
	maxBackups  int
	mu          sync.RWMutex
	backupIndex map[string]*BackupInfo
}

// BackupInfo 备份信息
type BackupInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	OriginalPath string   `json:"original_path"`
	CreatedAt   time.Time `json:"created_at"`
	Size        int64     `json:"size"`
	Checksum    string    `json:"checksum"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
	Automatic   bool      `json:"automatic"`
}

// BackupConfig 备份配置
type BackupConfig struct {
	BackupDir       string        `json:"backup_dir"`
	MaxBackups      int           `json:"max_backups"`
	AutoCleanup     bool          `json:"auto_cleanup"`
	RetentionPeriod time.Duration `json:"retention_period"`
	CompressionLevel int          `json:"compression_level"`
}

// BackupStats 备份统计信息
type BackupStats struct {
	TotalBackups    int   `json:"total_backups"`
	TotalSize       int64 `json:"total_size"`
	OldestBackup    *time.Time `json:"oldest_backup,omitempty"`
	NewestBackup    *time.Time `json:"newest_backup,omitempty"`
	AutomaticBackups int   `json:"automatic_backups"`
	ManualBackups   int   `json:"manual_backups"`
}

// NewBackupManagerImpl 创建备份管理器实现
func NewBackupManagerImpl(logger logger.Logger, backupDir string, maxBackups int) (*BackupManagerImpl, error) {
	if backupDir == "" {
		backupDir = "/tmp/mhost-backups"
	}

	if maxBackups <= 0 {
		maxBackups = 10
	}

	// 确保备份目录存在
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, errors.NewFileSystemError(errors.ErrCodeDirectoryCreateFailed, "failed to create backup directory", err)
	}

	bm := &BackupManagerImpl{
		logger:      logger,
		backupDir:   backupDir,
		maxBackups:  maxBackups,
		backupIndex: make(map[string]*BackupInfo),
	}

	// 加载现有备份信息
	if err := bm.loadBackupIndex(); err != nil {
		logger.Warn("Failed to load backup index", "error", err)
	}

	return bm, nil
}

// CreateBackup 创建备份
func (bm *BackupManagerImpl) CreateBackup(sourcePath, name, description string, tags []string, automatic bool) (*BackupInfo, error) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	// 验证源文件
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return nil, errors.NewFileSystemError(errors.ErrCodeFileNotFound, fmt.Sprintf("source file does not exist: %s", sourcePath), err)
	}

	// 生成备份ID和路径
	backupID := bm.generateBackupID(sourcePath, name)
	backupPath := filepath.Join(bm.backupDir, fmt.Sprintf("%s.backup", backupID))

	// 检查是否已存在相同备份
	if existing, exists := bm.backupIndex[backupID]; exists {
		bm.logger.Info("Backup already exists", "id", backupID, "path", existing.Path)
		return existing, nil
	}

	// 复制文件
	if err := bm.copyFile(sourcePath, backupPath); err != nil {
		bm.logger.ErrorWithContext(nil, err, "Failed to copy file for backup", "source", sourcePath, "backup", backupPath)
		return nil, errors.NewFileSystemError(errors.ErrCodeBackupFailed, "failed to copy file", err)
	}

	// 计算校验和
	checksum, err := bm.calculateChecksum(backupPath)
	if err != nil {
		bm.logger.Warn("Failed to calculate checksum", "error", err)
		checksum = ""
	}

	// 获取文件大小
	fileInfo, err := os.Stat(backupPath)
	if err != nil {
		bm.logger.ErrorWithContext(nil, err, "Failed to get backup file info", "path", backupPath)
		return nil, errors.NewFileSystemError(errors.ErrCodeFileReadFailed, "failed to get backup file info", err)
	}

	// 创建备份信息
	backupInfo := &BackupInfo{
		ID:           backupID,
		Name:         name,
		Path:         backupPath,
		OriginalPath: sourcePath,
		CreatedAt:    time.Now(),
		Size:         fileInfo.Size(),
		Checksum:     checksum,
		Description:  description,
		Tags:         tags,
		Automatic:    automatic,
	}

	// 添加到索引
	bm.backupIndex[backupID] = backupInfo

	// 保存索引
	if err := bm.saveBackupIndex(); err != nil {
		bm.logger.Warn("Failed to save backup index", "error", err)
	}

	// 清理旧备份
	if err := bm.cleanupOldBackups(); err != nil {
		bm.logger.Warn("Failed to cleanup old backups", "error", err)
	}

	bm.logger.Info("Backup created successfully", "id", backupID, "path", backupPath, "size", fileInfo.Size())
	return backupInfo, nil
}

// RestoreBackup 恢复备份
func (bm *BackupManagerImpl) RestoreBackup(backupID, targetPath string) error {
	bm.mu.RLock()
	backupInfo, exists := bm.backupIndex[backupID]
	bm.mu.RUnlock()

	if !exists {
		return errors.NewValidationError(errors.ErrCodeBackupNotFound, fmt.Sprintf("backup not found: %s", backupID), nil)
	}

	// 验证备份文件存在
	if _, err := os.Stat(backupInfo.Path); os.IsNotExist(err) {
		bm.logger.Error("Backup file does not exist", "path", backupInfo.Path, "backup_id", backupID)
		return errors.NewFileSystemError(errors.ErrCodeFileNotFound, fmt.Sprintf("backup file does not exist: %s", backupInfo.Path), err)
	}

	// 验证校验和（如果存在）
	if backupInfo.Checksum != "" {
		currentChecksum, err := bm.calculateChecksum(backupInfo.Path)
		if err != nil {
			bm.logger.Warn("Failed to verify backup checksum", "error", err)
		} else if currentChecksum != backupInfo.Checksum {
			bm.logger.Error("Backup file corrupted: checksum mismatch", "backup_id", backupID, "expected", backupInfo.Checksum, "actual", currentChecksum)
		return errors.NewValidationError(errors.ErrCodeBackupCorrupted, "backup file corrupted: checksum mismatch", map[string]interface{}{
			"expected_checksum": backupInfo.Checksum,
			"actual_checksum":   currentChecksum,
		})
		}
	}

	// 如果目标路径为空，使用原始路径
	if targetPath == "" {
		targetPath = backupInfo.OriginalPath
	}

	// 创建目标目录（如果需要）
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// 复制文件
	if err := bm.copyFile(backupInfo.Path, targetPath); err != nil {
		return fmt.Errorf("failed to restore file: %w", err)
	}

	bm.logger.Info("Backup restored successfully", "id", backupID, "target", targetPath)
	return nil
}

// DeleteBackup 删除备份
func (bm *BackupManagerImpl) DeleteBackup(backupID string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	backupInfo, exists := bm.backupIndex[backupID]
	if !exists {
		return fmt.Errorf("backup not found: %s", backupID)
	}

	// 删除备份文件
	if err := os.Remove(backupInfo.Path); err != nil && !os.IsNotExist(err) {
		bm.logger.Warn("Failed to delete backup file", "path", backupInfo.Path, "error", err)
	}

	// 从索引中删除
	delete(bm.backupIndex, backupID)

	// 保存索引
	if err := bm.saveBackupIndex(); err != nil {
		bm.logger.Warn("Failed to save backup index", "error", err)
	}

	bm.logger.Info("Backup deleted successfully", "id", backupID)
	return nil
}

// ListBackups 列出所有备份
func (bm *BackupManagerImpl) ListBackups() []*BackupInfo {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	backups := make([]*BackupInfo, 0, len(bm.backupIndex))
	for _, backup := range bm.backupIndex {
		backups = append(backups, backup)
	}

	// 按创建时间排序（最新的在前）
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})

	return backups
}

// GetBackup 获取指定备份信息
func (bm *BackupManagerImpl) GetBackup(backupID string) (*BackupInfo, error) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	backupInfo, exists := bm.backupIndex[backupID]
	if !exists {
		return nil, fmt.Errorf("backup not found: %s", backupID)
	}

	return backupInfo, nil
}

// GetBackupStats 获取备份统计信息
func (bm *BackupManagerImpl) GetBackupStats() *BackupStats {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	stats := &BackupStats{
		TotalBackups: len(bm.backupIndex),
	}

	var totalSize int64
	var oldestTime, newestTime *time.Time
	automaticCount := 0
	manualCount := 0

	for _, backup := range bm.backupIndex {
		totalSize += backup.Size

		if backup.Automatic {
			automaticCount++
		} else {
			manualCount++
		}

		if oldestTime == nil || backup.CreatedAt.Before(*oldestTime) {
			oldestTime = &backup.CreatedAt
		}

		if newestTime == nil || backup.CreatedAt.After(*newestTime) {
			newestTime = &backup.CreatedAt
		}
	}

	stats.TotalSize = totalSize
	stats.OldestBackup = oldestTime
	stats.NewestBackup = newestTime
	stats.AutomaticBackups = automaticCount
	stats.ManualBackups = manualCount

	return stats
}

// CleanupOldBackups 清理旧备份
func (bm *BackupManagerImpl) CleanupOldBackups() error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	return bm.cleanupOldBackups()
}

// cleanupOldBackups 内部清理方法（需要持有锁）
func (bm *BackupManagerImpl) cleanupOldBackups() error {
	if len(bm.backupIndex) <= bm.maxBackups {
		return nil
	}

	// 获取所有备份并按时间排序
	backups := make([]*BackupInfo, 0, len(bm.backupIndex))
	for _, backup := range bm.backupIndex {
		backups = append(backups, backup)
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.Before(backups[j].CreatedAt)
	})

	// 删除最旧的备份
	toDelete := len(backups) - bm.maxBackups
	for i := 0; i < toDelete; i++ {
		backup := backups[i]
		bm.logger.Info("Cleaning up old backup", "id", backup.ID, "created", backup.CreatedAt)

		// 删除文件
		if err := os.Remove(backup.Path); err != nil && !os.IsNotExist(err) {
			bm.logger.Warn("Failed to delete backup file during cleanup", "path", backup.Path, "error", err)
		}

		// 从索引中删除
		delete(bm.backupIndex, backup.ID)
	}

	// 保存索引
	return bm.saveBackupIndex()
}

// generateBackupID 生成备份ID
func (bm *BackupManagerImpl) generateBackupID(sourcePath, name string) string {
	timestamp := time.Now().Format("20060102-150405")
	baseName := filepath.Base(sourcePath)
	if name != "" {
		baseName = name
	}
	return fmt.Sprintf("%s-%s-%s", baseName, timestamp, bm.shortHash(sourcePath))
}

// shortHash 生成短哈希
func (bm *BackupManagerImpl) shortHash(input string) string {
	hash := md5.Sum([]byte(input))
	return fmt.Sprintf("%x", hash)[:8]
}

// copyFile 复制文件
func (bm *BackupManagerImpl) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// 同步到磁盘
	return destFile.Sync()
}

// calculateChecksum 计算文件校验和
func (bm *BackupManagerImpl) calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// loadBackupIndex 加载备份索引
func (bm *BackupManagerImpl) loadBackupIndex() error {
	// 扫描备份目录
	entries, err := os.ReadDir(bm.backupDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".backup") {
			continue
		}

		backupPath := filepath.Join(bm.backupDir, entry.Name())
		backupID := strings.TrimSuffix(entry.Name(), ".backup")

		// 获取文件信息
		fileInfo, err := entry.Info()
		if err != nil {
			bm.logger.Warn("Failed to get backup file info", "path", backupPath, "error", err)
			continue
		}

		// 计算校验和
		checksum, err := bm.calculateChecksum(backupPath)
		if err != nil {
			bm.logger.Warn("Failed to calculate checksum for existing backup", "path", backupPath, "error", err)
			checksum = ""
		}

		// 创建备份信息（从文件名解析信息）
		backupInfo := &BackupInfo{
			ID:        backupID,
			Name:      bm.parseNameFromID(backupID),
			Path:      backupPath,
			CreatedAt: fileInfo.ModTime(),
			Size:      fileInfo.Size(),
			Checksum:  checksum,
			Automatic: false, // 默认为手动备份
		}

		bm.backupIndex[backupID] = backupInfo
	}

	bm.logger.Info("Loaded backup index", "count", len(bm.backupIndex))
	return nil
}

// saveBackupIndex 保存备份索引
func (bm *BackupManagerImpl) saveBackupIndex() error {
	// 这里可以实现将索引保存到文件的逻辑
	// 为了简化，暂时只记录日志
	bm.logger.Debug("Backup index saved", "count", len(bm.backupIndex))
	return nil
}

// parseNameFromID 从备份ID解析名称
func (bm *BackupManagerImpl) parseNameFromID(backupID string) string {
	parts := strings.Split(backupID, "-")
	if len(parts) > 0 {
		return parts[0]
	}
	return backupID
}

// ValidateBackup 验证备份完整性
func (bm *BackupManagerImpl) ValidateBackup(backupID string) error {
	bm.mu.RLock()
	backupInfo, exists := bm.backupIndex[backupID]
	bm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("backup not found: %s", backupID)
	}

	// 检查文件是否存在
	fileInfo, err := os.Stat(backupInfo.Path)
	if os.IsNotExist(err) {
		return fmt.Errorf("backup file does not exist: %s", backupInfo.Path)
	}
	if err != nil {
		return fmt.Errorf("failed to access backup file: %w", err)
	}

	// 检查文件大小
	if fileInfo.Size() != backupInfo.Size {
		return fmt.Errorf("backup file size mismatch: expected %d, got %d", backupInfo.Size, fileInfo.Size())
	}

	// 验证校验和
	if backupInfo.Checksum != "" {
		currentChecksum, err := bm.calculateChecksum(backupInfo.Path)
		if err != nil {
			return fmt.Errorf("failed to calculate checksum: %w", err)
		}
		if currentChecksum != backupInfo.Checksum {
			return fmt.Errorf("backup file corrupted: checksum mismatch")
		}
	}

	bm.logger.Debug("Backup validation passed", "id", backupID)
	return nil
}