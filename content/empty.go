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

type Empty struct {
	*display.View
}

func (empty *Empty) Init(view *display.View) (views []*display.View) {
	empty.View = view
	view.Fill(0, display.White, display.Black).
		SetTextArea(&fonts.CourierStd20pt8b, 10, 10).
		WriteCenteredIn(0, 0, view.InnerW, view.InnerH, "Ready when you are.", display.Black, display.White).
		Update()
	return append(views, view)
}

func (empty *Empty) Type() string {
	return "empty"
}

func (empty *Empty) Load() {
}

func (empty *Empty) Refresh() {
}

func (empty *Empty) Save() {
}

func (empty *Empty) GetTitle() string {
	return "Let's Do This!"
}

func (empty *Empty) Print() {

}
