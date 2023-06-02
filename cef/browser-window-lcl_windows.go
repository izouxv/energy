//----------------------------------------
//
// Copyright © yanghy. All Rights Reserved.
//
// Licensed under Apache License Version 2.0, January 2004
//
// https://www.apache.org/licenses/LICENSE-2.0
//
//----------------------------------------

//go:build windows

// LCL窗口组件定义和实现-windows平台

package cef

import (
	"github.com/energye/energy/v2/consts"
	"github.com/energye/golcl/lcl"
	"github.com/energye/golcl/lcl/rtl"
	"github.com/energye/golcl/lcl/types"
	"github.com/energye/golcl/lcl/win"
)

// 定义四角和边框范围
var (
	angleRange  int32 = 10 //四角
	borderRange int32 = 5  //四边框
)

// customWindowCaption 自定义窗口标题栏
//
// 隐藏窗口标题栏，通过html+css实现自定义窗口标题栏，实现窗口拖拽等
type customWindowCaption struct {
	bw                   *LCLBrowserWindow     //
	canCaption           bool                  //当前鼠标是否在标题栏区域
	canBorder            bool                  //当前鼠标是否在边框
	borderHT, borderWMSZ int                   //borderHT: 鼠标所在边框位置, borderWMSZ: 窗口改变大小边框方向 borderMD:
	borderMD             bool                  //borderMD: 鼠标调整窗口大小，已按下后，再次接收到132消息应该忽略该消息
	regions              *TCefDraggableRegions //窗口内html拖拽区域
	rgn                  *HRGN                 //
}

// ShowTitle 显示标题栏
func (m *LCLBrowserWindow) ShowTitle() {
	m.WindowProperty().EnableHideCaption = false
	//win.SetWindowLong(m.Handle(), win.GWL_STYLE, uintptr(win.GetWindowLong(m.Handle(), win.GWL_STYLE)|win.WS_CAPTION))
	//win.SetWindowPos(m.Handle(), m.Handle(), 0, 0, 0, 0, win.SWP_NOSIZE|win.SWP_NOMOVE|win.SWP_NOZORDER|win.SWP_NOACTIVATE|win.SWP_FRAMECHANGED)
	m.EnabledMaximize(m.WindowProperty().EnableMaximize)
	m.EnabledMinimize(m.WindowProperty().EnableMinimize)
	m.SetBorderStyle(types.BsSizeable)
}

// HideTitle 隐藏标题栏
func (m *LCLBrowserWindow) HideTitle() {
	m.WindowProperty().EnableHideCaption = true
	//win.SetWindowLong(m.Handle(), win.GWL_STYLE, uintptr(win.GetWindowLong(m.Handle(), win.GWL_STYLE)&^win.WS_CAPTION))
	//win.SetWindowPos(m.Handle(), 0, 0, 0, m.Width(), m.Height()+500, win.SWP_NOMOVE|win.SWP_NOZORDER|win.SWP_NOACTIVATE|win.SWP_FRAMECHANGED|win.SWP_DRAWFRAME)
	//无标题栏情况会导致任务栏不能切换窗口，不知道为什么要这样设置一下
	m.EnabledMaximize(false)
	m.SetBorderStyle(types.BsNone)

}

// freeRgn
func (m *customWindowCaption) freeRgn() {
	if m.rgn != nil {
		WinSetRectRgn(m.rgn, 0, 0, 0, 0)
		WinDeleteObject(m.rgn)
		m.rgn.Free()
	}
}

// freeRegions
func (m *customWindowCaption) freeRegions() {
	if m.regions != nil {
		m.regions.regions = nil
		m.regions = nil
	}
}

// free
func (m *customWindowCaption) free() {
	if m != nil {
		m.freeRgn()
		m.freeRegions()
	}
}

// onNCMouseMove NC 非客户区鼠标移动
func (m *customWindowCaption) onNCMouseMove(message *types.TMessage, lResult *types.LRESULT, aHandled *bool) {
	if m.canCaption { // 当前在标题栏
	} else if m.canBorder { // 当前在边框
		*lResult = types.LRESULT(m.borderHT)
		*aHandled = true
	}
}

