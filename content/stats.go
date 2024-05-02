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

func StatsNew(document *Document) display.Content {
	return &Stats{document}
}

func (stats *Stats) Load() {
	fields = Fields{
		FieldDefinition{"Title", "%s", stats.Title},
		FieldDefinition{"File", "%s", stats.Filename},
		FieldDefinition{"Words", "%d", len(stats.Words)},
		FieldDefinition{"Paragraphs", "%d", len(stats.Paragraphs)},
	}
}

func (stats *Stats) Refresh() {
}

func (stats *Stats) Save() {
}

func (stats *Stats) GetTitle() string {
	return "Stats"
}

func (stats *Stats) Print(view *display.View) {
	view.TextArea.BgColor = 0xe // ensure transparent BG
	labelWidth := 0
	var xb, yb, wb, hb int
	labelFont := &fonts.CourierStd20pt8b
	font := view.TextArea.Font
	view.TextArea.SetFont(labelFont)
	for _, field := range fields {
		view.GetTextBounds(field.name, 0, 0, &xb, &yb, &wb, &hb)
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
