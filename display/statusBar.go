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
	"bufio"
	"bytes"
	"fmt"
	it8951 "github.com/peergum/IT8951-go"
	pi_sugar "github.com/peergum/pi-sugar"
	"nv/display/fonts-go"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	dateBarWidth      = 400
	batteryViewWidth  = 50
	batteryValueWidth = 100
	wifiViewWidth     = 50
)

type StatusBar struct {
	date              *View
	battery           *View
	batteryValue      *View
	wifi              *View
	W                 int
	iconRefreshTicker *time.Ticker
}

var (
	batteryIconIndex           = 0
	batteryMinIconIndex        = 0 // used for charging moving icon
	wifiIconIndex              = 0
	batteryBlinkState          = false
	wifiActive                 = true
	wifiIcon            string = iconWifi0 // default = show no icon
	done                       = false
	forceCheck                 = false
)

var batteryIcons []string = []string{
	iconBattery0,
	iconBattery1,
	iconBattery2,
	iconBattery3,
	iconBattery4,
}
var wifiIcons []string = []string{
	iconWifi0,
	iconWifi1,
	iconWifi2,
	iconWifi3,
	iconWifi4,
}
var wifiOffIcon = iconWifiOff

func (view *View) NewStatusBar(bgColor it8951.Color) (statusBar *StatusBar) {
	pos := dateBarWidth
	dateView := view.NewView(view.W-pos, 0, dateBarWidth, titleHeight, 1)
	pos += batteryViewWidth
	batteryIconView := view.NewView(view.W-pos, 0, batteryViewWidth, titleHeight, 1)
	pos += batteryValueWidth
	batteryValueView := view.NewView(view.W-pos, 0, batteryValueWidth, titleHeight, 1)
	pos += wifiViewWidth
	wifiIconView := view.NewView(view.W-pos, 0, batteryViewWidth, titleHeight, 1)
	statusBar = &StatusBar{
		date:              dateView,
		battery:           batteryIconView,
		batteryValue:      batteryValueView,
		wifi:              wifiIconView,
		W:                 dateBarWidth + batteryViewWidth + batteryValueWidth + wifiViewWidth,
		iconRefreshTicker: time.NewTicker(time.Duration(5000) * time.Millisecond),
	}

	dateView.
		SetTextArea(&fonts.UbuntuSans_Bold16pt8b, 0, 0).
		BgColor = uint16(bgColor)
	batteryIconView.Fill(0, White, Black).Update()
	batteryValueView.
		SetTextArea(&fonts.UbuntuSans_Bold16pt8b, 0, 0).
		BgColor = uint16(bgColor)
	wifiIconView.Fill(0, White, Black).Update()
	return statusBar
}

func (statusBar *StatusBar) SetWifiState(enabled bool) {
	wifiActive = enabled
	forceCheck = true
}

func (statusBar *StatusBar) Close() {
	done = true
	statusBar.iconRefreshTicker.Stop()
}

func (statusBar *StatusBar) Refresh(piSugar *pi_sugar.PiSugar) {
	select {
	case <-statusBar.iconRefreshTicker.C:
		forceCheck = true
	default:
	}
	if forceCheck {
		// check battery
		piSugar.Refresh()
		if wifiActive {
			// check wifi signal
			cmd := exec.Command("/usr/sbin/iwconfig")
			var out []byte
			var err error
			if out, err = cmd.Output(); err != nil {
				Debug("Err: %v", err)
			}
			scanner := bufio.NewScanner(bytes.NewReader(out))
			scanner.Split(bufio.ScanLines)
			for scanner.Scan() {
				text := scanner.Text()
				Debug("%s", text)
				if strings.Contains(text, "Quality=") {
					q := strings.Split(strings.Split(text, "=")[1], "/")[0]
					quality, err := strconv.Atoi(q)
					if err == nil {
						wifiIconIndex = (quality * 5) / 71 // ensure max is 4 bars (70)
					}
					// we don't use signal strength for now
					//signal := strings.Split(text, "=")[2]
					break
				}
			}
			wifiIcon = wifiIcons[wifiIconIndex]
		} else {
			wifiIcon = wifiOffIcon
		}
	}
	// define icons
	batteryCharge := piSugar.Charge()
	power := piSugar.Power()
	switch {
	case batteryCharge >= 75:
		batteryMinIconIndex = 4
		batteryBlinkState = true
	case batteryCharge >= 60:
		batteryMinIconIndex = 3
		batteryBlinkState = true
	case batteryCharge >= 45:
		batteryMinIconIndex = 2
		batteryBlinkState = true
	case batteryCharge >= 30:
		batteryMinIconIndex = 1
		batteryBlinkState = true
	case batteryCharge < 30:
		batteryMinIconIndex = 0
		batteryBlinkState = !batteryBlinkState
	default:
	}
	if power {
		batteryIconIndex = batteryMinIconIndex + (batteryIconIndex+1)%(5-batteryMinIconIndex)
		batteryBlinkState = true
	} else {
		batteryIconIndex = batteryMinIconIndex
	}
	dateBar := statusBar.date
	dateBar.Fill(0, White, Black).
		WriteCenteredIn(0,
			0,
			dateBar.W,
			dateBar.H,
			time.Now().Format("Mon Jan 2 03:04:05 PM"),
			Black, it8951.Color(Screen.BgColor)).
		Update()

	wifiIconView := statusBar.wifi
	if wifiIconBar, err := wifiIconView.LoadBitmapVCenteredAt(0, wifiIcon, 1); err == nil {
		wifiIconBar.Update()
	}
	batteryIconView := statusBar.battery
	if batteryIconBar, err := batteryIconView.LoadBitmapVCenteredAt(0, batteryIcons[batteryIconIndex], 1); err == nil {
		if !batteryBlinkState {
			batteryIconBar.Fill(0, White, White)
		}
		batteryIconBar.Update()
	}
	batteryValueView := statusBar.batteryValue
	batteryValueView.Fill(0, White, Black).
		WriteCenteredIn(0,
			0,
			batteryValueView.W,
			batteryValueView.H,
			fmt.Sprintf("%d%%", piSugar.Charge()),
			Black, it8951.Color(Screen.BgColor)).
		Update()
}
