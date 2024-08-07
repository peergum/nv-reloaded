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
	"github.com/peergum/pi-sugar"
	"io"
	"log"
	"net/http"
	"nv/content"
	"nv/display"
	"nv/display/fonts-go"
	"nv/input"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
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
	wifiWidth          = 800
	wifiHeight         = 600
	screenBgColor      = display.White
	// TODO: screen orientation is not functional yet: keep value at 0º
	defaultOrientationCCWDegrees = 0     // orientation in degrees counterclockwise (0,90,180,270)
	defaultVcom                  = -2.13 // for 6" 1448x1072 HD Waveshare display
)

var (
	shouldTerminate       bool // program should exit
	shouldRestart         bool // program should restart
	shouldPowerOff        bool
	shouldReboot          bool
	freshStart            bool           = true
	signalChannel         chan os.Signal = make(chan os.Signal, 1)
	terminateChannel      chan bool      = make(chan bool, 1)
	debug                 bool
	noWelcome             bool
	vcom                  float64
	orientationCCWDegrees int
	photoBorder           int = 10
	statsWidth                = defaultStatsWidth
	statsHeight               = defaultStatsHeight
	statsMargin               = defaultStatsMargin
	functionHeight            = 100

	fnWindowToggle    = false
	statsWindowToggle = false
	wifiWindowToggle  = false

	mainWindow  *display.Window
	fnWindow    *display.Window
	statsWindow *display.Window
	wifiWindow  *display.Window
	alertBox    *display.Window
	statusBar   *display.StatusBar

	currentDoc *content.Document
	stats      *content.Stats
	wifiPanel  *content.WifiPanel

	wifiActive = true
	btActive   = true

	metaKey uint16

	//keyboards []*input.Keyboard
)

