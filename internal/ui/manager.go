package ui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
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
	mainContainer   *fyne.Container
	toolbar         *fyne.Container
	profileList     *widget.List
	hostEntryList   *widget.List
	statusBar       *widget.Label
	menuBar         *fyne.MainMenu
	profileSelector *widget.Select

	// 当前状态
	currentProfile   *models.Profile
	currentHostEntry *models.HostEntry
	appConfig        *models.AppConfig
	profiles         []*models.Profile
	hostEntries      []*models.HostEntry
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
		fyne.NewMenuItem("刷新", m.onRefresh),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("退出", func() { m.window.Close() }),
	)

	// 编辑菜单
	editMenu := fyne.NewMenu("编辑",
		fyne.NewMenuItem("编辑Profile", m.onEditProfile),
		fyne.NewMenuItem("删除Profile", m.onDeleteProfile),
		fyne.NewMenuItem("复制Profile", m.onCopyProfile),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("添加Host条目", m.onAddHostEntry),
		fyne.NewMenuItem("编辑Host条目", m.onEditHostEntry),
		fyne.NewMenuItem("删除Host条目", m.onDeleteHostEntry),
		fyne.NewMenuItem("启用/禁用Host条目", m.onToggleHostEntry),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("应用Profile", m.onApplyProfile),
	)

	// 工具菜单
	toolsMenu := fyne.NewMenu("工具",
		fyne.NewMenuItem("验证Hosts文件", m.onValidateHosts),
		fyne.NewMenuItem("清理无效条目", m.onCleanupHosts),
		fyne.NewMenuItem("清理备份文件", m.onCleanupBackups),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("设置", m.onShowSettings),
	)

	// 视图菜单
	viewMenu := fyne.NewMenu("视图",
		fyne.NewMenuItem("快速切换Profile", m.showQuickSwitchDialog),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("显示所有Profile", func() {
			m.onFilterProfiles("")
		}),
		fyne.NewMenuItem("显示激活的Profile", func() {
			m.onFilterProfiles("active")
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("展开所有", m.onExpandAll),
		fyne.NewMenuItem("折叠所有", m.onCollapseAll),
	)

	// 帮助菜单
	helpMenu := fyne.NewMenu("帮助",
		fyne.NewMenuItem("用户手册", m.onShowHelp),
		fyne.NewMenuItem("快捷键", m.onShowShortcuts),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("检查更新", m.onCheckUpdates),
		fyne.NewMenuItem("关于", m.onShowAbout),
	)

	m.menuBar = fyne.NewMainMenu(fileMenu, editMenu, toolsMenu, viewMenu, helpMenu)
	m.window.SetMainMenu(m.menuBar)
}

// createToolbar 创建工具栏
func (m *Manager) createToolbar() {
	// 创建Profile快速切换下拉框
	profileSelect := widget.NewSelect([]string{}, func(selected string) {
		m.onQuickSwitchProfile(selected)
	})
	profileSelect.PlaceHolder = "快速切换Profile"
	
	// 简化工具栏，暂时不使用图标
	m.toolbar = container.NewHBox(
		// Profile快速切换
		widget.NewLabel("快速切换:"),
		profileSelect,
		widget.NewSeparator(),
		// Profile操作
		widget.NewButton("新建Profile", m.onNewProfile),
		widget.NewButton("编辑Profile", m.onEditProfile),
		widget.NewButton("删除Profile", m.onDeleteProfile),
		widget.NewSeparator(),
		// Host条目操作
		widget.NewButton("添加Host", m.onAddHostEntry),
		widget.NewButton("编辑Host", m.onEditHostEntry),
		widget.NewButton("删除Host", m.onDeleteHostEntry),
		widget.NewSeparator(),
		// 应用操作
		widget.NewButton("应用Profile", m.onApplyProfile),
		widget.NewButton("备份Hosts", m.onBackupHosts),
		widget.NewSeparator(),
		// 其他操作
		widget.NewButton("刷新", m.onRefresh),
		widget.NewButton("设置", m.onShowSettings),
	)
	
	// 保存Profile选择器的引用
	m.profileSelector = profileSelect
}

// createMainContainer 创建主容器
func (m *Manager) createMainContainer() {
	// 创建左侧Profile列表标题栏
	profileTitleBar := container.NewHBox(
		widget.NewLabelWithStyle("Profile列表", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		widget.NewButtonWithIcon("", theme.ContentAddIcon(), m.onNewProfile),
		widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), m.onEditProfile),
		widget.NewButtonWithIcon("", theme.DeleteIcon(), m.onDeleteProfile),
	)
	
	// 创建左侧Profile容器
	leftPanel := container.NewBorder(
		profileTitleBar,
		nil, nil, nil,
		m.profileList,
	)

	// 创建右侧Host条目标题栏
	hostTitleBar := container.NewHBox(
		widget.NewLabelWithStyle("Host条目", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		widget.NewButtonWithIcon("", theme.ContentAddIcon(), m.onAddHostEntry),
		widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), m.onEditHostEntry),
		widget.NewButtonWithIcon("", theme.DeleteIcon(), m.onDeleteHostEntry),
	)
	
	// 创建右侧Host条目容器
	rightPanel := container.NewBorder(
		hostTitleBar,
		nil, nil, nil,
		m.hostEntryList,
	)

	// 主内容区域
	mainContent := container.NewHSplit(leftPanel, rightPanel)
	mainContent.SetOffset(0.35) // 左侧占35%，右侧占65%

	// 创建状态栏容器，添加更多信息
	statusContainer := container.NewHBox(
		m.statusBar,
		layout.NewSpacer(),
		widget.NewLabel("mHost v1.0"),
	)

	// 创建主容器
	m.mainContainer = container.NewBorder(
		m.toolbar,       // 顶部：工具栏
		statusContainer, // 底部：状态栏
		nil, nil,        // 左右：无
		mainContent,     // 中心：主内容
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
			// 创建Profile条目的布局
			name := widget.NewLabel("")
			name.TextStyle.Bold = true
			desc := widget.NewLabel("")
			desc.TextStyle.Italic = true
			status := widget.NewLabel("")
			
			// 创建状态指示器
			statusIcon := widget.NewIcon(nil)
			
			// 创建水平布局的状态行
			statusRow := container.NewHBox(
				statusIcon,
				status,
				layout.NewSpacer(),
			)
			
			return container.NewVBox(
				name,
				desc,
				statusRow,
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= 0 && id < len(m.profiles) {
				profile := m.profiles[id]
				vbox := obj.(*fyne.Container)
				
				// 更新名称
				nameLabel := vbox.Objects[0].(*widget.Label)
				nameLabel.SetText(profile.Name)
				
				// 更新描述
				descLabel := vbox.Objects[1].(*widget.Label)
				if profile.Description != "" {
					descLabel.SetText(profile.Description)
				} else {
					descLabel.SetText("无描述")
				}
				
				// 更新状态行
				statusRow := vbox.Objects[2].(*fyne.Container)
				statusIcon := statusRow.Objects[0].(*widget.Icon)
				statusLabel := statusRow.Objects[1].(*widget.Label)
				
				statusText := fmt.Sprintf("条目数: %d", len(profile.Entries))
				if profile.IsActive {
					statusText += " (当前激活)"
					statusIcon.SetResource(theme.ConfirmIcon())
				} else {
					statusIcon.SetResource(theme.RadioButtonIcon())
				}
				statusLabel.SetText(statusText)
			}
		},
	)
	
	// 设置选择事件处理
	m.profileList.OnSelected = func(id widget.ListItemID) {
		if id >= 0 && id < len(m.profiles) {
			m.onProfileSelected(id)
		}
	}
}

