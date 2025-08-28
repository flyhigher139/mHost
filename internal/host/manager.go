package host

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/flyhigher139/mhost/pkg/models"
)

// Manager 定义hosts文件管理器接口
type Manager interface {
	// ReadHostsFile 读取hosts文件内容
	ReadHostsFile() ([]string, error)

	// WriteHostsFile 写入hosts文件内容
	WriteHostsFile(lines []string) error

	// ApplyProfile 应用Profile到hosts文件
	ApplyProfile(profile *models.Profile) error

	// BackupHostsFile 备份当前hosts文件
	BackupHostsFile() (*models.Backup, error)

	// RestoreFromBackup 从备份恢复hosts文件
	RestoreFromBackup(backup *models.Backup) error

	// GetHostsFilePath 获取hosts文件路径
	GetHostsFilePath() string

	// ValidateHostsFile 验证hosts文件格式
	ValidateHostsFile() error

	// ParseHostsFile 解析hosts文件为HostEntry列表
	ParseHostsFile() ([]*models.HostEntry, error)

	// GetManagedSection 获取mHost管理的section
	GetManagedSection() ([]string, error)

	// UpdateManagedSection 更新mHost管理的section
	UpdateManagedSection(entries []*models.HostEntry) error
}

// ManagerImpl hosts文件管理器实现
type ManagerImpl struct {
	hostsPath   string
	backupDir   string
	managedMark string
}

// NewManager 创建新的hosts文件管理器
func NewManager(hostsPath, backupDir string) Manager {
	if hostsPath == "" {
		hostsPath = getDefaultHostsPath()
	}

	return &ManagerImpl{
		hostsPath:   hostsPath,
		backupDir:   backupDir,
		managedMark: "# mHost managed section",
	}
}

// getDefaultHostsPath 获取默认hosts文件路径
func getDefaultHostsPath() string {
	return "/etc/hosts"
}

// ReadHostsFile 读取hosts文件内容
func (m *ManagerImpl) ReadHostsFile() ([]string, error) {
	file, err := os.Open(m.hostsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open hosts file: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read hosts file: %w", err)
	}

	return lines, nil
}

// WriteHostsFile 写入hosts文件内容
func (m *ManagerImpl) WriteHostsFile(lines []string) error {
	// 创建临时文件
	tempFile := m.hostsPath + ".tmp"
	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer file.Close()

	// 写入内容
	for _, line := range lines {
		if _, err := file.WriteString(line + "\n"); err != nil {
			os.Remove(tempFile)
			return fmt.Errorf("failed to write to temp file: %w", err)
		}
	}

	// 同步到磁盘
	if err := file.Sync(); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	file.Close()

	// 原子性替换
	if err := os.Rename(tempFile, m.hostsPath); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("failed to replace hosts file: %w", err)
	}

	return nil
}

// ApplyProfile 应用Profile到hosts文件
func (m *ManagerImpl) ApplyProfile(profile *models.Profile) error {
	if profile == nil {
		return models.ErrInvalidProfile
	}

	// 读取当前hosts文件
	lines, err := m.ReadHostsFile()
	if err != nil {
		return err
	}

	// 移除现有的mHost管理section
	newLines := m.removeManagedSection(lines)

	// 添加新的mHost管理section
	if len(profile.Entries) > 0 {
		newLines = append(newLines, "")
		newLines = append(newLines, m.managedMark+" START")
		newLines = append(newLines, fmt.Sprintf("# Profile: %s", profile.Name))
		newLines = append(newLines, fmt.Sprintf("# Applied at: %s", time.Now().Format(time.RFC3339)))

		for _, entry := range profile.Entries {
			if entry.Enabled {
				line := fmt.Sprintf("%s\t%s", entry.IP, entry.Hostname)
				if entry.Comment != "" {
					line += fmt.Sprintf("\t# %s", entry.Comment)
				}
				newLines = append(newLines, line)
			}
		}

		newLines = append(newLines, m.managedMark+" END")
	}

	// 写入hosts文件
	return m.WriteHostsFile(newLines)
}

