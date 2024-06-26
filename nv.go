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

package main

import (
	"flag"
	"fmt"
	it8951 "github.com/peergum/IT8951-go"
	"log"
	"nv/content"
	"nv/display"
	"nv/display/fonts-go"
	"nv/input"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

const (
	name               = "NV-Reloaded"
	version            = 1
	release            = 0
	patch              = 0
	copyright          = "Copyright (c) 2024 Phil Hilger & Collaborators"
	defaultStatsWidth  = 900 // defines the width of the Stats Window
	defaultStatsHeight = 700 // defines the height of the Stats Window
	defaultStatsMargin = 10  // defines the distance from the stats window to the top-right embedding window
)

var (
	shouldTerminate bool // program should exit
	shouldPowerOff  bool
	shouldReboot    bool

	signalChannel  chan os.Signal = make(chan os.Signal, 1)
	debug          bool
	noWelcome      bool
	epd            bool
	Rotation       it8951.Rotate = it8951.Rotate0
	photoBorder    int           = 10
	statsWidth                   = defaultStatsWidth
	statsHeight                  = defaultStatsHeight
	statsMargin                  = defaultStatsMargin
	functionHeight               = 100

	fnWindowToggle    = false
	statsWindowToggle = false

	mainWindow  *display.Window
	fnWindow    *display.Window
	currentDoc  *content.Document
	stats       *content.Stats
	statsWindow *display.Window
	alertBox    *display.Window

	metaKey uint16

	//keyboards []*input.Keyboard
)

func init() {
	flag.BoolVar(&debug, "d", false, "debug mode")
	flag.BoolVar(&noWelcome, "nw", false, "Skip welcome screen")
}

func nv() string {
	return fmt.Sprintf("%s %d.%d.%d", name, version, release, patch)
}

func nvVersion() string {
	return fmt.Sprintf("%d.%d.%d", version, release, patch)
}

func nvVersionNum() float32 {
	return float32(version)*100 + float32(release) + float32(patch)/1000
}

func Debug(format string, args ...interface{}) {
	if debug {
		log.Printf("[NV] "+format, args...)
	}
}

func main() {
	flag.Parse()
	Debug("%s starting", nv())
	Debug("%s", copyright)

	// external signal handler
	signal.Notify(signalChannel, os.Interrupt, os.Kill)
	go signalHandler(signalChannel)

	// init
	defer it8951.Close()
	defer terminate()

	Debug("Starting serious business")

	// start searching/updating keyboards
	keyboardChannel := input.KeyboardChannel()

	keyChannel := input.KeyChannel()
	eventChannel := make(chan bool, 10)
	go input.Search(eventChannel)

	display.InitDisplay()
	display.InitScreen()

	mainWindow = display.Screen.NewWindow(0, 0, display.Screen.W, display.Screen.H, display.WindowOptions{
		Title:       "Let's Do This!",
		TitleBar:    true,
		Border:      2,
		BgColor:     display.White,
		BorderColor: display.Black,
	})

	if noWelcome {
		empty := &content.Empty{}
		mainWindow.SetContent(empty, 0, 0).Load()
	} else {
		welcome := &content.Welcome{}
		mainWindow.SetContent(welcome, 150, 150).Load()
	}

	//fnNewDocument()
	shouldTerminate = false

mainLoop:
	for !shouldTerminate {
		select {
		case kbd := <-keyboardChannel:
			if kbd == nil {
				panic("Keyboard search failed")
			}
			if kbd.File == "none" {
				Debug("No keyboard found")
				mainWindow.AlertBox("No Keyboard...", 0)
			} else {
				Debug("Keyboard %s added", kbd.Name)
				mainWindow.AlertBox(fmt.Sprintf("%s Found", kbd.Name), 1000*time.Millisecond)
			}
		case event := <-keyChannel:
			mainWindow.CancelAlert()
			//Debug("Received event: %v", event)
			if event.TypeName == "EV_KEY" && event.KeyName == "KEY_ESC" && event.Value == 1 {
				Debug("FN Toggle")
				fnToggleFnWindow()
			} else if event.TypeName == "EV_KEY" && event.SpecialKeys {
				if fnWindowToggle && metaKey != event.MetaKeys {
					functions.SetMeta(event.MetaKeys)
					fnWindow.SetUpdated().Load() //.Update()
				}
				metaKey = event.MetaKeys
			} else if event.TypeName == "EV_KEY" && strings.Contains(event.KeyName, "KEY_F") && len(event.KeyName) > 5 && event.Value == 1 {
				fnNum, err := strconv.Atoi(strings.Replace(event.KeyName, "KEY_F", "", 1))
				Debug("Fn%d", fnNum)
				if err == nil && fnNum > 0 && fnNum < 13 {
					functions.FunctionKeys[fnNum-1][metaKey].Command()
				}
			} else if mainWindow.GetContentType() == "document" && event.TypeName == "EV_KEY" && event.Value > 0 && currentDoc.Ready {
				currentDoc.Editor(event)
			}

		default:
			if shouldTerminate {
				break mainLoop
			}
		}
		if mainWindow.GetContentType() == "document" && currentDoc.Ready {
			if currentDoc.RefreshNeeded {
				currentDoc.Editor(nil)
			}
			currentDoc.ToggleCursor()
		}
	}
	Debug("We're done")
	eventChannel <- true

	win := display.Screen.NewWindow(0, 0, display.Screen.W, display.Screen.H, display.WindowOptions{
		//Title:       "NV PowerOff",
		TitleBar:    false,
		Border:      0,
		BgColor:     display.White,
		BorderColor: display.Black,
	})
	font := &fonts.IsoMetrixNF_Bold30pt8b
	if shouldPowerOff {
		win.SetTextArea(font, 0, 0).
			WriteCenteredIn(0, 0, win.W, win.H, "Powering Off...", display.Black, display.White).
			Update()

		time.Sleep(2000 * time.Millisecond)
		win.Fill(0, display.White, display.Black).
			Update()
		display.ShowLogo()
		if err := exec.Command("shutdown", "-P", "now").Run(); err != nil {
			Debug("Shutdown Error: %s", err)
		}
	} else if shouldReboot {
		win.SetTextArea(font, 0, 0).
			WriteCenteredIn(0, 0, win.W, win.H, "Rebooting...", display.Black, display.White).
			Update()

		time.Sleep(2000 * time.Millisecond)
		win.Fill(0, display.White, display.Black).
			Update()
		display.ShowLogo()
		if err := exec.Command("shutdown", "-r", "now").Run(); err != nil {
			Debug("Error Rebooting: %s", err)
		}
	} else {
		//win.SetTextArea(font, 0, 0).
		//	WriteCenteredIn(0, 0, win.W, win.H, "Terminating", display.Black, display.White).
		//	Update()

		//time.Sleep(500 * time.Millisecond)
		win.Fill(0, display.White, display.Black).
			Update()
	}
}

func terminate() {
	Debug("Terminating on request")
}

func signalHandler(c chan os.Signal) {
	for {
		select {
		case s := <-c:
			Debug("Got signal: %v", s)
			shouldTerminate = true
		default:
		}
	}
}
