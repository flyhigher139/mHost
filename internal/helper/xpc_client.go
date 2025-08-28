package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// XPCClient XPC客户端，用于与Helper Tool通信
type XPCClient struct {
	serviceName string
	logger      Logger
	connected   bool
	mu          sync.RWMutex
	timeout     time.Duration
}

// NewXPCClient 创建新的XPC客户端
func NewXPCClient(serviceName string, logger Logger) *XPCClient {
	return &XPCClient{
		serviceName: serviceName,
		logger:      logger,
		connected:   false,
		timeout:     30 * time.Second,
	}
}

// Connect 连接到Helper Tool
func (c *XPCClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	c.logger.Info("Connecting to XPC service", "service", c.serviceName)

	// 在实际实现中，这里会使用macOS的XPC API
	// 目前使用模拟实现
	c.connected = true
	c.logger.Info("Connected to XPC service successfully")

	return nil
}

// Disconnect 断开与Helper Tool的连接
func (c *XPCClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	c.logger.Info("Disconnecting from XPC service")
	c.connected = false
	c.logger.Info("Disconnected from XPC service")

	return nil
}

// IsConnected 检查是否已连接
func (c *XPCClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// SendRequest 发送请求到Helper Tool
func (c *XPCClient) SendRequest(ctx context.Context, operation string, params map[string]interface{}) (*XPCResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("XPC client is not connected")
	}

	// 创建请求
	req := &XPCRequest{
		Operation:  operation,
		ClientID:   c.generateClientID(),
		Parameters: params,
		Timestamp:  time.Now(),
	}

	c.logger.Debug("Sending XPC request", "operation", operation, "client_id", req.ClientID)

	// 序列化请求
	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 发送请求并等待响应
	respData, err := c.sendXPCMessage(ctx, reqData)
	if err != nil {
		return nil, fmt.Errorf("failed to send XPC message: %w", err)
	}

	// 反序列化响应
	var resp XPCResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	resp.Timestamp = time.Now()
	c.logger.Debug("Received XPC response", "success", resp.Success, "client_id", req.ClientID)

	return &resp, nil
}

// WriteHosts 写入hosts文件
func (c *XPCClient) WriteHosts(ctx context.Context, entries []HostEntry) error {
	params := map[string]interface{}{
		"entries": entries,
	}

	resp, err := c.SendRequest(ctx, "write_hosts", params)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("write hosts failed: %s", resp.Error)
	}

	return nil
}

// BackupHosts 备份hosts文件
func (c *XPCClient) BackupHosts(ctx context.Context) (string, error) {
	resp, err := c.SendRequest(ctx, "backup_hosts", nil)
	if err != nil {
		return "", err
	}

	if !resp.Success {
		return "", fmt.Errorf("backup hosts failed: %s", resp.Error)
	}

	backupPath, ok := resp.Data["backup_path"].(string)
	if !ok {
		return "", fmt.Errorf("invalid backup path in response")
	}

	return backupPath, nil
}

// RestoreHosts 恢复hosts文件
func (c *XPCClient) RestoreHosts(ctx context.Context, backupPath string) error {
	params := map[string]interface{}{
		"backup_path": backupPath,
	}

	resp, err := c.SendRequest(ctx, "restore_hosts", params)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("restore hosts failed: %s", resp.Error)
	}

	return nil
}

// ValidateHosts 验证hosts文件
func (c *XPCClient) ValidateHosts(ctx context.Context) error {
	resp, err := c.SendRequest(ctx, "validate_hosts", nil)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("validate hosts failed: %s", resp.Error)
	}

	return nil
}

// GetStatus 获取Helper Tool状态
func (c *XPCClient) GetStatus(ctx context.Context) (map[string]interface{}, error) {
	resp, err := c.SendRequest(ctx, "get_status", nil)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("get status failed: %s", resp.Error)
	}

	return resp.Data, nil
}

// SetTimeout 设置请求超时时间
func (c *XPCClient) SetTimeout(timeout time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.timeout = timeout
}

// GetTimeout 获取请求超时时间
func (c *XPCClient) GetTimeout() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.timeout
}

// generateClientID 生成客户端ID
func (c *XPCClient) generateClientID() string {
	return fmt.Sprintf("client_%d", time.Now().UnixNano())
}

// sendXPCMessage 发送XPC消息（模拟实现）
func (c *XPCClient) sendXPCMessage(ctx context.Context, reqData []byte) ([]byte, error) {
	// 在实际实现中，这里会使用macOS的XPC API发送消息
	// 目前使用模拟实现

	// 模拟网络延迟
	select {
	case <-time.After(100 * time.Millisecond):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// 模拟成功响应
	resp := &XPCResponse{
		Success:   true,
		Data:      map[string]interface{}{"status": "ok"},
		Timestamp: time.Now(),
	}

	return json.Marshal(resp)
}

// XPCClientPool XPC客户端池，用于管理多个连接
type XPCClientPool struct {
	serviceName string
	logger      Logger
	clients     []*XPCClient
	maxClients  int
	currentIdx  int
	mu          sync.RWMutex
}

// NewXPCClientPool 创建XPC客户端池
func NewXPCClientPool(serviceName string, logger Logger, maxClients int) *XPCClientPool {
	if maxClients <= 0 {
		maxClients = 5
	}

	return &XPCClientPool{
		serviceName: serviceName,
		logger:      logger,
		clients:     make([]*XPCClient, 0, maxClients),
		maxClients:  maxClients,
		currentIdx:  0,
	}
}

// GetClient 获取可用的XPC客户端
func (p *XPCClientPool) GetClient() (*XPCClient, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 如果池中没有客户端，创建新的
	if len(p.clients) == 0 {
		client := NewXPCClient(p.serviceName, p.logger)
		if err := client.Connect(); err != nil {
			return nil, fmt.Errorf("failed to connect XPC client: %w", err)
		}
		p.clients = append(p.clients, client)
		return client, nil
	}

	// 轮询选择客户端
	client := p.clients[p.currentIdx%len(p.clients)]
	p.currentIdx++

	// 检查连接状态
	if !client.IsConnected() {
		if err := client.Connect(); err != nil {
			return nil, fmt.Errorf("failed to reconnect XPC client: %w", err)
		}
	}

	return client, nil
}

// Close 关闭客户端池
func (p *XPCClientPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, client := range p.clients {
		if err := client.Disconnect(); err != nil {
			p.logger.Error("Error disconnecting XPC client", "error", err)
		}
	}

	p.clients = nil
	return nil
}