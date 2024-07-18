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
	btViewWidth       = 50
	kbdViewWidth      = 90
)

type StatusBar struct {
	date              *View
	battery           *View
	batteryValue      *View
	wifi              *View
	bt                *View
	keyboard          *View
	W                 int
	heartBeatTicker   *time.Ticker
	iconRefreshTicker *time.Ticker
	piSugar           *pi_sugar.PiSugar
	done              bool
	DoneChannel       chan bool
}

var (
	batteryIconIndex            = 0
	batteryMinIconIndex         = 0 // used for charging moving icon
	wifiIconIndex               = 0
	batteryBlinkState           = false
	wifiActive                  = true
	wifiIcon             string = iconWifi0 // default = show no icon
	btActive                    = true
	kbdActive                   = true
	btIcon               string = iconBtOn // default = show no icon
	kbdIcon                     = iconNoKeyboard
	done                        = false
	forceCheck                  = true
	lastWifiIcon         string
	lastBtIcon           string
	lastKbdIcon          string
	lastBatteryIconIndex int
	lastBatteryValue     int
	lastTimestamp        string
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

func (view *View) NewStatusBar(piSugar *pi_sugar.PiSugar, bgColor it8951.Color) (statusBar *StatusBar) {
	pos := dateBarWidth
	dateView := view.NewView(view.W-pos, 0, dateBarWidth, titleHeight, 1)
	pos += batteryViewWidth
	batteryIconView := view.NewView(view.W-pos, 0, batteryViewWidth, titleHeight, 1)
	pos += batteryValueWidth
	batteryValueView := view.NewView(view.W-pos, 0, batteryValueWidth, titleHeight, 1)
	pos += wifiViewWidth
	wifiIconView := view.NewView(view.W-pos, 0, wifiViewWidth, titleHeight, 1)
	pos += btViewWidth
	btIconView := view.NewView(view.W-pos, 0, btViewWidth, titleHeight, 1)
	pos += kbdViewWidth
	kbdIconView := view.NewView(view.W-pos, 0, kbdViewWidth, titleHeight, 1)
	statusBar = &StatusBar{
		date:              dateView,
		battery:           batteryIconView,
		batteryValue:      batteryValueView,
		wifi:              wifiIconView,
		bt:                btIconView,
		keyboard:          kbdIconView,
		W:                 pos,
		heartBeatTicker:   time.NewTicker(time.Duration(1000) * time.Millisecond),
		iconRefreshTicker: time.NewTicker(time.Duration(5000) * time.Millisecond),
		piSugar:           piSugar,
		DoneChannel:       make(chan bool),
	}

	dateView.
		SetTextArea(&fonts.UbuntuSans_Bold16pt8b, 0, 0).
		BgColor = uint16(bgColor)
	batteryIconView.Fill(0, White, Black).Update()
	batteryValueView.
		SetTextArea(&fonts.UbuntuSans_Bold16pt8b, 0, 0).
		BgColor = uint16(bgColor)
	wifiIconView.Fill(0, White, Black).Update()
	btIconView.Fill(0, White, Black).Update()
	kbdIconView.Fill(0, White, Black).Update()
	return statusBar
}

func (statusBar *StatusBar) SetWifiState(enabled bool) {
	wifiActive = enabled
	forceCheck = true
}

func (statusBar *StatusBar) SetBtState(enabled bool) {
	btActive = enabled
	forceCheck = true
}

func (statusBar *StatusBar) SetKbdState(enabled bool) {
	kbdActive = enabled
	forceCheck = true
}

func (statusBar *StatusBar) Close() {
	statusBar.done = true
}

func (statusBar *StatusBar) ForceRefresh() {
	lastBtIcon = ""
	lastKbdIcon = ""
	lastWifiIcon = ""
	lastBatteryValue = -1
	lastTimestamp = ""
}

func (statusBar *StatusBar) Run() {
	statusBar.done = false
	for !statusBar.done {
		select {
		case <-statusBar.heartBeatTicker.C:
			statusBar.Refresh()
			//default:
		}
		//time.Sleep(time.Duration(500) * time.Millisecond)
	}
	statusBar.heartBeatTicker.Stop()
	statusBar.iconRefreshTicker.Stop()
	statusBar.DoneChannel <- true
	Debug("status bar done")
}

func (statusBar *StatusBar) Refresh() {
	piSugar := statusBar.piSugar
	select {
	case <-statusBar.iconRefreshTicker.C:
		Debug("force check")
		forceCheck = true
	default:
	}
	if forceCheck {
		forceCheck = false
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
				//Debug("%s", text)
				if strings.Contains(text, "Quality=") {
					q := strings.Split(strings.Split(text, "=")[1], "/")[0]
					quality, err := strconv.Atoi(q)
					if err == nil {
						wifiIconIndex = (quality * 5) / 71 // ensure max is 4 bars (70)
					}
					// we don't use signal strength for now
					//signal := strings.Split(text, "=")[2]
				}
			}
			wifiIcon = wifiIcons[wifiIconIndex]
		} else {
			wifiIcon = wifiOffIcon
		}
		//check bt status, by default off
		btActive = false
		cmd := exec.Command("/usr/bin/bluetoothctl", "show")
		var out []byte
		var err error
		if out, err = cmd.Output(); err != nil {
			Debug("Err: %v", err)
		}
		scanner := bufio.NewScanner(bytes.NewReader(out))
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			text := scanner.Text()
			//Debug("%s", text)
			if strings.Contains(text, "Powered:") &&
				strings.Contains(text, "yes") {
				btActive = true
				Debug("BT is on")
				// we don't use signal strength for now
				//signal := strings.Split(text, "=")[2]
			}
		}
	}
	if btActive {
		btIcon = iconBtOn
	} else {
		btIcon = iconBtOff
	}
	if kbdActive {
		kbdIcon = iconKeyboard
	} else {
		kbdIcon = iconNoKeyboard
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
		batteryIconIndex = (batteryIconIndex + 1) % 5
		if batteryIconIndex < batteryMinIconIndex-1 {
			batteryIconIndex = batteryMinIconIndex - 1
		}
		batteryBlinkState = true
	} else {
		batteryIconIndex = batteryMinIconIndex
	}
	now := time.Now().Format("Mon Jan 2 03:04:05 PM")
	if lastTimestamp != now {
		lastTimestamp = now
		dateBar := statusBar.date
		dateBar.Fill(0, White, Black).
			WriteCenteredIn(0,
				0,
				dateBar.W,
				dateBar.H,
				now,
				Black, it8951.Color(Screen.BgColor)).
			Update()
	}

	if lastWifiIcon != wifiIcon {
		wifiIconView := statusBar.wifi
		if wifiIconBar, err := wifiIconView.LoadBitmapVCenteredAt(0, wifiIcon, 1); err == nil {
			wifiIconBar.Update()
			lastWifiIcon = wifiIcon
		}
	}
	if lastBatteryIconIndex != batteryIconIndex {
		batteryIconView := statusBar.battery
		if batteryIconBar, err := batteryIconView.LoadBitmapVCenteredAt(0, batteryIcons[batteryIconIndex], 1); err == nil {
			lastBatteryIconIndex = batteryIconIndex
			if !batteryBlinkState {
				lastBatteryIconIndex = -1 // icon off
				batteryIconBar.Fill(0, White, White)
			}
			batteryIconBar.Update()
		}
	}
	if lastBatteryValue != piSugar.Charge() {
		lastBatteryValue = piSugar.Charge()
		batteryValueView := statusBar.batteryValue
		batteryValueView.Fill(0, White, Black).
			WriteCenteredIn(0,
				0,
				batteryValueView.W,
				batteryValueView.H,
				fmt.Sprintf("%d%%", lastBatteryValue),
				Black, it8951.Color(Screen.BgColor)).
			Update()
	}
	if lastBtIcon != btIcon {
		btIconView := statusBar.bt
		if btIconBar, err := btIconView.LoadBitmapVCenteredAt(0, btIcon, 1); err == nil {
			btIconBar.Update()
			lastBtIcon = btIcon
		}
	}
	if lastKbdIcon != kbdIcon {
		kbdIconView := statusBar.keyboard
		if kbdIconBar, err := kbdIconView.LoadBitmapVCenteredAt(0, kbdIcon, 1); err == nil {
			kbdIconBar.Update()
			lastKbdIcon = kbdIcon
		}
	}
}
