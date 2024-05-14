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

type FnPanel struct {
	FunctionKeys
	view *display.View
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

func init() {
}

func (fnPanel *FnPanel) Type() string {
	return "functions"
}

func (fnPanel *FnPanel) SetMeta(keys uint16) {
	metaKeys = keys
}

func (fnPanel *FnPanel) Load(view *display.View) {
	fnPanel.view = view
}

func (fnPanel *FnPanel) Refresh() {
}

func (fnPanel *FnPanel) Save() {
}

func (fnPanel *FnPanel) GetTitle() string {
	return "Functions"
}

func (fnPanel *FnPanel) Print() {
	view := fnPanel.view
	view.TextArea.MarginX = 0
	view.TextArea.MarginY = 0
	if len(fnPanel.FunctionKeys) == 0 {
		return
	}
	// split keys on 2 rows
	numCols := numFnKeys / numRows
	fnWidth := view.InnerW / numCols
	fnHeight := view.InnerH / numRows
	for i := 0; i < numFnKeys; i++ {
		view.FillRectangle((i%numCols)*fnWidth, (i/numCols)*fnHeight, fnWidth, fnHeight, 3, panelBgColor, display.White)
		view.FillRectangle((i%numCols)*fnWidth+4, (i/numCols)*fnHeight+4, fnHeight, fnHeight-8, 1, display.White, display.Black)
		view.TextArea.SetFont(&fonts.Montserrat_Medium12pt8b)
		view.WriteCenteredIn((i%numCols)*fnWidth+4, (i/numCols)*fnHeight+4, fnHeight, fnHeight-8, fmt.Sprintf("F%d", i+1), display.Black, display.Transparent)
		fn := (fnPanel.FunctionKeys)[i]
		if fn != nil && fn[metaKeys] != nil {
			view.TextArea.SetFont(&fonts.Montserrat_Medium16pt8b)
			x0, y0 := 0, 0
			_, _, wb, _ := view.GetTextBounds(fn[metaKeys].Label, &x0, &y0)
			Debug("Fn %d, meta: %d -> %v", i+1, metaKeys, fn[metaKeys])
			if wb >= fnWidth-fnHeight-10 {
				view.TextArea.SetFont(&fonts.Montserrat_Medium12pt8b)
			}
			view.WriteCenteredIn((i%numCols)*fnWidth+fnHeight, (i/numCols)*fnHeight, fnWidth-fnHeight, fnHeight, fn[metaKeys].Label, display.White, panelBgColor)
		}
	}
}