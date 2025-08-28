package helper

import (
	"fmt"
	"log"
	"os"
	"time"
)

// SimpleLogger 简单的日志实现
type SimpleLogger struct {
	logger *log.Logger
	level  LogLevel
}

// LogLevel 日志级别
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// NewSimpleLogger 创建简单日志器
func NewSimpleLogger(level LogLevel) *SimpleLogger {
	return &SimpleLogger{
		logger: log.New(os.Stdout, "", log.LstdFlags),
		level:  level,
	}
}

// Debug 调试日志
func (l *SimpleLogger) Debug(msg string, keysAndValues ...interface{}) {
	if l.level <= LogLevelDebug {
		l.log("DEBUG", msg, keysAndValues...)
	}
}

// Info 信息日志
func (l *SimpleLogger) Info(msg string, keysAndValues ...interface{}) {
	if l.level <= LogLevelInfo {
		l.log("INFO", msg, keysAndValues...)
	}
}

// Warn 警告日志
func (l *SimpleLogger) Warn(msg string, keysAndValues ...interface{}) {
	if l.level <= LogLevelWarn {
		l.log("WARN", msg, keysAndValues...)
	}
}

// Error 错误日志
func (l *SimpleLogger) Error(msg string, keysAndValues ...interface{}) {
	if l.level <= LogLevelError {
		l.log("ERROR", msg, keysAndValues...)
	}
}

// log 内部日志方法
func (l *SimpleLogger) log(level, msg string, keysAndValues ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logMsg := fmt.Sprintf("[%s] %s: %s", timestamp, level, msg)
	
	// 处理键值对
	if len(keysAndValues) > 0 {
		logMsg += " |"
		for i := 0; i < len(keysAndValues); i += 2 {
			if i+1 < len(keysAndValues) {
				logMsg += fmt.Sprintf(" %v=%v", keysAndValues[i], keysAndValues[i+1])
			} else {
				logMsg += fmt.Sprintf(" %v", keysAndValues[i])
			}
		}
	}
	
	l.logger.Println(logMsg)
}