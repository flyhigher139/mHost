package ui

import (
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/flyhigher139/mhost/internal/config"
	"github.com/flyhigher139/mhost/internal/host"
	"github.com/flyhigher139/mhost/internal/profile"
	"github.com/flyhigher139/mhost/pkg/models"
)

// Manager UI管理器
type Manager struct {
	window         fyne.Window
	configManager  config.Manager
	profileManager profile.Manager
	hostManager    host.Manager

	// UI组件
	mainContainer *fyne.Container
	toolbar       *fyne.Container
	profileList   *widget.List
	hostEntryList *widget.List
	statusBar     *widget.Label
	menuBar       *fyne.MainMenu

	// 当前状态
	currentProfile *models.Profile
	appConfig      *models.AppConfig
	profiles       []*models.Profile
	hostEntries    []*models.HostEntry
}

// NewManager 创建新的UI管理器
func NewManager(window fyne.Window) (*Manager, error) {
	// 初始化管理器
	configManager := config.NewManager("", "")

	// 获取用户主目录作为数据目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}
	dataDir := filepath.Join(homeDir, ".mhost")

	profileManager, err := profile.NewManager(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create profile manager: %w", err)
	}
	hostManager := host.NewManager("", "")

	// 加载配置
	appConfig, err := configManager.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// 创建UI管理器
	manager := &Manager{
		window:         window,
		configManager:  configManager,
		profileManager: profileManager,
		hostManager:    hostManager,
		appConfig:      appConfig,
	}

	// 初始化UI组件
	if err := manager.initializeUI(); err != nil {
		return nil, fmt.Errorf("failed to initialize UI: %w", err)
	}

	// 加载初始数据
	if err := manager.loadInitialData(); err != nil {
		return nil, fmt.Errorf("failed to load initial data: %w", err)
	}

	return manager, nil
}

// initializeUI 初始化UI组件
func (m *Manager) initializeUI() error {
	// 创建菜单栏
	m.createMenuBar()

	// 创建工具栏
	m.createToolbar()

	// 创建Profile列表
	m.createProfileList()

	// 创建Host条目列表
	m.createHostEntryList()

	// 创建状态栏
	m.statusBar = widget.NewLabel("就绪")

	// 创建主容器
	m.createMainContainer()

	return nil
}

// createMenuBar 创建菜单栏
func (m *Manager) createMenuBar() {
	// 文件菜单
	fileMenu := fyne.NewMenu("文件",
		fyne.NewMenuItem("新建Profile", m.onNewProfile),
		fyne.NewMenuItem("导入Profile", m.onImportProfile),
		fyne.NewMenuItem("导出Profile", m.onExportProfile),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("备份Hosts文件", m.onBackupHosts),
		fyne.NewMenuItem("恢复Hosts文件", m.onRestoreHosts),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("退出", func() { m.window.Close() }),
	)

	// 编辑菜单
	editMenu := fyne.NewMenu("编辑",
		fyne.NewMenuItem("添加Host条目", m.onAddHostEntry),
		fyne.NewMenuItem("编辑Host条目", m.onEditHostEntry),
		fyne.NewMenuItem("删除Host条目", m.onDeleteHostEntry),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("应用Profile", m.onApplyProfile),
	)

	// 工具菜单
	toolsMenu := fyne.NewMenu("工具",
		fyne.NewMenuItem("验证Hosts文件", m.onValidateHosts),
		fyne.NewMenuItem("清理无效条目", m.onCleanupHosts),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("设置", m.onShowSettings),
	)

	// 帮助菜单
	helpMenu := fyne.NewMenu("帮助",
		fyne.NewMenuItem("关于", m.onShowAbout),
		fyne.NewMenuItem("用户手册", m.onShowHelp),
	)

	m.menuBar = fyne.NewMainMenu(fileMenu, editMenu, toolsMenu, helpMenu)
	m.window.SetMainMenu(m.menuBar)
}

// createToolbar 创建工具栏
func (m *Manager) createToolbar() {
	// 简化工具栏，暂时不使用图标
	m.toolbar = container.NewHBox(
		widget.NewButton("添加", m.onAddHostEntry),
		widget.NewButton("编辑", m.onEditHostEntry),
		widget.NewButton("删除", m.onDeleteHostEntry),
		widget.NewSeparator(),
		widget.NewButton("应用", m.onApplyProfile),
		widget.NewButton("备份", m.onBackupHosts),
	)
}

// createMainContainer 创建主容器
func (m *Manager) createMainContainer() {
	// 左侧面板：Profile列表
	leftPanel := container.NewBorder(
		widget.NewLabel("Profiles"),
		nil, nil, nil,
		m.profileList,
	)

	// 右侧面板：Host条目列表
	rightPanel := container.NewBorder(
		widget.NewLabel("Host条目"),
		nil, nil, nil,
		m.hostEntryList,
	)

	// 主内容区域
	mainContent := container.NewHSplit(leftPanel, rightPanel)
	mainContent.SetOffset(0.3) // 左侧占30%，右侧占70%

	// 创建主容器
	m.mainContainer = container.NewBorder(
		m.toolbar,   // 顶部：工具栏
		m.statusBar, // 底部：状态栏
		nil, nil,    // 左右：无
		mainContent, // 中心：主内容
	)
}