// createHostEntryList 创建Host条目列表
func (m *Manager) createHostEntryList() {
	m.hostEntryList = widget.NewList(
		func() int {
			return len(m.hostEntries)
		},
		func() fyne.CanvasObject {
			// 创建Host条目的布局
			hostname := widget.NewLabel("")
			hostname.TextStyle.Bold = true
			ip := widget.NewLabel("")
			comment := widget.NewLabel("")
			comment.TextStyle.Italic = true
			status := widget.NewLabel("")
			
			// 创建启用/禁用的复选框
			enabled := widget.NewCheck("", nil)
			enabled.Disable() // 只读显示
			
			// 创建状态指示器
			statusIcon := widget.NewIcon(nil)
			
			// 创建主机名行（包含启用状态和图标）
			hostnameRow := container.NewHBox(
				enabled,
				statusIcon,
				hostname,
				layout.NewSpacer(),
			)
			
			// 创建IP地址行（带图标）
			ipIcon := widget.NewIcon(theme.ComputerIcon())
			ipRow := container.NewHBox(
				ipIcon,
				ip,
			)
			
			// 创建注释行（带图标）
			commentIcon := widget.NewIcon(theme.DocumentIcon())
			commentRow := container.NewHBox(
				commentIcon,
				comment,
			)
			
			return container.NewVBox(
				hostnameRow,
				ipRow,
				commentRow,
				status,
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= 0 && id < len(m.hostEntries) {
				entry := m.hostEntries[id]
				vbox := obj.(*fyne.Container)
				
				// 更新主机名行
				hostnameRow := vbox.Objects[0].(*fyne.Container)
				enabled := hostnameRow.Objects[0].(*widget.Check)
				statusIcon := hostnameRow.Objects[1].(*widget.Icon)
				hostname := hostnameRow.Objects[2].(*widget.Label)
				
				enabled.SetChecked(entry.Enabled)
				hostname.SetText(entry.Hostname)
				
				// 设置状态图标
				if entry.Enabled {
					statusIcon.SetResource(theme.ConfirmIcon())
				} else {
					statusIcon.SetResource(theme.CancelIcon())
				}
				
				// 更新IP地址行
				ipRow := vbox.Objects[1].(*fyne.Container)
				ip := ipRow.Objects[1].(*widget.Label)
				ip.SetText(entry.IP)
				
				// 更新注释行
				commentRow := vbox.Objects[2].(*fyne.Container)
				comment := commentRow.Objects[1].(*widget.Label)
				if entry.Comment != "" {
					comment.SetText(entry.Comment)
				} else {
					comment.SetText("无注释")
				}
				
				// 更新状态
				status := vbox.Objects[3].(*widget.Label)
				statusText := fmt.Sprintf("创建时间: %s", entry.CreatedAt.Format("2006-01-02 15:04"))
				if !entry.Enabled {
					statusText += " (已禁用)"
				}
				status.SetText(statusText)
			}
		},
	)
	
	// 设置双击编辑事件
	m.hostEntryList.OnSelected = func(id widget.ListItemID) {
		if id >= 0 && id < len(m.hostEntries) {
			m.currentHostEntry = m.hostEntries[id]
			m.statusBar.SetText(fmt.Sprintf("已选择Host条目: %s -> %s", m.hostEntries[id].Hostname, m.hostEntries[id].IP))
		}
	}
}

// onProfileSelected Profile选择事件
func (m *Manager) onProfileSelected(id widget.ListItemID) {
	if id >= 0 && id < len(m.profiles) {
		// 设置当前选中的Profile
		m.currentProfile = m.profiles[id]
		
		// 加载Profile的Host条目
		m.hostEntries = m.currentProfile.Entries
		
		// 刷新Host条目列表
		m.hostEntryList.Refresh()
		
		// 更新状态栏
		m.statusBar.SetText(fmt.Sprintf("已选择Profile: %s (包含 %d 个Host条目)", m.currentProfile.Name, len(m.currentProfile.Entries)))
	}
}

// onHostEntryChanged Host条目变化事件
func (m *Manager) onHostEntryChanged() {
	m.updateStatusBar()
	// 可以在这里添加自动保存逻辑
}

// onAddHostEntry 添加Host条目事件处理
func (m *Manager) onAddHostEntry() {
	if m.currentProfile == nil {
		dialog.ShowInformation("提示", "请先选择一个Profile", m.window)
		return
	}
	
	// 显示Host条目编辑对话框
	m.showHostEntryDialog(nil)
}

// onEditHostEntry 编辑Host条目事件处理
func (m *Manager) onEditHostEntry() {
	if m.currentHostEntry == nil {
		dialog.ShowInformation("提示", "请先选择要编辑的Host条目", m.window)
		return
	}
	
	// 显示Host条目编辑对话框
	m.showHostEntryDialog(m.currentHostEntry)
}

// onDeleteHostEntry 删除Host条目事件处理
func (m *Manager) onDeleteHostEntry() {
	if m.currentHostEntry == nil {
		dialog.ShowInformation("提示", "请先选择要删除的Host条目", m.window)
		return
	}
	
	// 显示确认删除对话框
	message := fmt.Sprintf("确定要删除Host条目 '%s -> %s' 吗？\n\n此操作不可撤销。", m.currentHostEntry.Hostname, m.currentHostEntry.IP)
	dialog.ShowConfirm("确认删除", message, func(confirmed bool) {
		if !confirmed {
			return
		}
		
		// 从当前Profile中删除Host条目
		if m.currentProfile != nil {
			m.currentProfile.RemoveEntry(m.currentHostEntry.ID)
			
			// 更新Profile
			err := m.profileManager.UpdateProfile(m.currentProfile)
			if err != nil {
				dialog.ShowError(err, m.window)
				return
			}
			
			// 刷新Host条目列表
			m.hostEntries = m.currentProfile.Entries
			m.hostEntryList.Refresh()
			m.currentHostEntry = nil
			
			m.statusBar.SetText("Host条目删除成功")
		}
	}, m.window)
}

// onApplyProfile 应用Profile事件处理
func (m *Manager) onApplyProfile() {
	if m.currentProfile == nil {
		dialog.ShowInformation("提示", "请先选择要应用的Profile", m.window)
		return
	}
	
	// 显示确认对话框
	message := fmt.Sprintf("确定要应用Profile '%s' 吗？\n\n这将会：\n1. 备份当前hosts文件\n2. 将Profile中的%d个Host条目写入hosts文件\n3. 设置此Profile为当前激活状态", 
		m.currentProfile.Name, len(m.currentProfile.Entries))
	
	dialog.ShowConfirm("确认应用Profile", message, func(confirmed bool) {
		if !confirmed {
			return
		}
		
		// 显示进度对话框
		progressDialog := dialog.NewProgressInfinite("应用Profile", "正在应用Profile，请稍候...", m.window)
		progressDialog.Show()
		
		// 在goroutine中执行应用操作
		go func() {
			defer progressDialog.Hide()
			
			// 应用Profile
			err := m.hostManager.ApplyProfile(m.currentProfile)
			if err != nil {
				dialog.ShowError(fmt.Errorf("应用Profile失败: %v", err), m.window)
				return
			}
			
			// 更新Profile状态
			// 先将所有Profile设为非激活状态
			for _, profile := range m.profiles {
				profile.IsActive = false
				m.profileManager.UpdateProfile(profile)
			}
			
			// 设置当前Profile为激活状态
			m.currentProfile.IsActive = true
			err = m.profileManager.UpdateProfile(m.currentProfile)
			if err != nil {
				dialog.ShowError(fmt.Errorf("更新Profile状态失败: %v", err), m.window)
				return
			}
			
			// 刷新界面
			m.refreshProfileList()
			m.statusBar.SetText(fmt.Sprintf("Profile '%s' 应用成功", m.currentProfile.Name))
			
			// 显示成功提示
			dialog.ShowInformation("成功", fmt.Sprintf("Profile '%s' 已成功应用到hosts文件", m.currentProfile.Name), m.window)
		}()
	}, m.window)
}

// onBackupHosts 备份hosts文件事件处理
func (m *Manager) onBackupHosts() {
	// 显示确认对话框
	message := "确定要备份当前hosts文件吗？\n\n备份文件将保存到应用数据目录中。"
	dialog.ShowConfirm("确认备份", message, func(confirmed bool) {
		if !confirmed {
			return
		}
		
		// 显示进度对话框
		progressDialog := dialog.NewProgressInfinite("备份hosts文件", "正在备份hosts文件，请稍候...", m.window)
		progressDialog.Show()
		
		// 在goroutine中执行备份操作
		go func() {
			defer progressDialog.Hide()
			
			// 执行备份
			backup, err := m.hostManager.BackupHostsFile()
			if err != nil {
				dialog.ShowError(fmt.Errorf("备份失败: %v", err), m.window)
				return
			}
			
			m.statusBar.SetText("hosts文件备份成功")
			
			// 显示成功提示
			message := fmt.Sprintf("hosts文件备份成功！\n\n备份文件路径：\n%s", backup.FilePath)
			dialog.ShowInformation("备份成功", message, m.window)
		}()
	}, m.window)
}

// onShowSettings 显示设置事件处理
func (m *Manager) onShowSettings() {
	// 添加panic恢复
	defer m.handlePanic()
	
	// 创建设置表单组件
	hostsPathEntry := widget.NewEntry()
	hostsPathEntry.SetText("/etc/hosts") // 默认hosts文件路径
	hostsPathEntry.Disable() // 只读显示
	
	backupDirEntry := widget.NewEntry()
	backupDirEntry.SetText(m.appConfig.Backup.BackupPath)
	
	// 添加浏览按钮用于选择备份目录
	backupDirButton := widget.NewButton("浏览...", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				m.showErrorDialog("选择目录失败", err)
				return
			}
			if uri != nil {
				backupDirEntry.SetText(uri.Path())
			}
		}, m.window)
	})
	
	backupDirContainer := container.NewBorder(nil, nil, nil, backupDirButton, backupDirEntry)
	
	autoBackupCheck := widget.NewCheck("启用自动备份", nil)
	autoBackupCheck.SetChecked(m.appConfig.Backup.Enabled)
	
	compressionCheck := widget.NewCheck("启用备份压缩", nil)
	compressionCheck.SetChecked(true) // 默认启用压缩
	
	retentionEntry := widget.NewEntry()
	retentionEntry.SetText(fmt.Sprintf("%d", m.appConfig.Backup.RetentionDays))
	
	maxBackupsEntry := widget.NewEntry()
	maxBackupsEntry.SetText(fmt.Sprintf("%d", m.appConfig.Backup.MaxBackups))
	
	backupIntervalSelect := widget.NewSelect([]string{"每小时", "每天", "每周", "手动"}, nil)
	backupIntervalSelect.SetSelected("手动") // 默认手动备份
	
	themeSelect := widget.NewSelect([]string{"light", "dark", "auto"}, nil)
	themeSelect.SetSelected(m.appConfig.UI.Theme)
	
	languageSelect := widget.NewSelect([]string{"zh-CN", "en"}, nil)
	languageSelect.SetSelected(m.appConfig.UI.Language)
	
	// 日志级别设置
	logLevelSelect := widget.NewSelect([]string{"DEBUG", "INFO", "WARN", "ERROR"}, nil)
	logLevelSelect.SetSelected("INFO") // 默认INFO级别
	
	// 安全设置
	requireAdminCheck := widget.NewCheck("需要管理员权限", nil)
	requireAdminCheck.SetChecked(true) // 默认需要管理员权限
	
	backupOnApplyCheck := widget.NewCheck("应用Profile前自动备份", nil)
	backupOnApplyCheck.SetChecked(true) // 默认启用
	
	// 创建分组容器
	backupForm := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "备份目录", Widget: backupDirContainer},
			{Text: "自动备份", Widget: autoBackupCheck},
			{Text: "备份间隔", Widget: backupIntervalSelect},
			{Text: "备份压缩", Widget: compressionCheck},
			{Text: "保留天数", Widget: retentionEntry, HintText: "1-365天"},
			{Text: "最大备份数", Widget: maxBackupsEntry, HintText: "1-100个"},
		},
	}
	backupGroup := widget.NewCard("备份设置", "", backupForm)
	
	uiForm := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "主题", Widget: themeSelect},
			{Text: "语言", Widget: languageSelect},
		},
	}
	uiGroup := widget.NewCard("界面设置", "", uiForm)
	
	systemForm := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Hosts文件路径", Widget: hostsPathEntry},
			{Text: "日志级别", Widget: logLevelSelect},
		},
	}
	systemGroup := widget.NewCard("系统设置", "", systemForm)
	
	securityForm := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "管理员权限", Widget: requireAdminCheck},
			{Text: "自动备份", Widget: backupOnApplyCheck},
		},
	}
	securityGroup := widget.NewCard("安全设置", "", securityForm)
	
	// 创建滚动容器
	content := container.NewVBox(
		systemGroup,
		backupGroup,
		uiGroup,
		securityGroup,
	)
	
	scroll := container.NewScroll(content)
	scroll.SetMinSize(fyne.NewSize(500, 400))
	
	// 创建设置对话框
	d := dialog.NewCustomConfirm("应用设置", "保存", "取消", scroll, func(confirmed bool) {
		if !confirmed {
			return
		}
		
		// 验证输入
		retentionDays, err := fmt.Sscanf(retentionEntry.Text, "%d", new(int))
		if err != nil || retentionDays != 1 {
			m.showErrorDialog("输入验证错误", errors.New("备份保留天数必须是有效数字"))
			return
		}
		
		maxBackups, err := fmt.Sscanf(maxBackupsEntry.Text, "%d", new(int))
		if err != nil || maxBackups != 1 {
			m.showErrorDialog("输入验证错误", errors.New("最大备份数量必须是有效数字"))
			return
		}
		
		// 更新配置
		if backupDirEntry.Text != "" {
			m.appConfig.Backup.BackupPath = backupDirEntry.Text
		}
		m.appConfig.Backup.Enabled = autoBackupCheck.Checked
		fmt.Sscanf(retentionEntry.Text, "%d", &m.appConfig.Backup.RetentionDays)
		fmt.Sscanf(maxBackupsEntry.Text, "%d", &m.appConfig.Backup.MaxBackups)
		m.appConfig.UI.Theme = themeSelect.Selected
		m.appConfig.UI.Language = languageSelect.Selected
		
		// 保存配置到文件
		err = m.configManager.SaveConfig(m.appConfig)
		if err != nil {
			m.showErrorDialog("保存失败", err)
			return
		}
		
		m.showSuccessDialog("成功", "设置保存成功，部分设置需要重启应用后生效")
	}, m.window)
	
	// 设置对话框大小并显示
	d.Resize(fyne.NewSize(600, 500))
	d.Show()
}

