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
	rand2 "math/rand"
)

const (
	LogoFilename   = "images/nv-reloaded-logo.bmp"
	iconWifi0      = "images/icons/wifi_0.bmp"
	iconWifi1      = "images/icons/wifi_1.bmp"
	iconWifi2      = "images/icons/wifi_2.bmp"
	iconWifi3      = "images/icons/wifi_3.bmp"
	iconWifi4      = "images/icons/wifi_4.bmp"
	iconWifiOff    = "images/icons/wifi_off.bmp"
	iconBattery0   = "images/icons/battery_0.bmp"
	iconBattery1   = "images/icons/battery_1.bmp"
	iconBattery2   = "images/icons/battery_2.bmp"
	iconBattery3   = "images/icons/battery_3.bmp"
	iconBattery4   = "images/icons/battery_4.bmp"
	iconCharging   = "images/icons/battery_chg.bmp"
	iconBtOn       = "images/icons/bt_on.bmp"
	iconBtOff      = "images/icons/bt_off.bmp"
	iconKeyboard   = "images/icons/keyboard.bmp"
	iconNoKeyboard = "images/icons/no_keyboard.bmp"
)

var (
	DeviceInfo it8951.DevInfo
	gallery    = []string{
		"images/machu-picchu.bmp",
		"images/teotihuacan.bmp",
	}
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
			//buffer: Buffer{
			//	X:    0,
			//	Y:    0,
			//	ww:   int(DeviceInfo.PanelW / 2),
			//	wh:   int(DeviceInfo.PanelH),
			//	data: make(it8951.DataBuffer, int(DeviceInfo.PanelW)*int(DeviceInfo.PanelH)/2),
			//},
		},
		Windows: make([]*Window, 0, 10),
	}
	Screen.setBuffer(1)
	if !noLogo {
		ShowLogo()
	}
	// the only view of Screen is itself
	// beware of possible recursions here!
	Screen.Views = make([]*View, 0)
	Screen.Views = append(Screen.Views, Screen.View)
}

func ShowLogo() {
	bpp := 8
	logoView, err := Screen.LoadBitmapCentered(LogoFilename, bpp)
	if err == nil {
		logoView.Update()
	}
}

func ShowGallery() {
	bpp := 8
	//for _, pic := range gallery {
	//	galleryView, err := Screen.LoadBitmapCentered(pic, bpp)
	//	if err == nil {
	//		galleryView.Update()
	//	}
	//}
	galleryView, err := Screen.LoadBitmapCentered(gallery[rand2.Int()%len(gallery)], bpp)
	if err == nil {
		galleryView.Update()
	}
}