// loadInitialData 加载初始数据
func (m *Manager) loadInitialData() error {
	// 加载Profile列表
	profileSummaries, err := m.profileManager.ListProfiles()
	if err != nil {
		return fmt.Errorf("failed to list profiles: %w", err)
	}

	// 转换为完整的Profile对象
	m.profiles = make([]*models.Profile, 0, len(profileSummaries))
	for _, summary := range profileSummaries {
		profile, err := m.profileManager.GetProfile(summary.ID)
		if err != nil {
			continue // 跳过无法加载的profile
		}
		m.profiles = append(m.profiles, profile)
	}
	m.profileList.Refresh()

	// 获取活动Profile
	activeProfile, err := m.profileManager.GetActiveProfile()
	if err == nil && activeProfile != nil {
		m.currentProfile = activeProfile
		m.hostEntries = activeProfile.Entries
		m.hostEntryList.Refresh()
	}

	// 更新状态栏
	m.updateStatusBar()

	return nil
}

// GetMainContainer 获取主容器
func (m *Manager) GetMainContainer() fyne.CanvasObject {
	return m.mainContainer
}

// OnWindowClose 窗口关闭回调
func (m *Manager) OnWindowClose() {
	// 保存当前配置
	if m.appConfig != nil {
		// 保存窗口大小和位置
		size := m.window.Content().Size()
		m.appConfig.Window.Width = int(size.Width)
		m.appConfig.Window.Height = int(size.Height)

		// 保存配置
		if err := m.configManager.SaveConfig(m.appConfig); err != nil {
			fmt.Printf("Failed to save config: %v\n", err)
		}
	}

	// 停止配置监听
	m.configManager.StopWatching()
}

// updateStatusBar 更新状态栏
func (m *Manager) updateStatusBar() {
	var message string
	if m.currentProfile != nil {
		message = fmt.Sprintf("当前Profile: %s (%d个条目)",
			m.currentProfile.Name, len(m.currentProfile.Entries))
	} else {
		message = "未选择Profile"
	}
	m.statusBar.SetText(message)
}

// 事件处理方法

// createProfileList 创建Profile列表
func (m *Manager) createProfileList() {
	m.profileList = widget.NewList(
		func() int {
			return len(m.profiles)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= 0 && id < len(m.profiles) {
				obj.(*widget.Label).SetText(m.profiles[id].Name)
			}
		},
	)
	m.profileList.OnSelected = m.onProfileSelected
}

// createHostEntryList 创建Host条目列表
func (m *Manager) createHostEntryList() {
	m.hostEntryList = widget.NewList(
		func() int {
			return len(m.hostEntries)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= 0 && id < len(m.hostEntries) {
				entry := m.hostEntries[id]
				obj.(*widget.Label).SetText(fmt.Sprintf("%s -> %s", entry.IP, entry.Hostname))
			}
		},
	)
}

// onProfileSelected Profile选择事件
func (m *Manager) onProfileSelected(id widget.ListItemID) {
	if id >= 0 && id < len(m.profiles) {
		m.currentProfile = m.profiles[id]
		m.hostEntries = m.currentProfile.Entries
		m.hostEntryList.Refresh()
		m.updateStatusBar()
	}
}

// onHostEntryChanged Host条目变化事件
func (m *Manager) onHostEntryChanged() {
	m.updateStatusBar()
	// 可以在这里添加自动保存逻辑
}

// onAddHostEntry 添加Host条目事件
func (m *Manager) onAddHostEntry() {
	// TODO: 实现添加Host条目对话框
	m.statusBar.SetText("添加Host条目功能待实现")
}

// onEditHostEntry 编辑Host条目事件
func (m *Manager) onEditHostEntry() {
	// TODO: 实现编辑Host条目对话框
	m.statusBar.SetText("编辑Host条目功能待实现")
}

// onDeleteHostEntry 删除Host条目事件
func (m *Manager) onDeleteHostEntry() {
	// TODO: 实现删除Host条目确认对话框
	m.statusBar.SetText("删除Host条目功能待实现")
}

// onApplyProfile 应用Profile事件
func (m *Manager) onApplyProfile() {
	if m.currentProfile == nil {
		m.statusBar.SetText("请先选择一个Profile")
		return
	}

	err := m.hostManager.ApplyProfile(m.currentProfile)
	if err != nil {
		m.statusBar.SetText(fmt.Sprintf("应用Profile失败: %v", err))
	} else {
		m.statusBar.SetText(fmt.Sprintf("已应用Profile: %s", m.currentProfile.Name))
	}
}

// onBackupHosts 备份Hosts事件
func (m *Manager) onBackupHosts() {
	backup, err := m.hostManager.BackupHostsFile()
	if err != nil {
		m.statusBar.SetText(fmt.Sprintf("备份失败: %v", err))
	} else {
		m.statusBar.SetText(fmt.Sprintf("Hosts文件备份成功: %s", backup.FilePath))
	}
}

// onShowSettings 显示设置事件
func (m *Manager) onShowSettings() {
	// TODO: 实现设置对话框
	m.statusBar.SetText("设置功能待实现")
}

// 菜单和工具栏事件处理方法（占位符）
func (m *Manager) onNewProfile()    { /* TODO: 实现新建Profile */ }
func (m *Manager) onImportProfile() { /* TODO: 实现导入Profile */ }
func (m *Manager) onExportProfile() { /* TODO: 实现导出Profile */ }
func (m *Manager) onRestoreHosts()  { /* TODO: 实现恢复Hosts */ }
func (m *Manager) onValidateHosts() { /* TODO: 实现验证Hosts */ }
func (m *Manager) onCleanupHosts()  { /* TODO: 实现清理Hosts */ }
func (m *Manager) onShowAbout()     { /* TODO: 实现显示关于 */ }
func (m *Manager) onShowHelp()      { /* TODO: 实现显示帮助 */ }