// onNewProfile 新建Profile事件
func (m *Manager) onNewProfile() {
	m.showProfileDialog(nil)
}

// showProfileDialog 显示Profile编辑对话框
func (m *Manager) showProfileDialog(profile *models.Profile) {
	// 添加panic恢复
	defer m.handlePanic()
	
	// 创建输入组件
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("请输入Profile名称")
	descEntry := widget.NewMultiLineEntry()
	descEntry.SetPlaceHolder("请输入Profile描述（可选）")
	
	// 如果是编辑模式，填充现有数据
	if profile != nil {
		nameEntry.SetText(profile.Name)
		descEntry.SetText(profile.Description)
	}
	
	// 创建表单
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "名称", Widget: nameEntry, HintText: "Profile的唯一名称"},
			{Text: "描述", Widget: descEntry, HintText: "Profile的详细描述"},
		},
	}
	
	// 设置对话框标题
	title := "新建Profile"
	if profile != nil {
		title = "编辑Profile"
	}
	
	// 创建确认对话框
	d := dialog.NewCustomConfirm(title, "确定", "取消", form, func(confirmed bool) {
		if !confirmed {
			return
		}
		
		name := strings.TrimSpace(nameEntry.Text)
		desc := strings.TrimSpace(descEntry.Text)
		
		// 使用新的验证方法
		if err := m.validateInput(name, "Profile名称", true, 50); err != nil {
			m.showErrorDialog("输入验证错误", err)
			return
		}
		
		if err := m.validateInput(desc, "描述", false, 500); err != nil {
			m.showErrorDialog("输入验证错误", err)
			return
		}
		
		var err error
		if profile == nil {
			// 创建新Profile
			_, err = m.profileManager.CreateProfile(name, desc)
			if err != nil {
				m.showErrorDialog("创建失败", err)
				return
			}
			m.showSuccessDialog("成功", "Profile创建成功")
		} else {
			// 更新现有Profile
			profile.Name = name
			profile.Description = desc
			err = m.profileManager.UpdateProfile(profile)
			if err != nil {
				m.showErrorDialog("更新失败", err)
				return
			}
			m.showSuccessDialog("成功", "Profile更新成功")
		}
		
		// 刷新Profile列表
		m.refreshProfileList()
	}, m.window)
	
	// 设置对话框大小并显示
	d.Resize(fyne.NewSize(400, 300))
	d.Show()
}