// onSetCursor 设置鼠标图标
func (m *customWindowCaption) onSetCursor(message *types.TMessage, lResult *types.LRESULT, aHandled *bool) {
	if m.canBorder { //当前在边框
		switch LOWORD(uint32(message.LParam)) {
		case consts.HTBOTTOMRIGHT, consts.HTTOPLEFT: //右下 左上
			*lResult = types.LRESULT(m.borderHT)
			*aHandled = true
			WinSetCursor(WinLoadCursor(0, IDC_SIZENWSE))
		case consts.HTRIGHT, consts.HTLEFT: //右 左
			*lResult = types.LRESULT(m.borderHT)
			*aHandled = true
			WinSetCursor(WinLoadCursor(0, IDC_SIZEWE))
		case consts.HTTOPRIGHT, consts.HTBOTTOMLEFT: //右上 左下
			*lResult = types.LRESULT(m.borderHT)
			*aHandled = true
			WinSetCursor(WinLoadCursor(0, IDC_SIZENESW))
		case consts.HTTOP, consts.HTBOTTOM: //上 下
			*lResult = types.LRESULT(m.borderHT)
			*aHandled = true
			WinSetCursor(WinLoadCursor(0, IDC_SIZENS))
		}
	}
}

// onCanBorder 鼠标是否在边框
func (m *customWindowCaption) onCanBorder(x, y int32, rect *types.TRect) (int, bool) {
	if m.canBorder = x <= rect.Width() && x >= rect.Width()-angleRange && y <= angleRange; m.canBorder { // 右上
		m.borderWMSZ = WMSZ_TOPRIGHT
		m.borderHT = consts.HTTOPRIGHT
		return m.borderHT, true
	} else if m.canBorder = x <= rect.Width() && x >= rect.Width()-angleRange && y <= rect.Height() && y >= rect.Height()-angleRange; m.canBorder { // 右下
		m.borderWMSZ = WMSZ_BOTTOMRIGHT
		m.borderHT = consts.HTBOTTOMRIGHT
		return m.borderHT, true
	} else if m.canBorder = x <= angleRange && y <= angleRange; m.canBorder { //左上
		m.borderWMSZ = WMSZ_TOPLEFT
		m.borderHT = consts.HTTOPLEFT
		return m.borderHT, true
	} else if m.canBorder = x <= angleRange && y >= rect.Height()-angleRange; m.canBorder { //左下
		m.borderWMSZ = WMSZ_BOTTOMLEFT
		m.borderHT = consts.HTBOTTOMLEFT
		return m.borderHT, true
	} else if m.canBorder = x > angleRange && x < rect.Width()-angleRange && y <= borderRange; m.canBorder { //上
		m.borderWMSZ = WMSZ_TOP
		m.borderHT = consts.HTTOP
		return m.borderHT, true
	} else if m.canBorder = x > angleRange && x < rect.Width()-angleRange && y >= rect.Height()-borderRange; m.canBorder { //下
		m.borderWMSZ = WMSZ_BOTTOM
		m.borderHT = consts.HTBOTTOM
		return m.borderHT, true
	} else if m.canBorder = x <= borderRange && y > angleRange && y < rect.Height()-angleRange; m.canBorder { //左
		m.borderWMSZ = WMSZ_LEFT
		m.borderHT = consts.HTLEFT
		return m.borderHT, true
	} else if m.canBorder = x <= rect.Width() && x >= rect.Width()-borderRange && y > angleRange && y < rect.Height()-angleRange; m.canBorder { // 右
		m.borderWMSZ = WMSZ_RIGHT
		m.borderHT = consts.HTRIGHT
		return m.borderHT, true
	}
	return 0, false
}

// onNCLButtonDown NC 鼠标左键按下
func (m *customWindowCaption) onNCLButtonDown(hWND types.HWND, message *types.TMessage, lResult *types.LRESULT, aHandled *bool) {
	if m.canCaption { // 标题栏
		x, y := m.toPoint(message)
		*lResult = consts.HTCAPTION
		m.borderMD = true
		*aHandled = true
		win.ReleaseCapture()
		rtl.PostMessage(hWND, consts.WM_NCLBUTTONDOWN, consts.HTCAPTION, rtl.MakeLParam(uint16(x), uint16(y)))
	} else if m.canBorder { // 边框
		x, y := m.toPoint(message)
		*lResult = types.LRESULT(m.borderHT)
		m.borderMD = true
		*aHandled = true
		win.ReleaseCapture()
		rtl.PostMessage(hWND, consts.WM_SYSCOMMAND, uintptr(consts.SC_SIZE|m.borderWMSZ), rtl.MakeLParam(uint16(x), uint16(y)))
		//rtl.PostMessage(hWND, WM_SYSCOMMAND, uintptr(SC_SIZE|m.borderWMSZ), 0)
	}
}

