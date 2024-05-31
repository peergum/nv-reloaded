/*
   config,
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
	"bufio"
	"bytes"
	"fmt"
	it8951 "github.com/peergum/IT8951-go"
	"nv/display"
	"nv/display/fonts-go"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	maxSSID            = 10
	wifiScrollBarWidth = 20
)

type WifiConfig struct {
	SSID       string
	Passphrase string
}

type WifiPanel struct {
	view *display.View
	*Panel
}

type SSIDList []SSID

type SSID struct {
	Name        string
	AccessPoint string
	Quality     int
	Signal      string
	Encryption  bool
	Frequency   float64
	Channel     int
	Active      bool
}

var (
	ssid      string
	pass      string
	ssidList  SSIDList
	firstSSID = 0
)

func (list *SSIDList) Len() int {
	return len(*list)
}

func (list *SSIDList) Less(i, j int) bool {
	// return Active first
	if (*list)[i].Active != (*list)[j].Active {
		return (*list)[i].Active
	}
	// return first higher quality
	if (*list)[i].Quality != (*list)[j].Quality {
		return (*list)[i].Quality > (*list)[j].Quality
	}
	return (*list)[i].Name < (*list)[j].Name
}

func (list *SSIDList) Swap(i, j int) {
	(*list)[i], (*list)[j] = (*list)[j], (*list)[i]
}

func (list *SSIDList) setActiveAP(AP string) {
	for i, v := range *list {
		//Debug("Comparing [%s] and [%s]", v.AccessPoint, AP)
		if v.AccessPoint == AP {
			(*list)[i].Active = true
		} else {
			(*list)[i].Active = false
		}
		//Debug("%s", *list)
	}
}

func (list *SSIDList) String() (res string) {
	for _, ssid := range *list {
		res += fmt.Sprintf("%s [%s] (%.1f [%d]) [Q=%d/70] %s [enc=%t][act=%t]\n",
			ssid.Name,
			ssid.AccessPoint,
			ssid.Frequency,
			ssid.Channel,
			ssid.Quality,
			ssid.Signal,
			ssid.Encryption,
			ssid.Active,
		)
	}
	return res
}

func (list *SSIDList) GetValues(doneChannel chan<- bool) {
	cmd := exec.Command("/usr/sbin/iwlist", "wlan0", "scan")
	var out []byte
	var err error
	if out, err = cmd.Output(); err != nil {
		return
	}
	scanner := bufio.NewScanner(bytes.NewReader(out))
	scanner.Split(bufio.ScanLines)
	ssid := SSID{}
	for scanner.Scan() {
		//Debug(scanner.Text())
		text := scanner.Text()
		if strings.Contains(text, "ESSID:") {
			ssid.Name = strings.Split(text, "\"")[1]
		}
		if strings.Contains(text, "Quality=") {
			ssid.Quality, err = strconv.Atoi(strings.Split(strings.Split(text, "=")[1], "/")[0])
			ssid.Signal = strings.Split(text, "=")[2]
		}
		if strings.Contains(text, "Encryption key:") {
			ssid.Encryption = strings.Split(text, ":")[1] == "on"
		}
		if strings.Contains(text, "Frequency:") {
			re := regexp.MustCompile("[ :)(]+")
			split := re.Split(text, -1)
			ssid.Frequency, err = strconv.ParseFloat(split[2], 64)
			ssid.Channel = -1
			if len(split) > 5 {
				ssid.Channel, err = strconv.Atoi(split[5])
			}
		}
		if strings.Contains(text, "Address:") {
			if len(ssid.Name) > 0 {
				*list = append(*list, ssid)
			}
			ssid = SSID{
				AccessPoint: strings.Fields(text)[4],
			}
		}
	}
	if len(ssid.Name) > 0 {
		*list = append(*list, ssid)
	}
	// check active SSID
	cmd = exec.Command("/usr/sbin/iwconfig")
	if out, err = cmd.Output(); err != nil {
		return
	}
	scanner = bufio.NewScanner(bytes.NewReader(out))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.Contains(text, "Access Point:") {
			ap := strings.Fields(text)[5]
			list.setActiveAP(ap)
			break
		}
	}
	sort.Sort(list)
	doneChannel <- true
	//Debug("%s", list)
}

func (wifiPanel *WifiPanel) Init(view *display.View) (views []*display.View) {
	wifiPanel.view = view.NewView(0, 0, view.InnerW, view.InnerH, 4)
	wifiPanel.view.Fill(1, display.White, display.Black).
		SetTextArea(&fonts.IsoMetrixNF20pt8b, 20, 20).
		Update()
	doneChannel := make(chan bool)
	spinner := wifiPanel.view.NewSpinner("Scanning WiFi")
	go spinner.Run(doneChannel)
	ssidList.GetValues(doneChannel)
	<-spinner.Done // wait for spinner
	wifiPanel.view.Update()
	wifiPanel.Panel = &Panel{
		InputFields: InputFields{
			&InputField{
				"SSID",
				"%s",
				SelectField,
				ssid,
				&ssidList,
			},
			&InputField{
				"Passphrase",
				"%s",
				TextField,
				pass,
				nil,
			},
		},
	}

	return append(views, wifiPanel.view)
}

func (wifiPanel *WifiPanel) GetTitle() string {
	return "WiFi Setup"
}
func (wifiPanel *WifiPanel) Load() {

}
func (wifiPanel *WifiPanel) Save()    {}
func (wifiPanel *WifiPanel) Refresh() {}
func (wifiPanel *WifiPanel) Type() string {
	return "wifi"
}

func (wifiPanel *WifiPanel) Print() {
	view := wifiPanel.view
	labelWidth := 0
	labelFont := &fonts.IsoMetrixNF_Bold20pt8b
	font := view.TextArea.Font
	lineHeight := max(int(font.YAdvance), int(labelFont.YAdvance))
	view.TextArea.SetFont(labelFont)
	for _, field := range wifiPanel.InputFields {
		x0, y0 := 0, 0
		xb, _, wb, _ := view.GetTextBounds(field.name, &x0, &y0)
		if wb+xb > labelWidth {
			labelWidth = wb + xb
		}
	}
	mx := view.TextArea.MarginX
	my := view.TextArea.MarginY
	labelWidth += 2 * mx
	view.FillRectangle(labelWidth, 0, view.InnerW-labelWidth, view.InnerH, 0, display.White, display.Black)
	view.DrawVLine(labelWidth, 0, view.InnerH, 1, 0x7)
	dataWidth := view.InnerW - labelWidth - 2*mx
	y := 0
	for _, field := range wifiPanel.InputFields {
		view.TextArea.SetFont(labelFont)
		view.WriteAt(0, y, field.name, display.Black, it8951.Color(view.BgColor))
		view.TextArea.SetFont(font)
		value := fmt.Sprintf(field.format, field.value)
		switch field.fieldType {
		case TextField:
			view.RoundedRectangle(labelWidth+mx, y+my-1, dataWidth, lineHeight-2, 1, display.Black, 10)
			view.WriteAt(labelWidth+mx+1, y+my, strings.Repeat("●", len(value)), display.Black, display.White)
			y += lineHeight
		case PasswordField:
			view.RoundedRectangle(labelWidth+mx, y+my-1, dataWidth, lineHeight-2, 1, display.Black, 10)
			y += lineHeight
		case SelectField:
			numOptions := min(maxSSID, len(ssidList))
			view.RoundedRectangle(labelWidth+mx, y+my-1, dataWidth-mx-wifiScrollBarWidth, numOptions*lineHeight-2, 1, display.Black, 10)
			for i := 0; firstSSID+i < maxSSID; i++ {
				ssid := ssidList[firstSSID+i]
				value := fmt.Sprintf("%.10s [%.1fG, ch.%02d] - %d/5",
					ssid.Name,
					ssid.Frequency,
					ssid.Channel,
					(ssid.Quality*5)/70)
				textColor := display.Black
				bgColor := display.White
				if ssid.Active {
					textColor = display.White
					bgColor = display.Gray8
					view.FillRoundedRectangle(labelWidth+mx+1, y+i*lineHeight+my, dataWidth-mx-wifiScrollBarWidth-2, lineHeight-4, 0, bgColor, textColor, 10)
				}
				view.WriteAt(labelWidth+mx, y+i*lineHeight+3, value, textColor, bgColor)
			}
			sbFull := len(ssidList)
			sbPixels := numOptions * lineHeight
			if firstSSID > 0 {
				view.FillRectangle(labelWidth+dataWidth, y+my, wifiScrollBarWidth-2, sbPixels*firstSSID/sbFull, 0, display.White, display.Black)
			}
			view.FillRectangle(labelWidth+dataWidth, y+my+sbPixels*firstSSID/sbFull, wifiScrollBarWidth-2, sbPixels*numOptions/sbFull, 0, display.Gray7, display.Black)
			if firstSSID+numOptions < sbFull {
				view.FillRectangle(labelWidth+dataWidth, y+my+sbPixels*(firstSSID+numOptions)/sbFull, wifiScrollBarWidth-2, sbPixels*(sbFull-firstSSID-numOptions)/sbFull, 0, display.White, display.Black)
			}
			y += numOptions * lineHeight
		default:
			numOptions := 1
			view.RoundedRectangle(labelWidth+mx, y+my-1, dataWidth, numOptions*lineHeight-2, 1, display.Black, 10)
			y += numOptions * lineHeight
		}
		y += lineHeight / 2
	}
	view.Update()
}