// refreshProfileList 刷新Profile列表
func (m *Manager) refreshProfileList() {
	profileSummaries, err := m.profileManager.ListProfiles()
	if err != nil {
		m.statusBar.SetText(fmt.Sprintf("加载Profile列表失败: %v", err))
		return
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
	
	// 更新Profile选择器
	m.updateProfileSelector()
}

// showHostEntryDialog 显示Host条目编辑对话框
func (m *Manager) showHostEntryDialog(hostEntry *models.HostEntry) {
	// 添加panic恢复
	defer m.handlePanic()
	
	// 创建输入组件
	hostnameEntry := widget.NewEntry()
	hostnameEntry.SetPlaceHolder("请输入主机名")
	ipEntry := widget.NewEntry()
	ipEntry.SetPlaceHolder("请输入IP地址")
	commentEntry := widget.NewEntry()
	commentEntry.SetPlaceHolder("请输入注释（可选）")
	enabledCheck := widget.NewCheck("启用此条目", nil)
	enabledCheck.SetChecked(true)
	
	// 如果是编辑模式，填充现有数据
	if hostEntry != nil {
		hostnameEntry.SetText(hostEntry.Hostname)
		ipEntry.SetText(hostEntry.IP)
		commentEntry.SetText(hostEntry.Comment)
		enabledCheck.SetChecked(hostEntry.Enabled)
	}
	
	// 创建表单
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "主机名", Widget: hostnameEntry, HintText: "例如: www.example.com"},
			{Text: "IP地址", Widget: ipEntry, HintText: "例如: 192.168.1.100"},
			{Text: "注释", Widget: commentEntry, HintText: "可选的描述信息"},
			{Text: "状态", Widget: enabledCheck, HintText: "是否启用此Host条目"},
		},
	}
	
	// 设置对话框标题
	title := "添加Host条目"
	if hostEntry != nil {
		title = "编辑Host条目"
	}
	
	// 创建确认对话框
	d := dialog.NewCustomConfirm(title, "确定", "取消", form, func(confirmed bool) {
		if !confirmed {
			return
		}
		
		hostname := strings.TrimSpace(hostnameEntry.Text)
		ip := strings.TrimSpace(ipEntry.Text)
		comment := strings.TrimSpace(commentEntry.Text)
		enabled := enabledCheck.Checked
		
		// 使用新的验证方法
		if err := m.validateHostname(hostname); err != nil {
			m.showErrorDialog("输入验证错误", err)
			return
		}
		
		if err := m.validateIPAddress(ip); err != nil {
			m.showErrorDialog("输入验证错误", err)
			return
		}
		
		if err := m.validateInput(comment, "注释", false, 200); err != nil {
			m.showErrorDialog("输入验证错误", err)
			return
		}
		
		var err error
		if hostEntry == nil {
			// 创建新Host条目
			newEntry := models.NewHostEntry(ip, hostname, comment)
			newEntry.Enabled = enabled
			m.currentProfile.AddEntry(newEntry)
		} else {
			// 更新现有Host条目
			hostEntry.Hostname = hostname
			hostEntry.IP = ip
			hostEntry.Comment = comment
			hostEntry.Enabled = enabled
			hostEntry.UpdatedAt = time.Now()
		}
		
		// 更新Profile
		err = m.profileManager.UpdateProfile(m.currentProfile)
		if err != nil {
			m.showErrorDialog("保存失败", err)
			return
		}
		
		// 刷新Host条目列表
		m.hostEntries = m.currentProfile.Entries
		m.hostEntryList.Refresh()
		
		if hostEntry == nil {
			m.showSuccessDialog("成功", "Host条目添加成功")
		} else {
			m.showSuccessDialog("成功", "Host条目更新成功")
		}
	}, m.window)
	
	// 设置对话框大小并显示
	d.Resize(fyne.NewSize(450, 350))
	d.Show()
}