// BackupHostsFile 备份当前hosts文件
func (m *ManagerImpl) BackupHostsFile() (*models.Backup, error) {
	// 确保备份目录存在
	if err := os.MkdirAll(m.backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	// 生成备份文件名
	timestamp := time.Now().Format("20060102_150405")
	backupFileName := fmt.Sprintf("hosts_backup_%s.txt", timestamp)
	backupPath := filepath.Join(m.backupDir, backupFileName)

	// 复制hosts文件
	srcFile, err := os.Open(m.hostsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open hosts file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup file: %w", err)
	}
	defer dstFile.Close()

	size, err := io.Copy(dstFile, srcFile)
	if err != nil {
		return nil, fmt.Errorf("failed to copy hosts file: %w", err)
	}

	// 创建备份记录
	backup := &models.Backup{
		ID:           fmt.Sprintf("backup_%d", time.Now().Unix()),
		Type:         models.BackupTypeManual,
		FilePath:     backupPath,
		OriginalPath: m.hostsPath,
		Size:         size,
		CreatedAt:    time.Now(),
		Metadata: models.BackupMetadata{
			Version:     "1.0",
			Description: "Manual hosts file backup",
			Tags:        []string{"manual", "hosts"},
		},
	}

	return backup, nil
}

// RestoreFromBackup 从备份恢复hosts文件
func (m *ManagerImpl) RestoreFromBackup(backup *models.Backup) error {
	if backup == nil {
		return models.ErrInvalidBackup
	}

	// 检查备份文件是否存在
	if _, err := os.Stat(backup.FilePath); os.IsNotExist(err) {
		return models.ErrBackupNotFound
	}

	// 复制备份文件到hosts文件
	srcFile, err := os.Open(backup.FilePath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(m.hostsPath)
	if err != nil {
		return fmt.Errorf("failed to create hosts file: %w", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to restore hosts file: %w", err)
	}

	return nil
}

// GetHostsFilePath 获取hosts文件路径
func (m *ManagerImpl) GetHostsFilePath() string {
	return m.hostsPath
}

// ValidateHostsFile 验证hosts文件格式
func (m *ManagerImpl) ValidateHostsFile() error {
	lines, err := m.ReadHostsFile()
	if err != nil {
		return err
	}

	// IP地址正则表达式
	ipv4Regex := regexp.MustCompile(`^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)
	ipv6Regex := regexp.MustCompile(`^([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$|^::1$|^::$`)
	hostnameRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?))*$`)

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// 跳过空行和注释行
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 分割IP和hostname
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return fmt.Errorf("invalid hosts entry at line %d: %s", i+1, line)
		}

		ip := fields[0]
		hostname := fields[1]

		// 验证IP地址
		if !ipv4Regex.MatchString(ip) && !ipv6Regex.MatchString(ip) {
			return fmt.Errorf("invalid IP address at line %d: %s", i+1, ip)
		}

		// 验证hostname
		if !hostnameRegex.MatchString(hostname) {
			return fmt.Errorf("invalid hostname at line %d: %s", i+1, hostname)
		}
	}

	return nil
}

// ParseHostsFile 解析hosts文件为HostEntry列表
func (m *ManagerImpl) ParseHostsFile() ([]*models.HostEntry, error) {
	lines, err := m.ReadHostsFile()
	if err != nil {
		return nil, err
	}

	var entries []*models.HostEntry

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 跳过空行和注释行
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 分割IP、hostname和注释
		parts := strings.SplitN(line, "#", 2)
		hostPart := strings.TrimSpace(parts[0])
		comment := ""
		if len(parts) > 1 {
			comment = strings.TrimSpace(parts[1])
		}

		fields := strings.Fields(hostPart)
		if len(fields) < 2 {
			continue
		}

		ip := fields[0]
		// 处理多个hostname
		for _, hostname := range fields[1:] {
			entry := &models.HostEntry{
				ID:        fmt.Sprintf("%s_%s_%d", ip, hostname, time.Now().UnixNano()),
				IP:        ip,
				Hostname:  hostname,
				Comment:   comment,
				Enabled:   true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			entries = append(entries, entry)
		}
	}

	return entries, nil
}

// GetManagedSection 获取mHost管理的section
func (m *ManagerImpl) GetManagedSection() ([]string, error) {
	lines, err := m.ReadHostsFile()
	if err != nil {
		return nil, err
	}

	var managedLines []string
	inManagedSection := false

	for _, line := range lines {
		if strings.Contains(line, m.managedMark+" START") {
			inManagedSection = true
			continue
		}
		if strings.Contains(line, m.managedMark+" END") {
			inManagedSection = false
			continue
		}
		if inManagedSection {
			managedLines = append(managedLines, line)
		}
	}

	return managedLines, nil
}

// UpdateManagedSection 更新mHost管理的section
func (m *ManagerImpl) UpdateManagedSection(entries []*models.HostEntry) error {
	// 读取当前hosts文件
	lines, err := m.ReadHostsFile()
	if err != nil {
		return err
	}

	// 移除现有的mHost管理section
	newLines := m.removeManagedSection(lines)

	// 添加新的mHost管理section
	if len(entries) > 0 {
		newLines = append(newLines, "")
		newLines = append(newLines, m.managedMark+" START")
		newLines = append(newLines, fmt.Sprintf("# Updated at: %s", time.Now().Format(time.RFC3339)))

		for _, entry := range entries {
			if entry.Enabled {
				line := fmt.Sprintf("%s\t%s", entry.IP, entry.Hostname)
				if entry.Comment != "" {
					line += fmt.Sprintf("\t# %s", entry.Comment)
				}
				newLines = append(newLines, line)
			}
		}

		newLines = append(newLines, m.managedMark+" END")
	}

	// 写入hosts文件
	return m.WriteHostsFile(newLines)
}

// removeManagedSection 移除mHost管理的section
func (m *ManagerImpl) removeManagedSection(lines []string) []string {
	var newLines []string
	inManagedSection := false

	for _, line := range lines {
		if strings.Contains(line, m.managedMark+" START") {
			inManagedSection = true
			continue
		}
		if strings.Contains(line, m.managedMark+" END") {
			inManagedSection = false
			continue
		}
		if !inManagedSection {
			newLines = append(newLines, line)
		}
	}

	return newLines
}