// toPoint 转换XY坐标
func (m *customWindowCaption) toPoint(message *types.TMessage) (x, y int32) {
	return GET_X_LPARAM(message.LParam), GET_Y_LPARAM(message.LParam)
}

// isCaption
// 鼠标是否在标题栏区域
//
// 如果启用了css拖拽则校验拖拽区域,否则只返回相对于浏览器窗口的x,y坐标
func (m *customWindowCaption) isCaption(hWND types.HWND, message *types.TMessage) (int32, int32, bool) {
	dx, dy := m.toPoint(message)
	p := &types.TPoint{
		X: dx,
		Y: dy,
	}
	WinScreenToClient(hWND, p)
	p.X -= m.bw.WindowParent().Left()
	p.Y -= m.bw.WindowParent().Top()
	if m.bw.WindowProperty().EnableWebkitAppRegion && m.rgn != nil {
		m.canCaption = WinPtInRegion(m.rgn, p.X, p.Y)
	} else {
		m.canCaption = false
	}
	return p.X, p.Y, m.canCaption
}

// doOnRenderCompMsg
func (m *LCLBrowserWindow) doOnRenderCompMsg(message *types.TMessage, lResult *types.LRESULT, aHandled *bool) {
	switch message.Msg {
	case consts.WM_NCLBUTTONDBLCLK: // 163 NC left dclick
		//标题栏拖拽区域 双击最大化和还原
		if m.cwcap.canCaption && m.WindowProperty().EnableWebkitAppRegionDClk {
			*lResult = consts.HTCAPTION
			*aHandled = true
			win.ReleaseCapture()
			m.windowProperty.windowState = m.WindowState()
			if m.windowProperty.windowState == types.WsNormal {
				rtl.PostMessage(m.Handle(), consts.WM_SYSCOMMAND, consts.SC_MAXIMIZE, 0)
			} else {
				rtl.PostMessage(m.Handle(), consts.WM_SYSCOMMAND, consts.SC_RESTORE, 0)
			}
			rtl.SendMessage(m.Handle(), consts.WM_NCLBUTTONUP, consts.HTCAPTION, 0)
		}
	case consts.WM_NCLBUTTONDOWN: // 161 nc left down
		m.cwcap.onNCLButtonDown(m.Handle(), message, lResult, aHandled)
	case consts.WM_NCLBUTTONUP: // 162 nc l up
		if m.cwcap.canCaption {
			*lResult = consts.HTCAPTION
			*aHandled = true
		}
	case consts.WM_NCMOUSEMOVE: // 160 nc mouse move
		m.cwcap.onNCMouseMove(message, lResult, aHandled)
	case consts.WM_SETCURSOR: // 32 设置鼠标图标样式
		m.cwcap.onSetCursor(message, lResult, aHandled)
	case consts.WM_NCHITTEST: // 132 NCHITTEST
		if m.cwcap.borderMD { //TODO 测试windows7, 161消息之后再次处理132消息导致消息错误
			m.cwcap.borderMD = false
			return
		}
		//鼠标坐标是否在标题区域
		x, y, caption := m.cwcap.isCaption(m.Handle(), message)
		if caption { //窗口标题栏
			*lResult = consts.HTCAPTION
			*aHandled = true
		} else if m.WindowProperty().EnableHideCaption && m.WindowProperty().EnableResize && m.WindowState() == types.WsNormal { //1.窗口隐藏标题栏 2.启用了调整窗口大小 3.非最大化、最小化、全屏状态
			rect := m.BoundsRect()
			if result, handled := m.cwcap.onCanBorder(x, y, &rect); handled {
				*lResult = types.LRESULT(result)
				*aHandled = true
			}
		}
	}
}