// onEditProfile 编辑Profile事件处理
func (m *Manager) onEditProfile() {
	if m.currentProfile == nil {
		dialog.ShowInformation("提示", "请先选择要编辑的Profile", m.window)
		return
	}
	
	// 显示编辑对话框
	m.showProfileDialog(m.currentProfile)
}

// onDeleteProfile 删除Profile事件处理
func (m *Manager) onDeleteProfile() {
	// 添加panic恢复
	defer m.handlePanic()
	
	if m.currentProfile == nil {
		dialog.ShowInformation("提示", "请先选择要删除的Profile", m.window)
		return
	}
	
	// 显示确认删除对话框
	message := fmt.Sprintf("确定要删除Profile '%s' 吗？\n\n此操作不可撤销。", m.currentProfile.Name)
	dialog.ShowConfirm("确认删除", message, func(confirmed bool) {
		if !confirmed {
			return
		}
		
		// 执行删除操作
		err := m.profileManager.DeleteProfile(m.currentProfile.ID)
		if err != nil {
			m.showErrorDialog("删除失败", err)
			return
		}
		
		// 清空当前选择
		m.currentProfile = nil
		m.hostEntries = nil
		m.hostEntryList.Refresh()
		
		// 刷新Profile列表
		m.refreshProfileList()
		m.showSuccessDialog("成功", "Profile删除成功")
	}, m.window)
}

