package main

import (
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"

	"github.com/flyhigher139/mhost/internal/ui"
)

// main 应用程序入口点
func main() {
	// 创建Fyne应用
	myApp := app.NewWithID("com.gevin.mhost")

	// 创建主窗口
	mainWindow := myApp.NewWindow("mHost - Hosts文件管理器")
	mainWindow.Resize(fyne.NewSize(1200, 800))
	mainWindow.CenterOnScreen()

	// 创建UI管理器
	uiManager, err := ui.NewManager(mainWindow)
	if err != nil {
		log.Printf("Failed to create UI manager: %v", err)
		showErrorDialog(mainWindow, "初始化失败", "无法初始化应用程序: "+err.Error())
		os.Exit(1)
	}

	// 设置窗口内容
	mainWindow.SetContent(uiManager.GetMainContainer())

	// 设置窗口关闭回调
	mainWindow.SetCloseIntercept(func() {
		uiManager.OnWindowClose()
		myApp.Quit()
	})

	// 显示窗口并运行应用
	mainWindow.ShowAndRun()
}

// showErrorDialog 显示错误对话框
func showErrorDialog(parent fyne.Window, title, message string) {
	errorDialog := widget.NewCard(title, "", widget.NewLabel(message))
	popup := widget.NewModalPopUp(errorDialog, parent.Canvas())
	popup.Resize(fyne.NewSize(400, 200))
	popup.Show()
}
