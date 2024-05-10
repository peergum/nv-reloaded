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
	"nv/display"
	"nv/display/fonts-go"
)

type Welcome struct {
	*Document
}

var ()

func init() {
}

func (welcome *Welcome) Type() string {
	return "welcome"
}

func (welcome *Welcome) Load(view *display.View) {
	welcome.Document = &Document{
		Filename: ".welcome.txt",
		Title:    "Welcome",
	}
	view.SetTextArea(&fonts.Montserrat_Medium20pt8b, 100, 100).Update()
	welcome.Document.Load(view)
	welcome.view.Rectangle(100, 100, view.W-200, view.H-200, 1, display.Black)
}

func (welcome *Welcome) Refresh() {
}

func (welcome *Welcome) Save() {
}

func (welcome *Welcome) GetTitle() string {
	return "Welcome"
}

func (welcome *Welcome) Print() {
	Debug("Loading welcome")
	view := welcome.view
	view.FillRectangle(100, 100, view.InnerW-200, view.InnerH-200, 10, display.Gray13, display.Black)
	welcome.Document.Print()
	//view := welcome.view
	//var text string
	//y := paragraphSpacing
	//for _, paragraph := range paragraphs {
	//	text = paragraph + "\n"
	//	x0, y0 := paragraphIndent, y
	//	_, _, _, hb := view.GetTextBounds(text, &x0, &y0)
	//	if y+hb > view.InnerH {
	//		break
	//	}
	//	view.WriteAt(paragraphIndent, y, text, 0x0, display.Gray13)
	//	y += hb + paragraphSpacing
	//}
}