// 新增的菜单和工具栏事件处理方法

// onRefresh 刷新数据
func (m *Manager) onRefresh() {
	// 添加panic恢复
	defer m.handlePanic()
	
	// 显示进度对话框
	progressDialog := dialog.NewProgressInfinite("刷新中", "正在刷新数据，请稍候...", m.window)
	progressDialog.Show()
	
	go func() {
		defer progressDialog.Hide()
		
		// 重新加载Profile列表
		if err := m.loadInitialData(); err != nil {
			m.showErrorDialog("刷新失败", err)
			return
		}
		
		m.showSuccessDialog("成功", fmt.Sprintf("已刷新，共加载%d个Profile", len(m.profiles)))
	}()
}

// onCopyProfile 复制Profile
func (m *Manager) onCopyProfile() {
	if m.currentProfile == nil {
		dialog.ShowInformation("提示", "请先选择要复制的Profile", m.window)
		return
	}
	
	// 创建输入对话框获取新名称
	nameEntry := widget.NewEntry()
	nameEntry.SetText(m.currentProfile.Name + "_副本")
	
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "新Profile名称", Widget: nameEntry},
		},
	}
	
	d := dialog.NewCustomConfirm("复制Profile", "确定", "取消", form, func(confirmed bool) {
		if !confirmed {
			return
		}
		
		newName := strings.TrimSpace(nameEntry.Text)
		if newName == "" {
			dialog.ShowError(errors.New("Profile名称不能为空"), m.window)
			return
		}
		
		// 创建新Profile
		newProfile, err := m.profileManager.CreateProfile(newName, m.currentProfile.Description)
		if err != nil {
			dialog.ShowError(err, m.window)
			return
		}
		
		// 复制Host条目
		for _, entry := range m.currentProfile.Entries {
			newEntry := models.NewHostEntry(entry.IP, entry.Hostname, entry.Comment)
			newEntry.Enabled = entry.Enabled
			newProfile.AddEntry(newEntry)
		}
		
		// 保存新Profile
		err = m.profileManager.UpdateProfile(newProfile)
		if err != nil {
			dialog.ShowError(err, m.window)
			return
		}
		
		// 刷新列表
		m.refreshProfileList()
		m.statusBar.SetText(fmt.Sprintf("Profile '%s' 复制成功", newName))
	}, m.window)
	
	d.Resize(fyne.NewSize(350, 150))
	d.Show()
}

// onToggleHostEntry 切换Host条目启用状态
func (m *Manager) onToggleHostEntry() {
	if m.currentHostEntry == nil {
		dialog.ShowInformation("提示", "请先选择要切换状态的Host条目", m.window)
		return
	}
	
	// 切换状态
	m.currentHostEntry.Enabled = !m.currentHostEntry.Enabled
	m.currentHostEntry.UpdatedAt = time.Now()
	
	// 更新Profile
	err := m.profileManager.UpdateProfile(m.currentProfile)
	if err != nil {
		dialog.ShowError(err, m.window)
		return
	}
	
	// 刷新列表
	m.hostEntryList.Refresh()
	
	status := "启用"
	if !m.currentHostEntry.Enabled {
		status = "禁用"
	}
	m.statusBar.SetText(fmt.Sprintf("Host条目 '%s' 已%s", m.currentHostEntry.Hostname, status))
}

// onCleanupBackups 清理备份文件
func (m *Manager) onCleanupBackups() {
	message := "确定要清理过期的备份文件吗？\n\n将删除超过保留期限的备份文件。"
	dialog.ShowConfirm("确认清理", message, func(confirmed bool) {
		if !confirmed {
			return
		}
		
		// 显示进度对话框
		progressDialog := dialog.NewProgressInfinite("清理备份文件", "正在清理过期备份文件，请稍候...", m.window)
		progressDialog.Show()
		
		// 在goroutine中执行清理操作
		go func() {
			defer progressDialog.Hide()
			
			// TODO: 实现备份文件清理逻辑
			// cleanedCount, err := m.hostManager.CleanupBackups()
			
			m.statusBar.SetText("备份文件清理完成")
			dialog.ShowInformation("清理完成", "过期备份文件清理完成", m.window)
		}()
	}, m.window)
}

