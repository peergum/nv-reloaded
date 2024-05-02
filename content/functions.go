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
)

type FnPanel []FunctionKey

type FunctionKey map[uint8]*FunctionDefinition

type FunctionDefinition struct {
	Label   string
	Command fnCommand
}

type fnCommand func()

const (
	Shift uint8 = 1 << iota
	Ctrl
	Alt
	Cmd

	None uint8 = 0
)

var (
	numFnKeys    = 12
	numRows      = 2
	panelBgColor = display.Gray8
	metaKeys     uint8
)

func init() {
}

func (fnPanel *FnPanel) SetMeta(keys uint8) {
	metaKeys = keys
}

func (fnPanel *FnPanel) Load() {}

func (fnPanel *FnPanel) Refresh() {
}

func (fnPanel *FnPanel) Save() {
}

func (fnPanel *FnPanel) GetTitle() string {
	return "Functions"
}

func (fnPanel *FnPanel) Print(view *display.View) {
	view.TextArea.MarginX = 0
	view.TextArea.MarginY = 0
	if len(*fnPanel) == 0 {
		return
	}
	// split keys on 2 rows
	numCols := numFnKeys / numRows
	fnWidth := view.InnerW / numCols
	fnHeight := view.InnerH / numRows
	var xb, yb, wb, hb int
	for i := 0; i < numFnKeys; i++ {
		view.FillRectangle((i%numCols)*fnWidth, (i/numCols)*fnHeight, fnWidth, fnHeight, 3, panelBgColor, display.White)
		view.FillRectangle((i%numCols)*fnWidth+4, (i/numCols)*fnHeight+4, fnHeight, fnHeight-8, 1, display.White, display.Black)
		view.TextArea.SetFont(&fonts.Montserrat_Medium12pt8b)
		view.WriteCenteredIn((i%numCols)*fnWidth+4, (i/numCols)*fnHeight+4, fnHeight, fnHeight-8, fmt.Sprintf("F%d", i+1), display.Black, display.Transparent)
		fn := (*fnPanel)[i]
		if fn != nil && fn[metaKeys] != nil {
			view.TextArea.SetFont(&fonts.Montserrat_Medium16pt8b)
			view.GetTextBounds(fn[metaKeys].Label, 0, 0, &xb, &yb, &wb, &hb)
			Debug("Fn %d, meta: %d -> %v", i+1, metaKeys, fn[metaKeys])
			if wb >= fnWidth-fnHeight-10 {
				view.TextArea.SetFont(&fonts.Montserrat_Medium12pt8b)
			}
			view.WriteCenteredIn((i%numCols)*fnWidth+fnHeight, (i/numCols)*fnHeight, fnWidth-fnHeight, fnHeight, fn[metaKeys].Label, display.White, panelBgColor)
		}
	}
}
