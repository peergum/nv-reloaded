/*
   nv, NV-Reloaded an e-ink writing device
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

const (
	LogoFilename = "images/nv-reloaded.bmp"
)

var (
	DeviceInfo it8951.DevInfo
)

func InitDisplay() {
	DeviceInfo = it8951.Init(2130)
	Debug(DeviceInfo.String())

	DeviceInfo.ClearRefresh(uint32(DeviceInfo.MemAddrL)+uint32(DeviceInfo.MemAddrH)<<16, it8951.InitMode)
	it8951.WaitForDisplayReady()
}

func InitScreen() {
	Screen = ScreenView{
		View: &View{
			X:      0,
			Y:      0,
			W:      int(DeviceInfo.PanelW),
			H:      int(DeviceInfo.PanelH),
			InnerX: 0,
			InnerY: 0,
			InnerW: int(DeviceInfo.PanelW),
			InnerH: int(DeviceInfo.PanelH),
			buffer: Buffer{
				X:    0,
				Y:    0,
				ww:   int(DeviceInfo.PanelW / 2),
				wh:   int(DeviceInfo.PanelH),
				data: make(it8951.DataBuffer, int(DeviceInfo.PanelW)*int(DeviceInfo.PanelH)/2),
			},
		},
		windows: make([]Window, 3),
	}
	ShowLogo()
}

func ShowLogo() {
	bpp := 4
	logo, err := LoadBitmap(LogoFilename, bpp)
	if err == nil {
		Screen.DrawCentered(logo, bpp)
	}
	it8951.WaitForDisplayReady()
	//for i := 0; i < 2; i++ {
	//	//Debug("%d...", i)
	//	time.Sleep(time.Duration(1) * time.Second)
	//}
}