// onFilterProfiles 过滤Profile显示
func (m *Manager) onFilterProfiles(filter string) {
	// TODO: 实现Profile过滤逻辑
	m.statusBar.SetText(fmt.Sprintf("应用过滤器: %s", filter))
}

// onExpandAll 展开所有
func (m *Manager) onExpandAll() {
	// TODO: 实现展开所有逻辑
	m.statusBar.SetText("已展开所有项目")
}

// onCollapseAll 折叠所有
func (m *Manager) onCollapseAll() {
	// TODO: 实现折叠所有逻辑
	m.statusBar.SetText("已折叠所有项目")
}

// onShowShortcuts 显示快捷键
func (m *Manager) onShowShortcuts() {
	shortcuts := `快捷键列表：

Ctrl+N - 新建Profile
Ctrl+E - 编辑当前Profile
Ctrl+D - 删除当前Profile
Ctrl+C - 复制当前Profile
Ctrl+Q - 快速切换Profile

Ctrl+Shift+N - 添加Host条目
Ctrl+Shift+E - 编辑Host条目
Ctrl+Shift+D - 删除Host条目
Space - 启用/禁用Host条目

Ctrl+A - 应用Profile
Ctrl+B - 备份Hosts文件
F5 - 刷新数据

Ctrl+, - 打开设置
F1 - 显示帮助`
	
	dialog.ShowInformation("快捷键", shortcuts, m.window)
}

// onCheckUpdates 检查更新
func (m *Manager) onCheckUpdates() {
	// TODO: 实现检查更新逻辑
	dialog.ShowInformation("检查更新", "当前版本已是最新版本", m.window)
}

// showErrorDialog 显示错误对话框
func (m *Manager) showErrorDialog(title string, err error) {
	if err == nil {
		return
	}
	
	// 根据错误类型显示不同的错误信息
	errorMsg := err.Error()
	detailedMsg := ""
	
	// 检查是否是已知的错误类型
	switch {
	case strings.Contains(errorMsg, "permission denied"):
		errorMsg = "权限不足，请以管理员身份运行应用程序"
		detailedMsg = "原始错误: " + err.Error()
	case strings.Contains(errorMsg, "file not found"):
		errorMsg = "文件未找到，请检查文件路径是否正确"
		detailedMsg = "原始错误: " + err.Error()
	case strings.Contains(errorMsg, "network"):
		errorMsg = "网络连接错误，请检查网络设置"
		detailedMsg = "原始错误: " + err.Error()
	case strings.Contains(errorMsg, "timeout"):
		errorMsg = "操作超时，请稍后重试"
		detailedMsg = "原始错误: " + err.Error()
	default:
		// 保持原始错误信息
		if len(errorMsg) > 100 {
			detailedMsg = errorMsg
			errorMsg = errorMsg[:100] + "..."
		}
	}
	
	// 创建错误对话框内容
	errorLabel := widget.NewLabel(errorMsg)
	errorLabel.Wrapping = fyne.TextWrapWord
	
	var content fyne.CanvasObject = errorLabel
	
	// 如果有详细信息，添加展开按钮
	if detailedMsg != "" {
		detailsShown := false
		detailsLabel := widget.NewLabel(detailedMsg)
		detailsLabel.Wrapping = fyne.TextWrapWord
		detailsLabel.Hide()
		
		var toggleButton *widget.Button
		toggleButton = widget.NewButton("显示详细信息", func() {
			if detailsShown {
				detailsLabel.Hide()
				toggleButton.SetText("显示详细信息")
				detailsShown = false
			} else {
				detailsLabel.Show()
				toggleButton.SetText("隐藏详细信息")
				detailsShown = true
			}
		})
		
		content = container.NewVBox(
			errorLabel,
			widget.NewSeparator(),
			toggleButton,
			detailsLabel,
		)
	}
	
	// 显示错误对话框
	d := dialog.NewCustom(title, "确定", content, m.window)
	d.Resize(fyne.NewSize(450, 200))
	d.Show()
	
	// 更新状态栏
	m.statusBar.SetText(fmt.Sprintf("错误: %s", errorMsg))
}

// showSuccessDialog 显示成功对话框
func (m *Manager) showSuccessDialog(title, message string) {
	dialog.ShowInformation(title, message, m.window)
	m.statusBar.SetText(message)
}

// showWarningDialog 显示警告对话框
func (m *Manager) showWarningDialog(title, message string, callback func(bool)) {
	dialog.ShowConfirm(title, message, callback, m.window)
}

// handlePanic 处理panic异常
func (m *Manager) handlePanic() {
	if r := recover(); r != nil {
		errorMsg := fmt.Sprintf("应用程序遇到严重错误: %v", r)
		m.showErrorDialog("严重错误", errors.New(errorMsg))
		
		// 记录错误日志
		fmt.Printf("Panic recovered: %v\n", r)
	}
}

// validateInput 验证用户输入
func (m *Manager) validateInput(input string, fieldName string, required bool, maxLength int) error {
	if required && strings.TrimSpace(input) == "" {
		return fmt.Errorf("%s不能为空", fieldName)
	}
	
	if maxLength > 0 && len(input) > maxLength {
		return fmt.Errorf("%s长度不能超过%d个字符", fieldName, maxLength)
	}
	
	return nil
}

// validateIPAddress 验证IP地址格式
func (m *Manager) validateIPAddress(ip string) error {
	if ip == "" {
		return errors.New("IP地址不能为空")
	}
	
	// 简单的IP地址格式验证
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return errors.New("IP地址格式不正确")
	}
	
	for _, part := range parts {
		if part == "" {
			return errors.New("IP地址格式不正确")
		}
		
		// 检查是否为数字
		for _, char := range part {
			if char < '0' || char > '9' {
				return errors.New("IP地址只能包含数字和点")
			}
		}
		
		// 检查范围
		var num int
		if _, err := fmt.Sscanf(part, "%d", &num); err != nil || num < 0 || num > 255 {
			return errors.New("IP地址每段必须在0-255之间")
		}
	}
	
	return nil
}

