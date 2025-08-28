package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"


)

// Logger 增强的日志接口
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	ErrorWithContext(ctx context.Context, err error, msg string, keysAndValues ...interface{})
	WithFields(fields map[string]interface{}) Logger
	WithContext(ctx context.Context) Logger
}

// LogLevel 日志级别
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// Field 日志字段
type Field struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

// LogEntry 日志条目
type LogEntry struct {
	Timestamp time.Time            `json:"timestamp"`
	Level     string               `json:"level"`
	Message   string               `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Error     *ErrorInfo           `json:"error,omitempty"`
	Context   *ContextInfo         `json:"context,omitempty"`
	Caller    *CallerInfo          `json:"caller,omitempty"`
}

// ErrorInfo 错误信息
type ErrorInfo struct {
	Message string                 `json:"message"`
	Code    string                 `json:"code,omitempty"`
	Type    string                 `json:"type,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
	Stack   string                 `json:"stack,omitempty"`
}

// ContextInfo 上下文信息
type ContextInfo struct {
	RequestID string `json:"request_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
	Operation string `json:"operation,omitempty"`
}

// CallerInfo 调用者信息
type CallerInfo struct {
	File     string `json:"file"`
	Function string `json:"function"`
	Line     int    `json:"line"`
}

// EnhancedLogger 增强的日志实现
type EnhancedLogger struct {
	logger     *log.Logger
	level      LogLevel
	fields     map[string]interface{}
	ctx        context.Context
	structured bool
	includeCaller bool
}

// NewEnhancedLogger 创建增强日志器
func NewEnhancedLogger(level LogLevel, structured bool) *EnhancedLogger {
	return &EnhancedLogger{
		logger:        log.New(os.Stdout, "", 0),
		level:         level,
		fields:        make(map[string]interface{}),
		structured:    structured,
		includeCaller: true,
	}
}

// NewFileLogger 创建文件日志器
func NewFileLogger(filePath string, level LogLevel, structured bool) (*EnhancedLogger, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &EnhancedLogger{
		logger:        log.New(file, "", 0),
		level:         level,
		fields:        make(map[string]interface{}),
		structured:    structured,
		includeCaller: true,
	}, nil
}

// Debug 调试日志
func (l *EnhancedLogger) Debug(msg string, keysAndValues ...interface{}) {
	if l.level <= LogLevelDebug {
		l.log("DEBUG", msg, nil, keysAndValues...)
	}
}

// Info 信息日志
func (l *EnhancedLogger) Info(msg string, keysAndValues ...interface{}) {
	if l.level <= LogLevelInfo {
		l.log("INFO", msg, nil, keysAndValues...)
	}
}

// Warn 警告日志
func (l *EnhancedLogger) Warn(msg string, keysAndValues ...interface{}) {
	if l.level <= LogLevelWarn {
		l.log("WARN", msg, nil, keysAndValues...)
	}
}

// Error 错误日志
func (l *EnhancedLogger) Error(msg string, keysAndValues ...interface{}) {
	if l.level <= LogLevelError {
		l.log("ERROR", msg, nil, keysAndValues...)
	}
}

// ErrorWithContext 带上下文的错误日志
func (l *EnhancedLogger) ErrorWithContext(ctx context.Context, err error, msg string, keysAndValues ...interface{}) {
	if l.level <= LogLevelError {
		logger := l.WithContext(ctx).(*EnhancedLogger)
		logger.log("ERROR", msg, err, keysAndValues...)
	}
}

// WithFields 添加字段
func (l *EnhancedLogger) WithFields(fields map[string]interface{}) Logger {
	newFields := make(map[string]interface{})
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}

	return &EnhancedLogger{
		logger:        l.logger,
		level:         l.level,
		fields:        newFields,
		ctx:           l.ctx,
		structured:    l.structured,
		includeCaller: l.includeCaller,
	}
}

// WithContext 添加上下文
func (l *EnhancedLogger) WithContext(ctx context.Context) Logger {
	return &EnhancedLogger{
		logger:        l.logger,
		level:         l.level,
		fields:        l.fields,
		ctx:           ctx,
		structured:    l.structured,
		includeCaller: l.includeCaller,
	}
}

// log 内部日志方法
func (l *EnhancedLogger) log(level, msg string, err error, keysAndValues ...interface{}) {
	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   msg,
		Fields:    l.buildFields(keysAndValues...),
	}

	// 添加错误信息
	if err != nil {
		entry.Error = l.buildErrorInfo(err)
	}

	// 添加上下文信息
	if l.ctx != nil {
		entry.Context = l.buildContextInfo(l.ctx)
	}

	// 添加调用者信息
	if l.includeCaller {
		entry.Caller = l.buildCallerInfo()
	}

	if l.structured {
		l.logStructured(entry)
	} else {
		l.logPlain(entry)
	}
}

// buildFields 构建字段
func (l *EnhancedLogger) buildFields(keysAndValues ...interface{}) map[string]interface{} {
	fields := make(map[string]interface{})

	// 添加基础字段
	for k, v := range l.fields {
		fields[k] = v
	}

	// 添加键值对
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key := fmt.Sprintf("%v", keysAndValues[i])
			fields[key] = keysAndValues[i+1]
		}
	}

	return fields
}

// buildErrorInfo 构建错误信息
func (l *EnhancedLogger) buildErrorInfo(err error) *ErrorInfo {
	errorInfo := &ErrorInfo{
		Message: err.Error(),
	}

	return errorInfo
}

// buildContextInfo 构建上下文信息
func (l *EnhancedLogger) buildContextInfo(ctx context.Context) *ContextInfo {
	contextInfo := &ContextInfo{}

	// 从上下文中提取信息
	if requestID := ctx.Value("request_id"); requestID != nil {
		contextInfo.RequestID = fmt.Sprintf("%v", requestID)
	}
	if userID := ctx.Value("user_id"); userID != nil {
		contextInfo.UserID = fmt.Sprintf("%v", userID)
	}
	if operation := ctx.Value("operation"); operation != nil {
		contextInfo.Operation = fmt.Sprintf("%v", operation)
	}

	return contextInfo
}

// buildCallerInfo 构建调用者信息
func (l *EnhancedLogger) buildCallerInfo() *CallerInfo {
	_, file, line, ok := runtime.Caller(3) // 跳过log, Debug/Info/Warn/Error, 调用者
	if !ok {
		return nil
	}

	pc, _, _, ok := runtime.Caller(3)
	if !ok {
		return nil
	}

	fn := runtime.FuncForPC(pc)
	funcName := "unknown"
	if fn != nil {
		funcName = fn.Name()
	}

	return &CallerInfo{
		File:     file,
		Function: funcName,
		Line:     line,
	}
}

// logStructured 结构化日志输出
func (l *EnhancedLogger) logStructured(entry *LogEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		l.logger.Printf("Failed to marshal log entry: %v", err)
		return
	}
	l.logger.Println(string(data))
}

// logPlain 普通日志输出
func (l *EnhancedLogger) logPlain(entry *LogEntry) {
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")
	logMsg := fmt.Sprintf("[%s] %s: %s", timestamp, entry.Level, entry.Message)

	// 添加字段
	if len(entry.Fields) > 0 {
		logMsg += " |"
		for k, v := range entry.Fields {
			logMsg += fmt.Sprintf(" %s=%v", k, v)
		}
	}

	// 添加错误信息
	if entry.Error != nil {
		logMsg += fmt.Sprintf(" | error=%s", entry.Error.Message)
		if entry.Error.Code != "" {
			logMsg += fmt.Sprintf(" code=%s", entry.Error.Code)
		}
	}

	// 添加调用者信息
	if entry.Caller != nil {
		logMsg += fmt.Sprintf(" | caller=%s:%d", entry.Caller.File, entry.Caller.Line)
	}

	l.logger.Println(logMsg)
}

// ErrorField 创建错误字段
func ErrorField(err error) Field {
	return Field{Key: "error", Value: err}
}

// StringField 创建字符串字段
func StringField(key, value string) Field {
	return Field{Key: key, Value: value}
}

// IntField 创建整数字段
func IntField(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// BoolField 创建布尔字段
func BoolField(key string, value bool) Field {
	return Field{Key: key, Value: value}
}