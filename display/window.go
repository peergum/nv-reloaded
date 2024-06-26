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
)

type Content interface {
	Init(*View) []*View
	GetTitle() string
	Load()
	Refresh()
	Save()
	Print()
	Type() string
}

type WindowOptions struct {
	Title       string
	TitleBar    bool
	Border      int
	BgColor     it8951.Color
	BorderColor it8951.Color
}

type Window struct {
	*View
	WindowOptions
	titleView *View
	Views     []*View
	updated   bool
	visible   bool
	parent    *View
}

type ScreenView struct {
	*View
	windows []Window
}

const (
	titleBorder     = 2
	titleHeight     = 48
	contentBorder   = 1
	titleBarBgColor = Gray11
)

var (
	Screen ScreenView
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

func (view *View) NewCenteredWindow(options WindowOptions) *Window {
	return view.NewWindow((view.InnerW-800)/2, (view.InnerH-300)/2, 800, 300, options)
}

func (view *View) NewWindow(x, y, w, h int, options WindowOptions) *Window {
	Debug("Create a New Window (%d,%d,%d,%d,%s)", x, y, w, h, options.Title)
	var window Window
	window.parent = view
	window.WindowOptions = options
	//if window.View == nil {
	window.View = view.NewView(x, y, w, h, 4)
	//}
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
	if window.updated {
		//window.View.Rectangle(window.X, window.Y, window.W, window.H, window.Border, window.BorderColor).Update()
		window.View.Fill(window.Border, window.WindowOptions.BgColor, window.BorderColor).Update()
	}
	window.InnerX = window.X + window.Border
	window.InnerY = window.Y + window.Border
	window.InnerW = window.W - 2*window.Border
	window.InnerH = window.H - 2*window.Border
	if window.updated {
		if window.WindowOptions.TitleBar {
			titleBarHeight := 0
			titleBarHeight = titleHeight + titleBorder
			window.titleView = window.NewView(0, 0, window.InnerW, titleBarHeight, 4).
				SetTextArea(&fonts.SF_Compact_Display_Black20pt8b, 0, 0)
			window.InnerY += titleBarHeight
			window.InnerH -= titleBarHeight
			window.RefreshTitleBar()
		}
		it8951.WaitForDisplayReady()
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
	view := Screen.NewView(window.X, window.Y, window.W, window.H, 4)

	for _, parentView := range window.parent.Views {
		////// copy content from parent
		for y := 0; y < window.H; y++ {
			for x := 0; x < window.W; x++ {
				if window.X+x >= parentView.X && window.X+x < parentView.X+parentView.W &&
					window.Y+y >= parentView.Y && window.Y+y < parentView.Y+parentView.H {
					view.buffer.writePixel(view.X+x, view.Y+y, parentView.buffer.readPixel(window.X+x, window.Y+y))
				}
			}
		}
	}

	view.Refresh(view.X, view.Y, view.W, view.H)

}

func (window *Window) Close() {
	window.Hide()
	window.View = nil
}

func (window *Window) RefreshTitleBar() {
	if !window.WindowOptions.TitleBar {
		return
	}
	window.titleView.Fill(0, titleBarBgColor, Black)
	window.titleView.
		DrawHLine(0, window.titleView.H-titleBorder, window.titleView.W, titleBorder, 0x0)
	window.titleView.WriteVCenteredAt(20, window.Title, 0x0, titleBarBgColor).
		Update()
}

// SetContent sets the internal view and returns the window for chaining purpose
func (window *Window) SetContent(content Content, mx, my int) *Window {
	Debug("Set Window Content")
	window.content = content
	window.Title = content.GetTitle()
	window.RefreshTitleBar()
	contentViews := content.Init(window.View)
	window.View.Views = append(window.View.Views, contentViews...)
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