// setDraggableRegions
// 每一次拖拽区域改变都需要重新设置
func (m *LCLBrowserWindow) setDraggableRegions() {
	//在主线程中运行
	QueueAsyncCall(func(id int) {
		if m.cwcap.rgn == nil {
			//第一次时创建RGN
			m.cwcap.rgn = WinCreateRectRgn(0, 0, 0, 0)
		} else {
			//每次重置RGN
			WinSetRectRgn(m.cwcap.rgn, 0, 0, 0, 0)
		}
		for i := 0; i < m.cwcap.regions.RegionsCount(); i++ {
			region := m.cwcap.regions.Region(i)
			creRGN := WinCreateRectRgn(region.Bounds.X, region.Bounds.Y, region.Bounds.X+region.Bounds.Width, region.Bounds.Y+region.Bounds.Height)
			if region.Draggable {
				WinCombineRgn(m.cwcap.rgn, m.cwcap.rgn, creRGN, consts.RGN_OR)
			} else {
				WinCombineRgn(m.cwcap.rgn, m.cwcap.rgn, creRGN, consts.RGN_DIFF)
			}
			WinDeleteObject(creRGN)
		}
	})
}

// registerWindowsCompMsgEvent
// 注册windows下CompMsg事件
func (m *LCLBrowserWindow) registerWindowsCompMsgEvent() {
	var bwEvent = BrowserWindow.browserEvent
	m.Chromium().SetOnRenderCompMsg(func(sender lcl.IObject, message *types.TMessage, lResult *types.LRESULT, aHandled *bool) {
		if bwEvent.onRenderCompMsg != nil {
			bwEvent.onRenderCompMsg(sender, message, lResult, aHandled)
		}
		if !*aHandled {
			m.doOnRenderCompMsg(message, lResult, aHandled)
		}
	})
	if m.WindowProperty().EnableWebkitAppRegion && m.WindowProperty().EnableWebkitAppRegionDClk {
		m.windowResize = func(sender lcl.IObject) bool {
			if m.WindowState() == types.WsMaximized && (m.WindowProperty().EnableHideCaption || m.BorderStyle() == types.BsNone || m.BorderStyle() == types.BsSingle) {
				var monitor = m.Monitor().WorkareaRect()
				m.SetBounds(monitor.Left, monitor.Top, monitor.Right-monitor.Left, monitor.Bottom-monitor.Top)
				m.SetWindowState(types.WsMaximized)
			}
			return false
		}
	}

	//if m.WindowProperty().EnableWebkitAppRegion {
	//
	//} else {
	//	if bwEvent.onRenderCompMsg != nil {
	//		m.chromium.SetOnRenderCompMsg(bwEvent.onRenderCompMsg)
	//	}
	//}
}

// Maximize Windows平台，窗口最大化/还原
func (m *LCLBrowserWindow) Maximize() {
	if m.TForm == nil {
		return
	}
	QueueAsyncCall(func(id int) {
		win.ReleaseCapture()
		m.windowProperty.windowState = m.WindowState()
		if m.windowProperty.windowState == types.WsNormal {
			rtl.PostMessage(m.Handle(), consts.WM_SYSCOMMAND, consts.SC_MAXIMIZE, 0)
		} else {
			rtl.SendMessage(m.Handle(), consts.WM_SYSCOMMAND, consts.SC_RESTORE, 0)
		}
	})
}

// 窗口透明
//func (m *LCLBrowserWindow) SetTransparentColor() {
//	m.SetColor(colors.ClNavy)
//	Exstyle := win.GetWindowLong(m.Handle(), win.GWL_EXSTYLE)
//	Exstyle = Exstyle | win.WS_EX_LAYERED&^win.WS_EX_TRANSPARENT
//	win.SetWindowLong(m.Handle(), win.GWL_EXSTYLE, uintptr(Exstyle))
//	win.SetLayeredWindowAttributes(m.Handle(),
//		colors.ClNavy, //crKey 指定需要透明的背景颜色值
//		255,           //bAlpha 设置透明度,0完全透明，255不透明
//		//LWA_ALPHA: crKey无效,bAlpha有效
//		//LWA_COLORKEY: 窗体中的所有颜色为crKey的地方全透明，bAlpha无效
//		//LWA_ALPHA | LWA_COLORKEY: crKey的地方全透明，其它地方根据bAlpha确定透明度
//		win.LWA_ALPHA|win.LWA_COLORKEY)
//}