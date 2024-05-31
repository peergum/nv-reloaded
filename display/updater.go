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
	it8951 "github.com/peergum/IT8951-go"
)

type UpdateType int

type ScreenUpdate struct {
	*View
	X, Y, W, H   int
	UpdateType   UpdateType
	CopyFrom     *View
	SrcX, SrcY   int
	Transparency float64
	BgColor      uint16
}

const (
	Refresh UpdateType = iota
	Copy
	CopyTransparent
)

var (
	screenUpdateChannel chan *ScreenUpdate = make(chan *ScreenUpdate, 20)
)

func UpdateScreen(update *ScreenUpdate) {
	if update != nil {
		update.View.buffer.Lock() // lock buffer for writing
	}
	screenUpdateChannel <- update
}

func ScreenUpdater() {
	done := false
	for !done {
		select {
		case update := <-screenUpdateChannel:
			// check if we need to terminate
			if update == nil {
				Debug("ScreenUpdater requested to terminate")
				done = true
				break
			}
			Debug("ScreenUpdater received update")
			view := update.View
			x, y, w, h := update.X, update.Y, update.W, update.H
			switch update.UpdateType {
			case Refresh:
				imageInfo := it8951.LoadImgInfo{
					EndianType:       it8951.LoadImgLittleEndian,
					PixelFormat:      pixelFormat(view.buffer.bpp),
					Rotate:           it8951.Rotate0,
					SourceBufferAddr: view.buffer.data,
					TargetMemAddr:    DeviceInfo.TargetAddress(),
				}

				areaInfo := it8951.AreaImgInfo{
					X: uint16(view.X),
					Y: uint16(view.Y),
					W: uint16(view.W),
					H: uint16(view.H),
				}
				imageInfo.HostAreaPackedPixelWrite(areaInfo, view.buffer.bpp, true)
				mode := it8951.GC16Mode
				if view.buffer.bpp == 1 {
					mode = it8951.A2Mode
					it8951.Display1bpp(uint16(x), uint16(y), uint16(w), uint16(h), mode, DeviceInfo.TargetAddress(), 0xff, 0x00)
				} else {
					it8951.DisplayArea(uint16(x), uint16(y), uint16(w), uint16(h), mode)
					it8951.WaitForDisplayReady()
				}
			default:
			}
			update.View.buffer.Unlock() // unlock when we're done
		}
	}
}
