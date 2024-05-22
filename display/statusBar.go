/*
   display,
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

type StatusBar struct {
	*View
}

func (view *View) NewStatusBar() (statusBar *StatusBar) {
	statusBar = &StatusBar{
		View: view.NewView(view.W/2, 0, view.W/2, titleHeight, 4),
	}
	statusBar.Fill(1, Gray11, White).Update()
	return
}

func (statusBar *StatusBar) Refresh() {
	Debug("Refreshing StatusBar")
}
