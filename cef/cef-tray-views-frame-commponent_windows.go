//----------------------------------------
//
// Copyright © yanghy. All Rights Reserved.
//
// Licensed under GNU General Public License v3.0
//
//----------------------------------------

//go:build windows
// +build windows

package cef

import (
	"github.com/energye/energy/common"
	"github.com/energye/energy/common/assetserve"
	"github.com/energye/energy/consts"
	"github.com/energye/energy/ipc"
	"github.com/energye/energy/logger"
	"github.com/energye/golcl/lcl"
	"github.com/energye/golcl/lcl/types"
)

func newViewsFrameTray(owner lcl.IComponent, width, height int32, url string) *ViewsFrameTray {
	var tray = &ViewsFrameTray{}
	cc := NewChromiumConfig()
	cc.SetEnableMenu(false)
	wp := NewWindowProperty()
	wp.Title = ""
	wp.Url = url
	wp.Width = width
	wp.Height = height
	wp.X = -width
	wp.Y = -height
	wp.AlwaysOnTop = true
	wp.EnableMaximize = false
	wp.EnableMinimize = false
	wp.EnableResize = false
	wp.EnableCenterWindow = false
	tray.trayWindow = NewViewsFrameworkBrowserWindow(cc, wp)
	tray.trayWindow.ResetWindowPropertyForEvent()
	tray.trayIcon = lcl.NewTrayIcon(owner)
	tray.trayIcon.SetVisible(true)
	tray.x = wp.X
	tray.y = wp.Y
	tray.w = wp.Width
	tray.h = wp.Height
	tray.registerMouseEvent()
	tray.registerChromiumEvent()
	return tray
}

func (m *ViewsFrameTray) registerMouseEvent() {
	m.trayWindow.WindowComponent().SetOnWindowActivationChanged(func(sender lcl.IObject, window *ICefWindow, active bool) {
		if active {
		} else {
			m.trayWindow.Hide()
		}
	})
	var isCreateTopLevelWindow = true
	m.trayIcon.SetOnMouseUp(func(sender lcl.IObject, button types.TMouseButton, shift types.TShiftState, x, y int32) {
		if isCreateTopLevelWindow {
			isCreateTopLevelWindow = false
			//m.trayWindow.windowId = BrowserWindow.GetNextWindowNum()
			//m.trayWindow.putChromiumWindowInfo()
			BrowserWindow.setOrIncNextWindowNum() //明确的生成下一个窗口序号
			m.trayWindow.CreateTopLevelWindow()
			m.trayWindow.HideTitle()
			m.trayWindow.SetNotInTaskBar()
			m.trayWindow.WindowComponent().SetAlwaysOnTop(true)
		}
		display := m.trayWindow.WindowComponent().Display()
		bounds := display.Bounds()
		var monitorWidth = bounds.Width
		width, height := m.w, m.h
		var mx = x + width
		var my = y + height
		if mx < monitorWidth {
			mx = x
		} else {
			mx = x - width
		}
		if my > m.h {
			my = y
		}
		if my > height {
			my = y - height
		}
		var ret bool
		if m.mouseUp != nil {
			ret = m.mouseUp(sender, button, shift, x, y)
		}
		if !ret {
			if button == types.MbRight {
				m.trayWindow.WindowComponent().SetBounds(NewCefRect(mx, my, width, height))
				m.trayWindow.Show()
				m.trayWindow.BrowserViewComponent().RequestFocus()
			}
		}
	})
}

func (m *ViewsFrameTray) registerChromiumEvent() {
	m.trayWindow.Chromium().SetOnBeforeContextMenu(func(sender lcl.IObject, browser *ICefBrowser, frame *ICefFrame, params *ICefContextMenuParams, model *ICefMenuModel) {
		model.Clear()
	})
	//m.trayWindow.Chromium().SetOnBeforeBrowser(func(sender lcl.IObject, browser *ICefBrowser, frame *ICefFrame) bool {
	//	BrowserWindow.setOrIncNextWindowNum(browser.Identifier() + 1)
	//	return false
	//})
	m.trayWindow.Chromium().SetOnBeforeResourceLoad(func(sender lcl.IObject, browser *ICefBrowser, frame *ICefFrame, request *ICefRequest, callback *ICefCallback, result *consts.TCefReturnValue) {
		if assetserve.AssetsServerHeaderKeyValue != "" {
			request.SetHeaderByName(assetserve.AssetsServerHeaderKeyName, assetserve.AssetsServerHeaderKeyValue, true)
		}
	})
	m.trayWindow.Chromium().SetOnBeforeClose(func(sender lcl.IObject, browser *ICefBrowser) {
		logger.Debug("tray.chromium.onBeforeClose")
		m.close()
	})
	m.trayWindow.Chromium().SetOnProcessMessageReceived(func(sender lcl.IObject, browser *ICefBrowser, frame *ICefFrame, sourceProcess consts.CefProcessId, message *ipc.ICefProcessMessage) bool {
		return false
	})
}

func (m *ViewsFrameTray) AsSysTray() *SysTray {
	return nil
}

func (m *ViewsFrameTray) AsViewsFrameTray() *ViewsFrameTray {
	return m
}

func (m *ViewsFrameTray) AsCEFTray() *CEFTray {
	return nil
}

func (m *ViewsFrameTray) AsLCLTray() *LCLTray {
	return nil
}

func (m *ViewsFrameTray) Show() {
	m.trayWindow.Show()
}

func (m *ViewsFrameTray) Hide() {
	m.trayWindow.Hide()
}

func (m *ViewsFrameTray) close() {
	if m.isClosing {
		return
	}
	m.trayIcon.SetVisible(false)
	m.Hide()
	m.trayIcon.Free()
}

func (m *ViewsFrameTray) SetOnDblClick(fn TrayICONClick) {
	m.trayIcon.SetOnDblClick(func(sender lcl.IObject) {
		fn()
	})
}

func (m *ViewsFrameTray) SetOnClick(fn TrayICONClick) {
	m.trayIcon.SetOnClick(func(sender lcl.IObject) {
		fn()
	})
}

func (m *ViewsFrameTray) SetOnMouseUp(fn TMouseEvent) {
	m.mouseUp = fn
}

func (m *ViewsFrameTray) SetOnMouseDown(fn lcl.TMouseEvent) {
	m.trayIcon.SetOnMouseDown(fn)
}

func (m *ViewsFrameTray) SetOnMouseMove(fn lcl.TMouseMoveEvent) {
	m.trayIcon.SetOnMouseMove(fn)
}

func (m *ViewsFrameTray) Visible() bool {
	return m.visible
}

func (m *ViewsFrameTray) SetVisible(v bool) {
	m.visible = v
	m.trayIcon.SetVisible(v)
}

func (m *ViewsFrameTray) SetHint(value string) {
	m.trayIcon.SetHint(value)
}

func (m *ViewsFrameTray) SetTitle(title string) {
}

//显示系统通知
//
//title 标题
//
//content 内容
//
//timeout 显示时间(毫秒)
func (m *ViewsFrameTray) Notice(title, content string, timeout int32) {
	if common.IsWindows() {
		notification(m.trayIcon, title, content, timeout)
	} else {
		notification(nil, title, content, timeout)
	}
}

//设置托盘图标
func (m *ViewsFrameTray) SetIconFS(iconResourcePath string) {
	m.trayIcon.Icon().LoadFromFSFile(iconResourcePath)
}

//设置托盘图标
func (m *ViewsFrameTray) SetIcon(iconResourcePath string) {
	m.trayIcon.Icon().LoadFromFile(iconResourcePath)
}