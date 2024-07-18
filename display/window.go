/*
   main,
   Copyright (C) 2024  Phil Hilger

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package display

import (
	it8951 "github.com/peergum/IT8951-go"
	"nv/display/fonts-go"
	"nv/input"
)

type WindowOptions struct {
	Title        string
	TitleBar     bool
	Border       int
	BgColor      it8951.Color
	BorderColor  it8951.Color
	Transparency float64 // 0 (opaque) to 1 (transparent)
	Radius       int
	TopRounded   bool
	StatusBar    *StatusBar // nil for most windows
	Bpp          int
}

type Window struct {
	*View
	WindowOptions
	titleView      *View
	Views          []*View
	updated        bool
	visible        bool
	parent         *View
	titleWidth     int
	RefreshChannel chan bool
}

type ScreenView struct {
	*View
	Windows []*Window
}

const (
	titleBorder        = 2
	titleHeight        = 48
	contentBorder      = 1
	titleBarBgColor    = Gray13
	titleColor         = Black
	defaultTitleRadius = 20
)

var (
	Screen       ScreenView
	titleBarFont = &fonts.SpecialElite_Regular20pt8b
)

func init() {
}

func Ppb(bpp int) int {
	switch bpp {
	case 1:
		return 8
	case 2:
		return 4
	case 3:
		return 2
	case 4:
		return 2
	case 8:
		return 1
	}
	return 1
}

func (view *View) NewCenteredWindow(W, H int, options WindowOptions) *Window {
	return view.NewWindow((view.InnerW-W)/2, (view.InnerH-H)/2, W, H, options)
}

func (view *View) NewWindow(x, y, w, h int, options WindowOptions) *Window {
	Debug("Create a New Window (%d,%d,%d,%d,\"%s\") inside (%d,%d,%d,%d)", x, y, w, h, options.Title,
		view.X, view.Y, view.W, view.H)
	CancelAlertBox()
	var window Window
	window.parent = view
	if options.Bpp == 0 {
		options.Bpp = 4
	}
	window.WindowOptions = options

	window.RefreshChannel = make(chan bool)

	//create an internal view to use for content
	window.View = view.NewView(x, y, w, h, options.Bpp)
	//if options.Bpp == 1 {
	//	window.View.Fill(0, options.BgColor, Black) //.Update()
	//}
	window.Views = make([]*View, 0, 5)

	// prepare for first appearance
	window.updated = true
	window.Show()
	return &window
}

func (window *Window) Show() {
	window.InnerX = window.X
	window.InnerY = window.Y
	window.InnerW = window.W
	window.InnerH = window.H
	titleBarHeight := 0
	if window.WindowOptions.TitleBar {
		titleBarHeight = titleHeight + titleBorder
	}
	if window.updated {
		if window.WindowOptions.Transparency > 0 {
			for _, parent := range window.parent.Views {
				window.View.CopyPixelsWithTransparency(window.X, window.Y+titleBarHeight, parent, window.X, window.Y+titleBarHeight, window.W, window.H-titleBarHeight, window.WindowOptions.Transparency, uint16(window.WindowOptions.BgColor))
			}
			//if window.WindowOptions.TopRounded {
			//	window.View.TopRoundedRectangle(0, 0, window.W, window.H, window.Border, window.BorderColor, window.WindowOptions.Radius).Update()
			//} else {
			window.View.RoundedRectangle(0, titleBarHeight, window.W, window.H-titleBarHeight, window.Border, window.BorderColor, window.WindowOptions.Radius).Update()
			//}
		} else {
			for _, parent := range window.parent.Views {
				window.View.CopyPixels(window.X, window.Y, parent, window.X, window.Y, window.W, window.H)
			}
			//if window.WindowOptions.TopRounded {
			//	window.View.FillTopRoundedRectangle(0, 0, window.W, window.H, window.Border, window.WindowOptions.BgColor, window.BorderColor, window.WindowOptions.Radius).Update()
			//} else {
			window.View.FillRoundedRectangle(0, titleBarHeight, window.W, window.H-titleBarHeight, window.Border, window.WindowOptions.BgColor, window.BorderColor, window.WindowOptions.Radius).Update()
			//}
		}
	}
	window.InnerX = window.X + window.Border
	window.InnerY = window.Y + window.Border
	window.InnerW = window.W - 2*window.Border
	window.InnerH = window.H - 2*window.Border

	if window.updated {
		if window.WindowOptions.TitleBar {
			window.InnerY += titleBarHeight
			window.InnerH -= titleBarHeight
			window.RefreshTitleBar()
		}
		//it8951.WaitForDisplayReady()
		if window.content != nil {
			window.content.Print() // updates content on screen
			//window.Update()
		}
	}
	window.updated = false
	window.visible = true
}

func (window *Window) Hide() {
	//Debug("Hide window %x", window)
	window.visible = false
	window.updated = true

	//window.parent.Refresh(window.X, window.Y, window.W, window.H)

	//// create a temporary new view to restore parent view partially faster

	//for _, parentView := range window.parent.Views {
	//	if parentView == window.View || parentView == nil {
	//		continue
	//	}
	//	//parentView.Refresh(window.X, window.Y, window.W, window.H)
	//	view := Screen.NewView(window.X, window.Y, window.W, window.H, 4)
	//	// copy content from parent
	//	view.CopyPixels(view.X, view.Y, parentView, window.X, window.Y, window.W, window.H)
	//	view.Update()
	//}

	for _, parentView := range window.parent.Views {
		if parentView == window.View {
			continue
		}
		//	// copy content from parent
		//	view.CopyPixels(view.X, view.Y, parentView, window.X, window.Y, window.W, window.H)
		parentView.Refresh(window.X, window.Y, window.W, window.H)
	}

}

func (window *Window) Close() {
	window.Hide()
	window.View = nil
}

func (window *Window) RefreshTitleBar() {
	if !window.WindowOptions.TitleBar {
		return
	}

	titleBarHeight := titleHeight + titleBorder
	window.InnerY -= titleBarHeight
	window.InnerH += titleBarHeight

	currentTitleWidth := window.titleWidth
	if window.WindowOptions.StatusBar != nil {
		//actualWidth = window.W - window.WindowOptions.StatusBar.W
	}
	if currentTitleWidth > 0 {
		for _, parent := range window.parent.Views {
			if parent != nil {
				window.CopyPixels(window.X, window.Y, parent, window.X, window.Y, currentTitleWidth, titleBarHeight)
			}
		}

		window.Refresh(window.X, window.Y, currentTitleWidth, titleBarHeight)
	}

	// determine title width
	x0, y0 := 0, 0
	window.SetTextArea(titleBarFont, 0, 0)
	title := window.Title
	ellipsis := "" // no ellipsis by default
	statusBarLength := 0
	if window.StatusBar != nil {
		statusBarLength = window.StatusBar.W
	}
	for {
		_, _, window.titleWidth, _ = window.GetTextBounds(title+ellipsis, &x0, &y0)
		window.titleWidth += 60 // allow for some margin
		if window.titleWidth <= window.W-statusBarLength-10 {
			title += ellipsis
			break
		}
		title = title[:len(title)-1]
		ellipsis = "..."
	}
	window.titleView = window.NewView(-1, 0, window.titleWidth, titleBarHeight, 4).
		Fill(0, titleBarBgColor, Black).
		SetTextArea(titleBarFont, 0, 0)

	window.titleView.FillTopRoundedRectangle(0, 0, window.titleView.W, window.titleView.H, window.WindowOptions.Border, titleBarBgColor, window.BorderColor, defaultTitleRadius)

	window.titleView.
		DrawHLine(0, window.titleView.H-titleBorder, window.titleView.W, titleBorder, 0x0) //.
	//DrawVLine(window.titleView.W/2-titleBorder, 0, window.titleView.H, titleBorder, 0x0)
	window.titleView.WriteCenteredIn(0, 0, window.titleView.W, window.titleView.H, title, titleColor, titleBarBgColor).
		Update() //Refresh(window.X, window.Y, window.W, window.H)

	window.InnerY += titleBarHeight
	window.InnerH -= titleBarHeight

	if window.WindowOptions.StatusBar != nil {
		window.WindowOptions.StatusBar.ForceRefresh() // ensure full redisplay of status bar
	}
}

// SetContent sets the internal view and returns the window for chaining purpose
func (window *Window) SetContent(content Content, mx, my int) *Window {
	Debug("Set Window Content")
	window.content = content
	title := content.GetTitle()
	if title != window.Title {
		window.Title = title
		window.RefreshTitleBar()
	}
	contentViews := content.Init(window.View, window.RefreshChannel)
	window.View.Views = contentViews
	return window
}

func (window *Window) GetContentType() string {
	if window.content == nil {
		return ""
	}
	return window.content.Type()
}

// Load initialize window content and returns window for chaining purpose
func (window *Window) Load() *Window {
	Debug("Load Window")
	if window.content != nil {
		window.content.Load()
		//window.updated = true
		window.content.Print()
	}
	return window
}

func (window *Window) SetUpdated() *Window {
	window.updated = true
	return window
}

func (window *Window) KeyEvent(event *input.KeyEvent) {
	if window.content != nil {
		window.content.KeyEvent(event)
	}
}
