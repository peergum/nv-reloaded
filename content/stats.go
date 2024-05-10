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

type Stats struct {
	*Document
	view *display.View
}

type Fields []FieldDefinition

type FieldDefinition struct {
	name   string
	format string
	value  interface{}
}

var (
	fields = make([]FieldDefinition, 10)
)

func init() {
}

func (stats *Stats) Type() string {
	return "stats"
}

func (stats *Stats) Load(view *display.View) {
	stats.view = view
	fields = Fields{
		FieldDefinition{"Title", "%s", stats.Title},
		FieldDefinition{"File", "%s", stats.Filename},
		FieldDefinition{"Characters", "%d", stats.cCount},
		FieldDefinition{"Lines", "%d", stats.lCount},
		FieldDefinition{"Words", "%d", stats.wCount},
		FieldDefinition{"Paragraphs", "%d", stats.pCount},
	}
}

func (stats *Stats) Refresh() {
}

func (stats *Stats) Save() {
}

func (stats *Stats) GetTitle() string {
	return "Stats"
}

func (stats *Stats) Print() {
	view := stats.view
	view.TextArea.BgColor = 0xe // ensure transparent BG
	labelWidth := 0
	labelFont := &fonts.CourierStd20pt8b
	font := view.TextArea.Font
	view.TextArea.SetFont(labelFont)
	for _, field := range fields {
		x0, y0 := 0, 0
		xb, _, wb, _ := view.GetTextBounds(field.name, &x0, &y0)
		if wb+xb > labelWidth {
			labelWidth = wb + xb
		}
	}
	labelWidth += 2 * view.TextArea.MarginX
	y := 0
	view.FillRectangle(labelWidth, 0, view.InnerW-labelWidth, view.InnerH, 0, 0xf, 0x0)
	view.DrawVLine(labelWidth, 0, view.InnerH, 1, 0x7)
	for _, field := range fields {
		view.TextArea.SetFont(labelFont)
		view.WriteAt(0, y, field.name, display.Black, display.Gray14)
		view.TextArea.SetFont(font)
		view.WriteAt(labelWidth+view.TextArea.MarginX, y, fmt.Sprintf(field.format, field.value), display.Black, display.Transparent)
		y += int(view.TextArea.Font.YAdvance)
	}
}