func init() {
	flag.BoolVar(&debug, "d", false, "debug mode")
	flag.BoolVar(&noWelcome, "nw", false, "Skip welcome screen")
	flag.Float64Var(&vcom, "vcom", defaultVcom, "Display VCOM value (default: -2.13)")
	flag.IntVar(&orientationCCWDegrees, "rotation", defaultOrientationCCWDegrees, "Orientation in degrees, counterclockwise (0,90,180,270)")
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

func ensureBtStarted() {
	cmd := exec.Command("/usr/bin/bluetoothctl", "power", "on")
	if err := cmd.Run(); err != nil {
		Debug("Err: %v", err)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	_, err := io.WriteString(w, "Yeah!\n")
	if err != nil {
		Debug("Err: %v", err)
	}
}

func server() {
	http.HandleFunc("/", rootHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()

	go server()

	// external signal handler
	signal.Notify(signalChannel, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGUSR1)
	go signalHandler(signalChannel)

	if err := pi_sugar.Init(); err != nil {
		Debug("Pi Sugar not available")
	}
	defer pi_sugar.End()
	defer it8951.Close()
	defer terminate()

	piSugar, err := pi_sugar.NewPiSugar()
	if err != nil {
		Debug("No Pi Sugar detected: %v", err)
	}

	piSugar.Refresh()

	if vcom >= 0 || vcom < -5.0 {
		Debug("Wrong vcom value: %f", vcom)
		return // we're done
	}
	display.InitDisplay(uint16(-vcom*1000), it8951.Rotate(orientationCCWDegrees/90))

	displayUpdatesDone := make(chan bool, 1)
	go display.ScreenUpdater(displayUpdatesDone) // start background updater

	//os.Exit(0)
	for shouldRestart || freshStart {
		Debug("%s starting", nv())
		Debug("%s", copyright)

		// init

		Debug("Starting serious business")

		// start searching/updating keyboards
		keyboardChannel := input.KeyboardChannel()

		keyChannel := input.KeyChannel()
		eventChannel := make(chan bool, 10)
		go input.Search(eventChannel)

		display.InitScreen()

		//time.Sleep(time.Duration(3000) * time.Millisecond)
		//return
		display.Screen.View.Fill(0, display.White, display.Black).Update()

		ensureBtStarted()
		statusBar = display.Screen.NewStatusBar(piSugar, screenBgColor)

		mainWindow = display.Screen.NewWindow(0, 0, display.Screen.W, display.Screen.H, display.WindowOptions{
			Title:        "NV-Reloaded",
			TitleBar:     true,
			Border:       1,
			BgColor:      display.White,
			BorderColor:  display.Black,
			Transparency: 0,
			StatusBar:    statusBar,
		})

		if noWelcome {
			empty := &content.Empty{}
			mainWindow.SetContent(empty, 0, 0).Load()
		} else {
			welcome := &content.Welcome{}
			mainWindow.SetContent(welcome, 150, 150).Load()
		}
		//fnNewDocument()
		freshStart = false
		shouldTerminate = false
		shouldRestart = false

		go statusBar.Run()
		//heartBeatTicker := time.NewTicker(time.Duration(1000) * time.Millisecond)

		for !shouldTerminate && !shouldRestart {
			//display.CheckAlertBox() // close alertbox if marked that way

			select {
			//case <-heartBeatTicker.C:
			//statusBar.Refresh()
			case kbd := <-keyboardChannel:
				if kbd == nil {
					panic("Keyboard search failed")
				}
				if kbd.File == "none" {
					Debug("No keyboard found")
					statusBar.SetKbdState(false)
					mainWindow.AlertBox("No Keyboard...", 0)
				} else {
					statusBar.SetKbdState(true)
					Debug("Keyboard %s added", kbd.Name)
					mainWindow.AlertBox(fmt.Sprintf("%s Found", kbd.Name), 1000*time.Millisecond)
				}
			case event := <-keyChannel:
				display.CancelAlertBox()
				//statusBar.SetKbdState(true)
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
				} else if event.TypeName == "EV_KEY" {
					mainWindow.KeyEvent(event)
				}
			case <-terminateChannel:
				shouldTerminate = true
			case <-mainWindow.RefreshChannel:
				mainWindow.KeyEvent(nil)
			}
		}
		display.CancelAlertBox()
		statusBar.Close() // ensure the go routine ends
		Debug("closed status bar")
		<-statusBar.DoneChannel // ensure status bar is off
		Debug("We're done")
		eventChannel <- true

		win := display.Screen.NewWindow(0, 0, display.Screen.W, display.Screen.H, display.WindowOptions{
			//Title:       "NV PowerOff",
			TitleBar:     false,
			Border:       0,
			BgColor:      display.Gray13,
			BorderColor:  display.Black,
			Transparency: 0.1,
		})
		font := &fonts.IsoMetrixNF_Bold30pt8b
		if shouldPowerOff {
			win.SetTextArea(font, 0, 0).
				WriteCenteredIn(0, 0, win.W, win.H, "Powering Off...", display.Black, display.Gray13).
				Update()

			time.Sleep(time.Duration(2000) * time.Millisecond)
			//win.Fill(0, display.White, display.Black).
			//	Update()
			display.ShowGallery()
			//display.ShowLogo()
			if err := exec.Command("shutdown", "-P", "now").Run(); err != nil {
				Debug("Shutdown Error: %s", err)
			}
		} else if shouldReboot {
			win.SetTextArea(font, 0, 0).
				WriteCenteredIn(0, 0, win.W, win.H, "Rebooting...", display.Black, display.White).
				Update()

			time.Sleep(time.Duration(2000) * time.Millisecond)
			//win.Fill(0, display.White, display.Black).
			//	Update()
			display.ShowGallery()
			//display.ShowLogo()
			if err := exec.Command("shutdown", "-r", "now").Run(); err != nil {
				Debug("Error Rebooting: %s", err)
			}
		} else if shouldRestart {
			win.SetTextArea(font, 0, 0).
				WriteCenteredIn(0, 0, win.W, win.H, "Restarting...", display.Black, display.White).
				Update()

			//time.Sleep(time.Duration(2000) * time.Millisecond)
		} else {
			win.SetTextArea(font, 0, 0).
				WriteCenteredIn(0, 0, win.W, win.H, "Terminating", display.Black, display.White).
				Update()
			display.ShowGallery()

			//time.Sleep(time.Duration(500) * time.Millisecond)
			//win.Fill(0, display.White, display.Black).
			//	Update()
		}
	}
	display.UpdateScreen(nil)
	<-displayUpdatesDone
}

func terminate() {
	Debug("Terminating on request")
}

func signalHandler(c chan os.Signal) {
	for {
		select {
		case s := <-c:
			Debug("Got signal: %v", s)
			if s == syscall.SIGKILL || s == syscall.SIGINT || s == syscall.SIGTERM || s == syscall.SIGHUP {
				terminateChannel <- true
			} else if s == syscall.SIGUSR1 {
				shouldRestart = true
			}
		default:
		}
	}
}
