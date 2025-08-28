package models

import "errors"

var (
	// Profile相关错误
	ErrInvalidProfile     = errors.New("invalid profile")
	ErrInvalidProfileName = errors.New("invalid profile name")
	ErrProfileNotFound    = errors.New("profile not found")
	ErrProfileExists      = errors.New("profile already exists")
	ErrNoActiveProfile    = errors.New("no active profile")
	ErrActiveProfile      = errors.New("active profile error")

	// HostEntry相关错误
	ErrInvalidIP         = errors.New("invalid IP address")
	ErrInvalidHostname   = errors.New("invalid hostname")
	ErrHostEntryExists   = errors.New("host entry already exists")
	ErrHostEntryNotFound = errors.New("host entry not found")

	// 备份相关错误
	ErrInvalidBackup  = errors.New("invalid backup")
	ErrBackupNotFound = errors.New("backup not found")
	ErrBackupFailed   = errors.New("backup operation failed")
	ErrRestoreFailed  = errors.New("restore operation failed")

	// 配置相关错误
	ErrInvalidConfig    = errors.New("invalid configuration")
	ErrConfigNotFound   = errors.New("configuration not found")
	ErrConfigLoadFailed = errors.New("failed to load configuration")
	ErrConfigSaveFailed = errors.New("failed to save configuration")

	// 文件操作相关错误
	ErrFileNotFound     = errors.New("file not found")
	ErrFileReadFailed   = errors.New("failed to read file")
	ErrFileWriteFailed  = errors.New("failed to write file")
	ErrInvalidFilePath  = errors.New("invalid file path")
	ErrPermissionDenied = errors.New("permission denied")
)
