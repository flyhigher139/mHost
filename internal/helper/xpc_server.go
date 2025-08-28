package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// XPCRequestHandler XPC请求处理函数类型
type XPCRequestHandler func(*XPCRequest) *XPCResponse

// XPCServerImpl XPC服务器实现
type XPCServerImpl struct {
	serviceName string
	logger      Logger
	handler     XPCRequestHandler
	running     bool
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.RWMutex
	stats       *XPCServerStats
}

// XPCServerStats XPC服务器统计信息
type XPCServerStats struct {
	TotalRequests    int64     `json:"total_requests"`
	SuccessRequests  int64     `json:"success_requests"`
	FailedRequests   int64     `json:"failed_requests"`
	StartTime        time.Time `json:"start_time"`
	LastRequestTime  time.Time `json:"last_request_time"`
	AverageLatency   float64   `json:"average_latency_ms"`
	mu               sync.RWMutex
}

// NewXPCServerImpl 创建新的XPC服务器实现
func NewXPCServerImpl(serviceName string, logger Logger) (*XPCServerImpl, error) {
	if serviceName == "" {
		return nil, fmt.Errorf("service name cannot be empty")
	}

	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &XPCServerImpl{
		serviceName: serviceName,
		logger:      logger,
		running:     false,
		ctx:         ctx,
		cancel:      cancel,
		stats: &XPCServerStats{
			StartTime: time.Now(),
		},
	}, nil
}

// Start 启动XPC服务器
func (s *XPCServerImpl) Start(ctx context.Context, handler XPCRequestHandler) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("XPC server is already running")
	}

	if handler == nil {
		return fmt.Errorf("request handler cannot be nil")
	}

	s.handler = handler
	s.logger.Info("Starting XPC server", "service", s.serviceName)

	// 在实际实现中，这里会注册XPC服务
	// 目前使用模拟实现
	if err := s.registerXPCService(); err != nil {
		return fmt.Errorf("failed to register XPC service: %w", err)
	}

	s.running = true
	s.stats.StartTime = time.Now()

	// 启动消息处理循环
	go s.messageLoop()

	s.logger.Info("XPC server started successfully", "service", s.serviceName)
	return nil
}

// Stop 停止XPC服务器
func (s *XPCServerImpl) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	s.logger.Info("Stopping XPC server", "service", s.serviceName)

	// 取消上下文
	s.cancel()

	// 注销XPC服务
	if err := s.unregisterXPCService(); err != nil {
		s.logger.Error("Error unregistering XPC service", "error", err)
	}

	s.running = false
	s.logger.Info("XPC server stopped successfully")

	return nil
}

// IsRunning 检查服务器是否正在运行
func (s *XPCServerImpl) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetStats 获取服务器统计信息
func (s *XPCServerImpl) GetStats() *XPCServerStats {
	s.stats.mu.RLock()
	defer s.stats.mu.RUnlock()

	// 返回统计信息的副本
	return &XPCServerStats{
		TotalRequests:   s.stats.TotalRequests,
		SuccessRequests: s.stats.SuccessRequests,
		FailedRequests:  s.stats.FailedRequests,
		StartTime:       s.stats.StartTime,
		LastRequestTime: s.stats.LastRequestTime,
		AverageLatency:  s.stats.AverageLatency,
	}
}

