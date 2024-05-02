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
	"nv/input"
	"os"
	"os/signal"
	"strconv"
	"strings"
)

const (
	Version            string = "1.0.0"
	defaultStatsWidth         = 900 // defines the width of the Stats Window
	defaultStatsHeight        = 700 // defines the height of the Stats Window
	defaultStatsMargin        = 10  // defines the distance from the stats window to the top-right embedding window
)

var (
	shouldTerminate bool           // program should exit
	signalChannel   chan os.Signal = make(chan os.Signal, 1)
	debug           bool
	epd             bool
	Rotation        it8951.Rotate = it8951.Rotate0
	photoBorder     int           = 10
	statsWidth                    = defaultStatsWidth
	statsHeight                   = defaultStatsHeight
	statsMargin                   = defaultStatsMargin
	functionHeight                = 100

	fnWindowToggle    = false
	statsWindowToggle = false

	F1 = content.FunctionKey{
		content.None:  {"Help", fnDoNothing},
		content.Shift: {"Shortcuts", fnDoNothing},
	}
	F2 = content.FunctionKey{
		content.None: {"New", fnNewDocument},
	}
	F3 = content.FunctionKey{
		content.None:  {"Load", fnDoNothing},
		content.Shift: {"Save", fnDoNothing},
		content.Ctrl:  {"Save As", fnDoNothing},
		content.Alt:   {"Reload", fnDoNothing},
	}
	F4 = content.FunctionKey{
		content.None: {"Close", fnDoNothing},
		content.Alt:  {"Exit", fnDoNothing},
	}
	F5 = content.FunctionKey{
		content.None:  {"Top", fnDoNothing},
		content.Shift: {"Bottom", fnDoNothing},
	}
	F6 = content.FunctionKey{
		content.None: {"Setup Wifi", fnDoNothing},
	}
	F7 = content.FunctionKey{
		content.None:                 {"Start Block", fnDoNothing},
		content.Shift:                {"End Block", fnDoNothing},
		content.Ctrl:                 {"Cut Block", fnDoNothing},
		content.Ctrl | content.Shift: {"Ins Block", fnDoNothing},
		content.Alt:                  {"Del Block", fnDoNothing},
	}
	F8 = content.FunctionKey{
		content.None:  {"Cpy Paragr.", fnDoNothing},
		content.Shift: {"Ins Paragr.", fnDoNothing},
		content.Ctrl:  {"Cut Paragr.", fnDoNothing},
		content.Alt:   {"Del Paragr.", fnDoNothing},
	}
	F9 = content.FunctionKey{
		content.None: {"Ins Picture", fnDoNothing},
		content.Ctrl: {"Cut Picture", fnDoNothing},
		content.Alt:  {"Del Picture", fnDoNothing},
	}
	F10 = content.FunctionKey{
		content.None:  {"E-mail", fnDoNothing},
		content.Shift: {"G-Drive", fnDoNothing},
		content.Ctrl:  {"Dropbox", fnDoNothing},
		content.Alt:   {"PDF", fnDoNothing},
	}
	F11 = content.FunctionKey{
		content.None: {"Stats", fnToggleStats},
	}
	F12 = content.FunctionKey{
		content.None:  {"Sleep", fnDoNothing},
		content.Shift: {"Reboot", fnDoNothing},
		content.Ctrl:  {"Shutdown", fnDoNothing},
		content.Shift | content.Ctrl | content.Alt: {"Factory Reset", fnDoNothing},
	}

	functions = content.FnPanel{
		0: F1, 1: F2, 2: F3, 3: F4, 4: F5, 5: F6, 6: F7, 7: F8, 8: F9, 9: F10, 10: F11, 11: F12,
	}

	mainWindow  *display.Window
	fnWindow    *display.Window
	currentDoc  *content.Document
	stats       *content.Stats
	statsWindow *display.Window

	metaKey uint8
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

	keyboards, err := input.Search()
	if err != nil {
		panic(err)
	}

	keyChannel := input.KeyChannel()
	eventChannel := make(chan string, 10)

	for i := range keyboards {
		Debug("%s", keyboards[i].Name)
		go input.ReadKeyboard(keyboards[i], eventChannel)
	}

	display.InitDisplay()
	display.InitScreen()

	mainWindow = display.Screen.NewWindow(0, 0, display.Screen.W, display.Screen.H, display.WindowOptions{
		Title:       "Welcome to NV-Reloaded",
		TitleBar:    true,
		Border:      2,
		BgColor:     display.White,
		BorderColor: display.Black,
	})

	//currentDoc = &content.Document{
	//	Filename: "abc.txt",
	//	Title:    "My ABC",
	//}
	//
	//mainWindow.SetContent(currentDoc, 10, 10).Load().Update()

	//statsWindow.Hide()
	//time.Sleep(time.Duration(1) * time.Second)
	//fnWindow.Hide()
	//time.Sleep(time.Duration(1) * time.Second)
	//statsWindow.Show()

	shouldTerminate = false

mainLoop:
	for !shouldTerminate {
		select {
		case event := <-keyChannel:
			//Debug("Received event: %v", event)
			if event.TypeName == "EV_KEY" && event.KeyName == "KEY_ESC" && event.Value == 1 {
				Debug("FN Toggle")
				fnToggleFnWindow()
			} else if event.TypeName == "EV_KEY" && strings.Contains(event.KeyName, "KEY_F") && event.Value == 1 {
				fnNum, err := strconv.Atoi(strings.Replace(event.KeyName, "KEY_F", "", 1))
				Debug("Fn%d", fnNum)
				if err == nil && fnNum > 0 && fnNum < 13 {
					functions[fnNum-1][metaKey].Command()
				}
			}
		default:
			if shouldTerminate {
				break mainLoop
			}
		}
	}
	Debug("We're done")
	eventChannel <- "done"

}

func fnDoNothing() {}

func fnToggleFnWindow() {
	fnWindowToggle = !fnWindowToggle
	if fnWindowToggle {
		if fnWindow == nil {
			fnWindow = mainWindow.NewWindow(0, mainWindow.InnerH-functionHeight, mainWindow.W, functionHeight, display.WindowOptions{
				Title:       "Functions",
				TitleBar:    false,
				Border:      1,
				BorderColor: display.Black,
				BgColor:     display.White,
			})
			functions.SetMeta(content.None)
			fnWindow.SetContent(&functions, 0, 0).
				Load().
				Update()
		} else {
			fnWindow.Show()
		}
	} else {
		fnWindow.Hide()
	}
}

func fnToggleStats() {
	statsWindowToggle = !statsWindowToggle

	if statsWindowToggle && mainWindow.View != nil {
		if statsWindow == nil {
			stats = &content.Stats{
				Document: currentDoc,
			}

			x := (mainWindow.InnerW - statsWidth) / 2
			y := (mainWindow.InnerH - statsHeight) / 2
			statsWindow = mainWindow.NewWindow(x, y, statsWidth, statsHeight, display.WindowOptions{
				Title:       "Stats",
				TitleBar:    true,
				Border:      5,
				BorderColor: display.Black,
				BgColor:     display.Gray14,
			})

			statsWindow.SetContent(stats, 10, 10).
				Load().
				Update()

		} else {
			statsWindow.Show()
		}
	} else {
		statsWindow.Hide()
	}
}

func fnNewDocument() {
	currentDoc = &content.Document{
		Filename: "",
		Title:    "New Document",
	}

	mainWindow.SetContent(currentDoc, 10, 10).Load().Update()

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
