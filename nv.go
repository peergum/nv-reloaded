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
	Version            string = "1.0.0"
	defaultStatsWidth         = 900 // defines the width of the Stats Window
	defaultStatsHeight        = 700 // defines the height of the Stats Window
	defaultStatsMargin        = 10  // defines the distance from the stats window to the top-right embedding window
)

var (
	shouldTerminate bool // program should exit
	shouldPowerOff  bool
	shouldReboot    bool

	signalChannel  chan os.Signal = make(chan os.Signal, 1)
	debug          bool
	epd            bool
	Rotation       it8951.Rotate = it8951.Rotate0
	photoBorder    int           = 10
	statsWidth                   = defaultStatsWidth
	statsHeight                  = defaultStatsHeight
	statsMargin                  = defaultStatsMargin
	functionHeight               = 100

	fnWindowToggle    = false
	statsWindowToggle = false

	F1 = content.FunctionKey{
		input.None:  {"Help", fnDoNothing},
		input.Shift: {"Shortcuts", fnDoNothing},
	}
	F2 = content.FunctionKey{
		input.None: {"New", fnNewDocument},
	}
	F3 = content.FunctionKey{
		input.None:  {"Load", fnLoadDocument},
		input.Shift: {"Save", fnDoNothing},
		input.Ctrl:  {"Save As", fnDoNothing},
		input.Alt:   {"Reload", fnDoNothing},
	}
	F4 = content.FunctionKey{
		input.None: {"Close", fnDoNothing},
		input.Alt:  {"Exit", fnDoNothing},
	}
	F5 = content.FunctionKey{
		input.None:  {"Top", fnDoNothing},
		input.Shift: {"Bottom", fnDoNothing},
	}
	F6 = content.FunctionKey{
		input.None: {"Setup Wifi", fnDoNothing},
	}
	F7 = content.FunctionKey{
		input.None:               {"Start Block", fnDoNothing},
		input.Shift:              {"End Block", fnDoNothing},
		input.Ctrl:               {"Cut Block", fnDoNothing},
		input.Ctrl | input.Shift: {"Ins Block", fnDoNothing},
		input.Alt:                {"Del Block", fnDoNothing},
	}
	F8 = content.FunctionKey{
		input.None:  {"Cpy Paragr.", fnDoNothing},
		input.Shift: {"Ins Paragr.", fnDoNothing},
		input.Ctrl:  {"Cut Paragr.", fnDoNothing},
		input.Alt:   {"Del Paragr.", fnDoNothing},
	}
	F9 = content.FunctionKey{
		input.None: {"Ins Picture", fnDoNothing},
		input.Ctrl: {"Cut Picture", fnDoNothing},
		input.Alt:  {"Del Picture", fnDoNothing},
	}
	F10 = content.FunctionKey{
		input.None:  {"E-mail", fnDoNothing},
		input.Shift: {"G-Drive", fnDoNothing},
		input.Ctrl:  {"Dropbox", fnDoNothing},
		input.Alt:   {"PDF", fnDoNothing},
	}
	F11 = content.FunctionKey{
		input.None: {"Stats", fnToggleStats},
	}
	F12 = content.FunctionKey{
		input.None:                           {"Sleep", fnDoNothing},
		input.Shift:                          {"Reboot", fnReboot},
		input.Ctrl:                           {"Shutdown", fnShutdown},
		input.Shift | input.Ctrl | input.Alt: {"Factory Reset", fnDoNothing},
	}

	functions = content.FnPanel{
		FunctionKeys: content.FunctionKeys{0: F1, 1: F2, 2: F3, 3: F4, 4: F5, 5: F6, 6: F7, 7: F8, 8: F9, 9: F10, 10: F11, 11: F12},
	}

	mainWindow  *display.Window
	fnWindow    *display.Window
	currentDoc  *content.Document
	stats       *content.Stats
	statsWindow *display.Window

	metaKey uint16

	//keyboards []*input.Keyboard
)

func init() {
	flag.BoolVar(&debug, "debug", false, "debug mode")
	//flag.BoolVar(&epd, "epd", false, "debug mode for EPD")
}

func Debug(format string, args ...interface{}) {
	if debug {
		log.Printf("[NV] "+format, args...)
	}
}

func main() {
	flag.Parse()
	Debug("NV %s starting", Version)

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
		Title:       "NV Starting...",
		TitleBar:    true,
		Border:      2,
		BgColor:     display.White,
		BorderColor: display.Black,
	})

	welcome := &content.Welcome{}
	mainWindow.SetContent(welcome, 150, 150).Load().Update()

	//fnNewDocument()
	shouldTerminate = false

mainLoop:
	for !shouldTerminate {
		select {
		case kbd := <-keyboardChannel:
			if kbd == nil {
				panic("Keyboard search failed")
			}
			Debug("Keyboard %s added", kbd.Name)
		case event := <-keyChannel:
			//Debug("Received event: %v", event)
			if event.TypeName == "EV_KEY" && event.KeyName == "KEY_ESC" && event.Value == 1 {
				Debug("FN Toggle")
				fnToggleFnWindow()
			} else if event.TypeName == "EV_KEY" && event.SpecialKey {
				newMetaKey := input.Metakey()
				if fnWindowToggle && metaKey != newMetaKey {
					functions.SetMeta(newMetaKey)
					fnWindow.SetUpdated().Load().Update()
				}
				metaKey = newMetaKey
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
		exec.Command("shutdown", "-P", "now").Run()
	} else if shouldReboot {
		win.SetTextArea(font, 0, 0).
			WriteCenteredIn(0, 0, win.W, win.H, "Rebooting...", display.Black, display.White).
			Update()

		time.Sleep(2000 * time.Millisecond)
		win.Fill(0, display.White, display.Black).
			Update()
		display.ShowLogo()
		exec.Command("shutdown", "-r", "now").Run()
	} else {
		win.SetTextArea(font, 0, 0).
			WriteCenteredIn(0, 0, win.W, win.H, "Terminating", display.Black, display.White).
			Update()

		time.Sleep(500 * time.Millisecond)
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
