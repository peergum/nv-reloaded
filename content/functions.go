/*
   content,
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

package content

import (
	"fmt"
	"nv/display"
	"nv/display/fonts-go"
	"nv/input"
)

type FnPanel struct {
	FunctionKeys
	view     *display.View
	metaKeys uint16
}

type FunctionKeys []FunctionKey

type FunctionKey map[uint16]*FunctionDefinition

type FunctionDefinition struct {
	Label   string
	Command fnCommand
}

type fnCommand func()

var (
	numFnKeys    = 12
	numRows      = 2
	panelBgColor = display.Gray8
	metaKeys     uint16
)

func (fnPanel *FnPanel) Init(view *display.View, refreshChannel chan bool) (views []*display.View) {
	fnPanel.view = view.NewView(0, 0, view.InnerW, view.InnerH, 1)
	fnPanel.view.Fill(0, display.White, display.Black).
		SetTextArea(&fonts.CourierStd20pt8b, 0, 0).
		Update()
	return append(views, fnPanel.view)
}

func (fnPanel *FnPanel) Close() {}
func (fnPanel *FnPanel) Type() string {
	return "functions"
}

func (fnPanel *FnPanel) SetMeta(keys uint16) {
	fnPanel.metaKeys = keys
}

func (fnPanel *FnPanel) Load() {}

func (fnPanel *FnPanel) Refresh() {
}

func (fnPanel *FnPanel) Save() {
}

func (fnPanel *FnPanel) GetTitle() string {
	return "Functions"
}

func (fnPanel *FnPanel) Print() {
	view := fnPanel.view
	//view.TextArea.MarginX = 0
	//view.TextArea.MarginY = 0
	if len(fnPanel.FunctionKeys) == 0 {
		return
	}
	nRows := numRows
	if view.InnerW < display.VirtualH {
		nRows *= 2
	}
	// split keys on 2 rows
	numCols := numFnKeys / nRows
	fnWidth := view.InnerW / numCols
	fnHeight := view.InnerH / nRows
	for i := 0; i < numFnKeys; i++ {
		view.FillRectangle((i%numCols)*fnWidth, (i/numCols)*fnHeight, fnWidth, fnHeight, 3 /*panelBgColor*/, display.Black, display.White)
		view.FillRectangle((i%numCols)*fnWidth+4, (i/numCols)*fnHeight+4, fnHeight, fnHeight-8, 1, display.White, display.Black)
		view.TextArea.SetFont(&fonts.Montserrat_Medium12pt8b)
		view.WriteCenteredIn((i%numCols)*fnWidth+4, (i/numCols)*fnHeight+4, fnHeight, fnHeight-8, fmt.Sprintf("F%d", i+1), display.Black, display.Transparent)
		fn := (fnPanel.FunctionKeys)[i]
		if fn != nil && fn[fnPanel.metaKeys] != nil {
			view.TextArea.SetFont(&fonts.Inter_Regular16pt8b)
			x0, y0 := 0, 0
			_, _, wb, _ := view.GetTextBounds(fn[fnPanel.metaKeys].Label, &x0, &y0)
			Debug("Fn %d, meta: %d -> %v", i+1, fnPanel.metaKeys, fn[fnPanel.metaKeys])
			if wb >= fnWidth-fnHeight-10 {
				view.TextArea.SetFont(&fonts.Inter_Regular12pt8b)
			}
			view.WriteCenteredIn((i%numCols)*fnWidth+fnHeight, (i/numCols)*fnHeight, fnWidth-fnHeight, fnHeight, fn[fnPanel.metaKeys].Label, display.White, display.Black /*panelBgColor*/)
		}
	}
	view.Update()
}

func (fnPanel *FnPanel) KeyEvent(event *input.KeyEvent) {}