// validateHostname 验证主机名格式
func (m *Manager) validateHostname(hostname string) error {
	if hostname == "" {
		return errors.New("主机名不能为空")
	}
	
	if len(hostname) > 253 {
		return errors.New("主机名长度不能超过253个字符")
	}
	
	// 检查是否包含非法字符
	for _, char := range hostname {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || 
			(char >= '0' && char <= '9') || char == '.' || char == '-' || char == '_') {
			return errors.New("主机名只能包含字母、数字、点、连字符和下划线")
		}
	}
	
	return nil
}

// updateProfileSelector 更新Profile选择器
func (m *Manager) updateProfileSelector() {
	if m.profileSelector == nil {
		return
	}
	
	// 构建Profile名称列表
	profileNames := make([]string, 0, len(m.profiles))
	for _, profile := range m.profiles {
		name := profile.Name
		if profile.IsActive {
			name += " (当前激活)"
		}
		profileNames = append(profileNames, name)
	}
	
	// 更新选择器选项
	m.profileSelector.Options = profileNames
	m.profileSelector.Refresh()
	
	// 设置当前选中项
	if m.currentProfile != nil {
		for i, profile := range m.profiles {
			if profile.ID == m.currentProfile.ID {
				m.profileSelector.SetSelectedIndex(i)
				break
			}
		}
	}
}

// onQuickSwitchProfile 快速切换Profile
func (m *Manager) onQuickSwitchProfile(selectedName string) {
	if selectedName == "" {
		return
	}
	
	// 移除 " (当前激活)" 后缀
	profileName := strings.Replace(selectedName, " (当前激活)", "", 1)
	
	// 查找对应的Profile
	var targetProfile *models.Profile
	for _, profile := range m.profiles {
		if profile.Name == profileName {
			targetProfile = profile
			break
		}
	}
	
	if targetProfile == nil {
		m.showErrorDialog("切换失败", errors.New("未找到指定的Profile"))
		return
	}
	
	// 切换到目标Profile
	m.switchToProfile(targetProfile)
}

// switchToProfile 切换到指定Profile
func (m *Manager) switchToProfile(profile *models.Profile) {
	if profile == nil {
		return
	}
	
	// 设置当前Profile
	m.currentProfile = profile
	m.hostEntries = profile.Entries
	
	// 更新Profile列表选择
	for i, p := range m.profiles {
		if p.ID == profile.ID {
			m.profileList.Select(i)
			break
		}
	}
	
	// 刷新Host条目列表
	m.hostEntryList.Refresh()
	
	// 更新状态栏
	m.statusBar.SetText(fmt.Sprintf("已切换到Profile: %s (包含 %d 个Host条目)", profile.Name, len(profile.Entries)))
	
	// 更新Profile选择器
	m.updateProfileSelector()
}

// showQuickSwitchDialog 显示快速切换对话框
func (m *Manager) showQuickSwitchDialog() {
	if len(m.profiles) == 0 {
		dialog.ShowInformation("提示", "没有可用的Profile", m.window)
		return
	}
	
	var selectedProfile *models.Profile
	
	// 创建Profile列表
	profileList := widget.NewList(
		func() int {
			return len(m.profiles)
		},
		func() fyne.CanvasObject {
			name := widget.NewLabel("")
			name.TextStyle.Bold = true
			status := widget.NewLabel("")
			return container.NewVBox(name, status)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= 0 && id < len(m.profiles) {
				profile := m.profiles[id]
				vbox := obj.(*fyne.Container)
				
				nameLabel := vbox.Objects[0].(*widget.Label)
				nameLabel.SetText(profile.Name)
				
				statusLabel := vbox.Objects[1].(*widget.Label)
				statusText := fmt.Sprintf("%d个条目", len(profile.Entries))
				if profile.IsActive {
					statusText += " (当前激活)"
				}
				statusLabel.SetText(statusText)
			}
		},
	)
	
	// 设置选择事件
	profileList.OnSelected = func(id widget.ListItemID) {
		if id >= 0 && id < len(m.profiles) {
			selectedProfile = m.profiles[id]
		}
	}
	
	// 设置当前选中项
	if m.currentProfile != nil {
		for i, profile := range m.profiles {
			if profile.ID == m.currentProfile.ID {
				profileList.Select(i)
				selectedProfile = profile
				break
			}
		}
	}
	
	// 创建对话框
	d := dialog.NewCustomConfirm("快速切换Profile", "切换", "取消", 
		container.NewBorder(nil, nil, nil, nil, profileList), 
		func(confirmed bool) {
			if !confirmed {
				return
			}
			
			if selectedProfile != nil {
				m.switchToProfile(selectedProfile)
			}
		}, m.window)
	
	d.Resize(fyne.NewSize(400, 300))
	d.Show()
}

func (m *Manager) onImportProfile() { /* TODO: 实现导入Profile */ }
func (m *Manager) onExportProfile() { /* TODO: 实现导出Profile */ }
func (m *Manager) onRestoreHosts()  { /* TODO: 实现恢复Hosts */ }
func (m *Manager) onValidateHosts() { /* TODO: 实现验证Hosts */ }
func (m *Manager) onCleanupHosts()  { /* TODO: 实现清理Hosts */ }
func (m *Manager) onShowAbout()     { /* TODO: 实现显示关于 */ }
func (m *Manager) onShowHelp()      { /* TODO: 实现显示帮助 */ }
