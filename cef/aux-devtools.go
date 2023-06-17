//----------------------------------------
//
// Copyright © yanghy. All Rights Reserved.
//
// Licensed under Apache License Version 2.0, January 2004
//
// https://www.apache.org/licenses/LICENSE-2.0
//
//----------------------------------------

// 辅助工具-开发者工具

package cef

import (
	"fmt"
	"github.com/energye/energy/v2/common"
	"github.com/energye/golcl/lcl"
)

const (
	devToolsName = "dev-tools"
)

type devToolsWindow struct {
	*lcl.TForm
	parent ICEFWindowParent
}

func updateBrowserDevTools(browser *ICefBrowser, title string) {
	if browserWinInfo := BrowserWindow.GetWindowInfo(browser.Identifier()); browserWinInfo != nil {
		if browserWinInfo.IsLCL() {
			window := browserWinInfo.AsLCLBrowserWindow().BrowserWindow()
			if window.GetAuxTools() != nil && window.GetAuxTools().DevTools() != nil {
				window.GetAuxTools().DevTools().SetCaption(fmt.Sprintf("%s - %s", devToolsName, browser.MainFrame().Url()))
			}
		}
	}
}

func (m *ICefBrowser) createBrowserDevTools(browserWindow IBrowserWindow) {
	if browserWindow.IsLCL() {
		if common.IsWindows() {
			window := browserWindow.AsLCLBrowserWindow().BrowserWindow()
			QueueAsyncCall(func(id int) { // show window, run is main ui thread
				window.GetAuxTools().DevTools().Show()
				browserWindow.Chromium().ShowDevTools(window.GetAuxTools().DevTools().WindowParent())
			})
		} else {
			browserWindow.Chromium().ShowDevTools(nil)
		}
	} else if browserWindow.IsViewsFramework() {
		browserWindow.Chromium().ShowDevTools(nil)
	}
}

func (m *devToolsWindow) WindowParent() ICEFWindowParent {
	return m.parent
}

func (m *devToolsWindow) SetWindowParent(parent ICEFWindowParent) {
	m.parent = parent
}