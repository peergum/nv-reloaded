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
	alertBox      *Window
	alertBoxTimer *time.Timer
)

func (window *Window) AlertBox(text string, duration time.Duration) {
	// if alert already open, cancel it first
	if alertBox != nil {
		window.CancelAlert()
	}
	alertBox = window.NewCenteredWindow(WindowOptions{
		TitleBar:    false,
		Border:      5,
		BorderColor: Black,
		BgColor:     White,
	})
	alertBox.View.
		Fill(2, White, Black).
		SetTextArea(&fonts.IsoMetrixNF_Bold24pt8b, 10, 10).
		WriteCenteredIn(0, 0, alertBox.InnerW, alertBox.InnerH, text, Black, White).
		Update()

	if duration > 0 {
		alertBoxTimer = time.AfterFunc(duration, func() {
			window.CancelAlert()
		})
	}
}

func (window *Window) CancelAlert() {
	if alertBox != nil {
		alertBox.Hide()
		alertBox = nil
		if alertBoxTimer != nil {
			alertBoxTimer.Stop()
			alertBoxTimer = nil
		}
	}
}
