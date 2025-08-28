package models

import (
	"time"
)

// EventType 事件类型
type EventType string

const (
	// Profile相关事件
	EventProfileCreated   EventType = "profile.created"
	EventProfileUpdated   EventType = "profile.updated"
	EventProfileDeleted   EventType = "profile.deleted"
	EventProfileActivated EventType = "profile.activated"

	// Host条目相关事件
	EventHostEntryAdded   EventType = "host_entry.added"
	EventHostEntryUpdated EventType = "host_entry.updated"
	EventHostEntryDeleted EventType = "host_entry.deleted"
	EventHostEntryToggled EventType = "host_entry.toggled"

	// 系统相关事件
	EventSystemHostsUpdated   EventType = "system.hosts_updated"
	EventSystemBackupCreated  EventType = "system.backup_created"
	EventSystemBackupRestored EventType = "system.backup_restored"
	EventSystemConfigChanged  EventType = "system.config_changed"

	// 错误事件
	EventError   EventType = "error"
	EventWarning EventType = "warning"
)

// Event 事件结构
type Event struct {
	ID        string                 `json:"id"`         // 事件ID
	Type      EventType              `json:"type"`       // 事件类型
	Timestamp time.Time              `json:"timestamp"`  // 事件时间戳
	Source    string                 `json:"source"`     // 事件源
	Data      map[string]interface{} `json:"data"`       // 事件数据
	UserID    string                 `json:"user_id"`    // 用户ID(可选)
	SessionID string                 `json:"session_id"` // 会话ID(可选)
}

// EventHandler 事件处理器函数类型
type EventHandler func(event Event) error

// EventSubscription 事件订阅信息
type EventSubscription struct {
	ID        string       `json:"id"`         // 订阅ID
	EventType EventType    `json:"event_type"` // 订阅的事件类型
	Handler   EventHandler `json:"-"`          // 处理器函数(不序列化)
	CreatedAt time.Time    `json:"created_at"` // 创建时间
	Active    bool         `json:"active"`     // 是否激活
}

// NewEvent 创建新事件
func NewEvent(eventType EventType, source string, data map[string]interface{}) *Event {
	return &Event{
		ID:        generateID(),
		Type:      eventType,
		Timestamp: time.Now(),
		Source:    source,
		Data:      data,
	}
}

// NewEventWithUser 创建带用户信息的新事件
func NewEventWithUser(eventType EventType, source string, data map[string]interface{}, userID, sessionID string) *Event {
	event := NewEvent(eventType, source, data)
	event.UserID = userID
	event.SessionID = sessionID
	return event
}

// AddData 向事件添加数据
func (e *Event) AddData(key string, value interface{}) {
	if e.Data == nil {
		e.Data = make(map[string]interface{})
	}
	e.Data[key] = value
}

// GetData 从事件获取数据
func (e *Event) GetData(key string) (interface{}, bool) {
	if e.Data == nil {
		return nil, false
	}
	value, exists := e.Data[key]
	return value, exists
}

// GetStringData 从事件获取字符串数据
func (e *Event) GetStringData(key string) (string, bool) {
	value, exists := e.GetData(key)
	if !exists {
		return "", false
	}
	str, ok := value.(string)
	return str, ok
}

// GetIntData 从事件获取整数数据
func (e *Event) GetIntData(key string) (int, bool) {
	value, exists := e.GetData(key)
	if !exists {
		return 0, false
	}
	num, ok := value.(int)
	return num, ok
}

// Clone 创建事件的深拷贝
func (e *Event) Clone() *Event {
	cloned := *e

	// 深拷贝数据map
	cloned.Data = make(map[string]interface{})
	for k, v := range e.Data {
		cloned.Data[k] = v
	}

	return &cloned
}

// IsProfileEvent 检查是否为Profile相关事件
func (e *Event) IsProfileEvent() bool {
	return e.Type == EventProfileCreated ||
		e.Type == EventProfileUpdated ||
		e.Type == EventProfileDeleted ||
		e.Type == EventProfileActivated
}

// IsHostEntryEvent 检查是否为Host条目相关事件
func (e *Event) IsHostEntryEvent() bool {
	return e.Type == EventHostEntryAdded ||
		e.Type == EventHostEntryUpdated ||
		e.Type == EventHostEntryDeleted ||
		e.Type == EventHostEntryToggled
}

// IsSystemEvent 检查是否为系统相关事件
func (e *Event) IsSystemEvent() bool {
	return e.Type == EventSystemHostsUpdated ||
		e.Type == EventSystemBackupCreated ||
		e.Type == EventSystemBackupRestored ||
		e.Type == EventSystemConfigChanged
}

// IsErrorEvent 检查是否为错误事件
func (e *Event) IsErrorEvent() bool {
	return e.Type == EventError || e.Type == EventWarning
}

// NewSubscription 创建新的事件订阅
func NewSubscription(eventType EventType, handler EventHandler) *EventSubscription {
	return &EventSubscription{
		ID:        generateID(),
		EventType: eventType,
		Handler:   handler,
		CreatedAt: time.Now(),
		Active:    true,
	}
}

// Activate 激活订阅
func (s *EventSubscription) Activate() {
	s.Active = true
}

// Deactivate 停用订阅
func (s *EventSubscription) Deactivate() {
	s.Active = false
}
