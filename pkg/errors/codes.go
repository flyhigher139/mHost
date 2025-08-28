package errors

// 预定义错误代码
const (
	// 基础错误代码
	ErrCodeInvalidIP       = "INVALID_IP"
	ErrCodeInvalidHostname = "INVALID_HOSTNAME"
	ErrCodeProfileNotFound = "PROFILE_NOT_FOUND"
	ErrCodePermissionDenied = "PERMISSION_DENIED"
	ErrCodeFileNotFound    = "FILE_NOT_FOUND"
	ErrCodeBackupFailed    = "BACKUP_FAILED"
	ErrCodeRestoreFailed   = "RESTORE_FAILED"
	ErrCodeValidationFailed = "VALIDATION_FAILED"
	ErrCodeInvalidConfig   = "INVALID_CONFIG"
	ErrCodeConfigLoadFailed = "CONFIG_LOAD_FAILED"
	ErrCodeConfigSaveFailed = "CONFIG_SAVE_FAILED"

	// XPC 相关错误代码
	ErrCodeXPCConnectionFailed    = "XPC_CONNECTION_FAILED"
	ErrCodeXPCRequestTimeout      = "XPC_REQUEST_TIMEOUT"
	ErrCodeXPCInvalidRequest      = "XPC_INVALID_REQUEST"
	ErrCodeXPCInvalidResponse     = "XPC_INVALID_RESPONSE"
	ErrCodeXPCAuthenticationFailed = "XPC_AUTHENTICATION_FAILED"
	ErrCodeXPCServiceUnavailable  = "XPC_SERVICE_UNAVAILABLE"

	// Helper Tool 相关错误代码
	ErrCodeHelperNotInstalled     = "HELPER_NOT_INSTALLED"
	ErrCodeHelperInstallFailed    = "HELPER_INSTALL_FAILED"
	ErrCodeHelperUninstallFailed  = "HELPER_UNINSTALL_FAILED"
	ErrCodeHelperVersionMismatch  = "HELPER_VERSION_MISMATCH"
	ErrCodeHelperHealthCheckFailed = "HELPER_HEALTH_CHECK_FAILED"
	ErrCodeHelperRestartFailed    = "HELPER_RESTART_FAILED"

	// 权限相关错误代码
	ErrCodeInsufficientPrivileges = "INSUFFICIENT_PRIVILEGES"
	ErrCodeSignatureVerificationFailed = "SIGNATURE_VERIFICATION_FAILED"
	ErrCodeCertificateInvalid     = "CERTIFICATE_INVALID"
	ErrCodeAuditLogFailed         = "AUDIT_LOG_FAILED"

	// 文件操作相关错误代码
	ErrCodeFileReadFailed  = "FILE_READ_FAILED"
	ErrCodeFileWriteFailed = "FILE_WRITE_FAILED"
	ErrCodeInvalidFilePath = "INVALID_FILE_PATH"
	ErrCodeDirectoryCreateFailed = "DIRECTORY_CREATE_FAILED"

	// 备份相关错误代码
	ErrCodeBackupNotFound     = "BACKUP_NOT_FOUND"
	ErrCodeBackupCorrupted    = "BACKUP_CORRUPTED"
	ErrCodeBackupIndexFailed  = "BACKUP_INDEX_FAILED"
	ErrCodeBackupCleanupFailed = "BACKUP_CLEANUP_FAILED"

	// 安全相关错误代码
	ErrCodeSecurityViolation  = "SECURITY_VIOLATION"
	ErrCodeRateLimitExceeded  = "RATE_LIMIT_EXCEEDED"
	ErrCodeClientBlacklisted  = "CLIENT_BLACKLISTED"
	ErrCodeOperationNotAllowed = "OPERATION_NOT_ALLOWED"
	ErrCodeRequestExpired     = "REQUEST_EXPIRED"

	// 主机文件相关错误代码
	ErrCodeHostsFileCorrupted = "HOSTS_FILE_CORRUPTED"
	ErrCodeHostsValidationFailed = "HOSTS_VALIDATION_FAILED"
	ErrCodeHostEntryExists    = "HOST_ENTRY_EXISTS"
	ErrCodeHostEntryNotFound  = "HOST_ENTRY_NOT_FOUND"

	// Profile相关错误代码
	ErrCodeProfileExists      = "PROFILE_EXISTS"
	ErrCodeInvalidProfile     = "INVALID_PROFILE"
	ErrCodeInvalidProfileName = "INVALID_PROFILE_NAME"
	ErrCodeNoActiveProfile    = "NO_ACTIVE_PROFILE"
	ErrCodeActiveProfileError = "ACTIVE_PROFILE_ERROR"
)