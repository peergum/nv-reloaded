/*
   main,
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
	"nv/content"
	"nv/display"
	"nv/input"
)

var (
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
		input.Alt:  {"Exit", fnExit},
	}
	F5 = content.FunctionKey{
		input.None:  {"Top", fnDoNothing},
		input.Shift: {"Bottom", fnDoNothing},
	}
	F6 = content.FunctionKey{
		input.None: {"Setup Wifi", fnWifiConfig},
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
		input.Alt:                            {"Restart", fnRestart},
		input.Shift | input.Ctrl | input.Alt: {"Factory Reset", fnDoNothing},
	}

	functions = content.FnPanel{
		FunctionKeys: content.FunctionKeys{0: F1, 1: F2, 2: F3, 3: F4, 4: F5, 5: F6, 6: F7, 7: F8, 8: F9, 9: F10, 10: F11, 11: F12},
	}
)

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
			functions.SetMeta(input.None)
			fnWindow.SetContent(&functions, 0, 0).
				Load()
		} else {
			fnWindow.Show()
		}
	} else {
		fnWindow.Hide()
	}
}

func fnToggleStats() {
	if mainWindow.GetContentType() != "document" {
		return
	}
	statsWindowToggle = !statsWindowToggle

	if statsWindowToggle {
		if statsWindow == nil {
			stats = &content.Stats{
				Document: currentDoc,
			}

			x := (mainWindow.InnerW - statsWidth) / 2
			y := (mainWindow.InnerH - statsHeight) / 2
			statsWindow = mainWindow.NewWindow(x, y, statsWidth, statsHeight, display.WindowOptions{
				Title:       "Stats",
				TitleBar:    true,
				Border:      2,
				BorderColor: display.Black,
				BgColor:     display.Gray14,
				TopRounded:  true,
			})

			statsWindow.SetContent(stats, 10, 10).
				Load()

		} else {
			statsWindow.Show()
		}
		statsWindowToggle = true
	} else {
		statsWindow.Hide()
		statsWindowToggle = false
	}
}

func fnNewDocument() {
	if fnWindowToggle {
		fnWindow.Hide()
		fnWindowToggle = false
	}
	currentDoc = &content.Document{
		Filename: "",
		Title:    "New Document",
	}

	mainWindow.SetContent(currentDoc, 10, 10).Load()
}

func fnLoadDocument() {
	if fnWindowToggle {
		fnWindow.Hide()
		fnWindowToggle = false
	}
	currentDoc = &content.Document{
		Filename: "sample2.txt",
		Filetype: "asset",
		//Filename: "document.docx",
		Title: "Example Document",
	}
	mainWindow.SetContent(currentDoc, 10, 10).Load()
}

func fnShutdown() {
	shouldTerminate = true // normal end
	shouldPowerOff = true
}

func fnReboot() {
	shouldTerminate = true // normal end
	shouldReboot = true
}

func fnExit() {
	shouldTerminate = true
}

func fnRestart() {
	shouldRestart = true
}

func fnWifiConfig() {
	wifiWindowToggle = !wifiWindowToggle

	if wifiWindowToggle {
		if wifiWindow == nil {
			wifiPanel = &content.WifiPanel{
				//Config: &content.WifiConfig{},
			}

			x := (mainWindow.InnerW - wifiWidth) / 2
			y := (mainWindow.InnerH - wifiHeight) / 2
			wifiWindow = mainWindow.NewWindow(x, y, wifiWidth, wifiHeight, display.WindowOptions{
				Title:        "Wifi Config",
				TitleBar:     true,
				Border:       2,
				BorderColor:  display.Black,
				BgColor:      display.White,
				Transparency: 0,
				TopRounded:   true,
				//Radius:       10,
			})

			wifiWindow.SetContent(wifiPanel, 10, 10).
				Load()

		} else {
			wifiWindow.Show()
		}
		wifiWindowToggle = true
	} else {
		wifiWindow.Hide()
		wifiWindowToggle = false
	}
}