// messageLoop 消息处理循环
func (s *XPCServerImpl) messageLoop() {
	s.logger.Debug("Starting XPC message loop")

	for {
		select {
		case <-s.ctx.Done():
			s.logger.Debug("XPC message loop stopped")
			return
		default:
			// 在实际实现中，这里会等待XPC消息
			// 目前使用模拟实现
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// handleMessage 处理XPC消息
func (s *XPCServerImpl) handleMessage(messageData []byte) []byte {
	start := time.Now()

	// 更新统计信息
	s.updateStats(true, false, 0)

	// 反序列化请求
	var req XPCRequest
	if err := json.Unmarshal(messageData, &req); err != nil {
		s.logger.Error("Failed to unmarshal XPC request", "error", err)
		s.updateStats(false, true, time.Since(start))
		return s.createErrorResponse("Invalid request format")
	}

	// 验证请求
	if err := s.validateRequest(&req); err != nil {
		s.logger.Error("Invalid XPC request", "error", err, "operation", req.Operation)
		s.updateStats(false, true, time.Since(start))
		return s.createErrorResponse(fmt.Sprintf("Invalid request: %v", err))
	}

	s.logger.Debug("Processing XPC request", "operation", req.Operation, "client", req.ClientID)

	// 调用处理函数
	resp := s.handler(&req)
	if resp == nil {
		s.logger.Error("Handler returned nil response", "operation", req.Operation)
		s.updateStats(false, true, time.Since(start))
		return s.createErrorResponse("Internal server error")
	}

	// 设置响应时间戳
	resp.Timestamp = time.Now()

	// 序列化响应
	respData, err := json.Marshal(resp)
	if err != nil {
		s.logger.Error("Failed to marshal XPC response", "error", err)
		s.updateStats(false, true, time.Since(start))
		return s.createErrorResponse("Failed to serialize response")
	}

	// 更新统计信息
	latency := time.Since(start)
	if resp.Success {
		s.updateStats(false, false, latency)
		s.logger.Debug("XPC request completed successfully", "operation", req.Operation, "latency", latency)
	} else {
		s.updateStats(false, true, latency)
		s.logger.Warn("XPC request failed", "operation", req.Operation, "error", resp.Error, "latency", latency)
	}

	return respData
}

// validateRequest 验证XPC请求
func (s *XPCServerImpl) validateRequest(req *XPCRequest) error {
	if req.Operation == "" {
		return fmt.Errorf("operation cannot be empty")
	}

	if req.ClientID == "" {
		return fmt.Errorf("client ID cannot be empty")
	}

	if req.Timestamp.IsZero() {
		return fmt.Errorf("timestamp cannot be zero")
	}

	// 检查请求是否过期（5分钟）
	if time.Since(req.Timestamp) > 5*time.Minute {
		return fmt.Errorf("request expired")
	}

	return nil
}

// createErrorResponse 创建错误响应
func (s *XPCServerImpl) createErrorResponse(errorMsg string) []byte {
	resp := &XPCResponse{
		Success:   false,
		Error:     errorMsg,
		Timestamp: time.Now(),
	}

	data, _ := json.Marshal(resp)
	return data
}

// updateStats 更新统计信息
func (s *XPCServerImpl) updateStats(isNew, isFailed bool, latency time.Duration) {
	s.stats.mu.Lock()
	defer s.stats.mu.Unlock()

	if isNew {
		s.stats.TotalRequests++
		s.stats.LastRequestTime = time.Now()
	}

	if isFailed {
		s.stats.FailedRequests++
	} else if !isNew {
		s.stats.SuccessRequests++
	}

	if latency > 0 {
		// 计算平均延迟（简单移动平均）
		latencyMs := float64(latency.Nanoseconds()) / 1e6
		if s.stats.AverageLatency == 0 {
			s.stats.AverageLatency = latencyMs
		} else {
			s.stats.AverageLatency = (s.stats.AverageLatency + latencyMs) / 2
		}
	}
}

// registerXPCService 注册XPC服务（模拟实现）
func (s *XPCServerImpl) registerXPCService() error {
	// 在实际实现中，这里会使用macOS的XPC API注册服务
	s.logger.Debug("Registering XPC service", "service", s.serviceName)
	return nil
}

// unregisterXPCService 注销XPC服务（模拟实现）
func (s *XPCServerImpl) unregisterXPCService() error {
	// 在实际实现中，这里会使用macOS的XPC API注销服务
	s.logger.Debug("Unregistering XPC service", "service", s.serviceName)
	return nil
}

// XPCServiceManager XPC服务管理器
type XPCServiceManager struct {
	logger  Logger
	servers map[string]*XPCServerImpl
	mu      sync.RWMutex
}

// NewXPCServiceManager 创建XPC服务管理器
func NewXPCServiceManager(logger Logger) *XPCServiceManager {
	return &XPCServiceManager{
		logger:  logger,
		servers: make(map[string]*XPCServerImpl),
	}
}

// RegisterService 注册XPC服务
func (m *XPCServiceManager) RegisterService(serviceName string, handler XPCRequestHandler) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.servers[serviceName]; exists {
		return fmt.Errorf("service %s already registered", serviceName)
	}

	server, err := NewXPCServerImpl(serviceName, m.logger)
	if err != nil {
		return fmt.Errorf("failed to create XPC server: %w", err)
	}

	if err := server.Start(context.Background(), handler); err != nil {
		return fmt.Errorf("failed to start XPC server: %w", err)
	}

	m.servers[serviceName] = server
	m.logger.Info("XPC service registered", "service", serviceName)

	return nil
}

// UnregisterService 注销XPC服务
func (m *XPCServiceManager) UnregisterService(serviceName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	server, exists := m.servers[serviceName]
	if !exists {
		return fmt.Errorf("service %s not found", serviceName)
	}

	if err := server.Stop(); err != nil {
		m.logger.Error("Error stopping XPC server", "service", serviceName, "error", err)
	}

	delete(m.servers, serviceName)
	m.logger.Info("XPC service unregistered", "service", serviceName)

	return nil
}

// GetServiceStats 获取服务统计信息
func (m *XPCServiceManager) GetServiceStats(serviceName string) (*XPCServerStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	server, exists := m.servers[serviceName]
	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	return server.GetStats(), nil
}

// Shutdown 关闭所有服务
func (m *XPCServiceManager) Shutdown() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for serviceName, server := range m.servers {
		if err := server.Stop(); err != nil {
			m.logger.Error("Error stopping XPC server during shutdown", "service", serviceName, "error", err)
		}
	}

	m.servers = make(map[string]*XPCServerImpl)
	m.logger.Info("All XPC services shut down")

	return nil
}