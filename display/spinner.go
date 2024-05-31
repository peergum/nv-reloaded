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

const (
	repeat        = 20 // total number of stripes for width
	slope         = -1 // negative = slanted right
	ratio         = 2  // ratio blank/stripe
	border        = 1  // frame border
	spinnerBorder = 2  // spinner rectangle border
	width         = 300
	height        = 20
	margin        = 40
)

type Spinner struct {
	*View
	parent *View
	pos    int
	Done   chan bool
}

func (view *View) NewSpinner(text string) *Spinner {
	spinner := &Spinner{
		view.NewView((view.InnerW-width-2*margin)/2, (view.InnerH-height-50-2*margin)/2, width+2*margin, height+40+2*margin, 1),
		view,
		0,
		make(chan bool),
	}
	spinner.Fill(border, White, Black).
		SetTextArea(&fonts.UbuntuSans_Bold16pt8b, 0, 0).
		WriteHCenteredAt(margin+height+20, text, Black, White).
		Update()
	spinner.InnerX += spinnerBorder + margin
	spinner.InnerY += spinnerBorder + margin
	spinner.InnerW -= 2 * (spinnerBorder + margin)
	spinner.InnerH -= 2 * (spinnerBorder + margin)
	spinner.refresh() // show initial bar
	return spinner
}

// Wait starts the spinner and wait for a done signal to stop
func (spinner *Spinner) Run(doneChannel <-chan bool) {
	spinnerTicker := time.NewTicker(time.Duration(100) * time.Millisecond)
	done := false
	for !done {
		select {
		case <-doneChannel:
			done = true
		case <-spinnerTicker.C:
			spinner.refresh()
		default:
		}
	}
	spinnerTicker.Stop()
	spinner.Done <- true
}

func (spinner *Spinner) refresh() {
	spinner.pos += 10
	var color uint16
	spinner.CopyPixels(0, 0, spinner.parent, spinner.X, spinner.Y, spinner.W, spinner.H)
	for y := 0; y < height-2*spinnerBorder; y++ {
		i := y / 3
		for x := 0; x < spinner.InnerW; x++ {
			color = 0xff
			if (x+i+spinner.pos)%(width/repeat) < width/repeat/ratio {
				color = 0x00
			}
			spinner.buffer.writePixel(spinner.InnerX+x, spinner.InnerY+y, color)
		}
	}
	spinner.View.Update()

}
