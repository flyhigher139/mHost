package ui

import (
	"testing"
	"time"

	"github.com/flyhigher139/mhost/pkg/models"
)

// TestValidateIPAddress 测试IP地址验证
func TestValidateIPAddress(t *testing.T) {
	// 创建一个模拟的Manager用于测试验证方法
	manager := &Manager{}

	// 测试用例
	testCases := []struct {
		name     string
		ip       string
		expected bool
	}{
		{"Valid IP", "192.168.1.1", true},
		{"Valid IP 2", "10.0.0.1", true},
		{"Valid IP 3", "127.0.0.1", true},
		{"Empty IP", "", false},
		{"Invalid format 1", "192.168.1", false},
		{"Invalid format 2", "192.168.1.1.1", false},
		{"Invalid range", "256.1.1.1", false},
		{"Invalid characters", "192.168.a.1", false},
		{"Negative number", "192.168.-1.1", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := manager.validateIPAddress(tc.ip)
			if tc.expected && err != nil {
				t.Errorf("Expected valid IP %s, but got error: %v", tc.ip, err)
			}
			if !tc.expected && err == nil {
				t.Errorf("Expected invalid IP %s, but got no error", tc.ip)
			}
		})
	}
}

// TestValidateInput 测试输入验证
func TestValidateInput(t *testing.T) {
	// 创建一个模拟的Manager用于测试验证方法
	manager := &Manager{}

	// 测试用例
	testCases := []struct {
		name      string
		input     string
		fieldName string
		required  bool
		maxLength int
		expected  bool
	}{
		{"Valid input", "test", "field", true, 10, true},
		{"Empty required", "", "field", true, 10, false},
		{"Empty not required", "", "field", false, 10, true},
		{"Too long", "very long text", "field", false, 5, false},
		{"Exact length", "12345", "field", false, 5, true},
		{"Whitespace only required", "   ", "field", true, 10, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := manager.validateInput(tc.input, tc.fieldName, tc.required, tc.maxLength)
			if tc.expected && err != nil {
				t.Errorf("Expected valid input, but got error: %v", err)
			}
			if !tc.expected && err == nil {
				t.Errorf("Expected invalid input, but got no error")
			}
		})
	}
}

// TestValidateHostname 测试主机名验证
func TestValidateHostname(t *testing.T) {
	// 创建一个模拟的Manager用于测试验证方法
	manager := &Manager{}

	// 测试用例
	testCases := []struct {
		name     string
		hostname string
		expected bool
	}{
		{"Valid hostname", "example.com", true},
		{"Valid subdomain", "www.example.com", true},
		{"Valid localhost", "localhost", true},
		{"Empty hostname", "", false},
		{"Whitespace only", "   ", false},
		{"Too long hostname", string(make([]byte, 300)), false},
		{"Invalid characters", "test@example.com", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := manager.validateHostname(tc.hostname)
			if tc.expected && err != nil {
				t.Errorf("Expected valid hostname %s, but got error: %v", tc.hostname, err)
			}
			if !tc.expected && err == nil {
				t.Errorf("Expected invalid hostname %s, but got no error", tc.hostname)
			}
		})
	}
}

// TestHandlePanic 测试panic处理
func TestHandlePanic(t *testing.T) {
	// 跳过这个测试，因为handlePanic方法需要完整的UI环境
	// 在实际应用中，panic处理会显示错误对话框
	t.Skip("Skipping panic test as it requires full UI environment")
}

// MockProfile 创建测试用的Profile
func createMockProfile() *models.Profile {
	profile := &models.Profile{
		ID:          "test-profile",
		Name:        "Test Profile",
		Description: "Test Description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Entries:     []*models.HostEntry{},
	}

	// 添加测试Host条目
	entry := models.NewHostEntry("192.168.1.100", "test.local", "Test entry")
	profile.AddEntry(entry)

	return profile
}

// TestProfileOperations 测试Profile操作
func TestProfileOperations(t *testing.T) {
	// 创建一个模拟的Manager
	manager := &Manager{}

	// 创建测试Profile
	mockProfile := createMockProfile()

	// 测试设置当前Profile
	manager.currentProfile = mockProfile
	if manager.currentProfile == nil {
		t.Error("Current profile should not be nil")
	}

	if manager.currentProfile.Name != "Test Profile" {
		t.Errorf("Expected profile name 'Test Profile', got '%s'", manager.currentProfile.Name)
	}

	// 测试Host条目
	if len(manager.currentProfile.Entries) != 1 {
		t.Errorf("Expected 1 host entry, got %d", len(manager.currentProfile.Entries))
	}

	entry := manager.currentProfile.Entries[0]
	if entry.Hostname != "test.local" {
		t.Errorf("Expected hostname 'test.local', got '%s'", entry.Hostname)
	}

	if entry.IP != "192.168.1.100" {
		t.Errorf("Expected IP '192.168.1.100', got '%s'", entry.IP)
	}
}

// BenchmarkValidateIPAddress 性能测试IP地址验证
func BenchmarkValidateIPAddress(b *testing.B) {
	// 创建一个模拟的Manager
	manager := &Manager{}
	testIP := "192.168.1.1"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.validateIPAddress(testIP)
	}
}

// BenchmarkValidateInput 性能测试输入验证
func BenchmarkValidateInput(b *testing.B) {
	// 创建一个模拟的Manager
	manager := &Manager{}
	testInput := "test input"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.validateInput(testInput, "field", true, 100)
	}
}

// BenchmarkValidateHostname 性能测试主机名验证
func BenchmarkValidateHostname(b *testing.B) {
	// 创建一个模拟的Manager
	manager := &Manager{}
	testHostname := "www.example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.validateHostname(testHostname)
	}
}