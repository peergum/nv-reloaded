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

import (
	"nv/display/fonts-go"
	"time"
)

var (
	waitBox      *Window
	waitBoxTimer *time.Timer
)

func (window *Window) WaitBox(text string, duration time.Duration) {
	// if wait already open, cancel it first
	if waitBox != nil {
		window.CancelWait()
	}
	waitBox = window.NewCenteredWindow(WindowOptions{
		TitleBar:    false,
		Border:      5,
		BorderColor: Black,
		BgColor:     White,
	})
	waitBox.View.
		Fill(2, White, Black).
		SetTextArea(&fonts.IsoMetrixNF_Bold24pt8b, 10, 10).
		WriteCenteredIn(0, 0, waitBox.InnerW, waitBox.InnerH, text, Black, White).
		Update()

	if duration > 0 {
		waitBoxTimer = time.AfterFunc(duration, func() {
			window.CancelWait()
		})
	}
}

func (window *Window) CancelWait() {
	if waitBox != nil {
		waitBox.Hide()
		waitBox = nil
		if waitBoxTimer != nil {
			waitBoxTimer.Stop()
			waitBoxTimer = nil
		}
	}
}
