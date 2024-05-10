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
	if mainWindow.GetContentType() != "document" {
		return
	}
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

	mainWindow.SetContent(currentDoc, 10, 10).Load().Update()
}

func fnLoadDocument() {
	if fnWindowToggle {
		fnWindow.Hide()
		fnWindowToggle = false
	}
	currentDoc = &content.Document{
		Filename: "abc.txt",
		Title:    "Example Document",
	}
	mainWindow.SetContent(currentDoc, 10, 10).Load().Update()
}

func fnShutdown() {
	shouldTerminate = true // normal end
	shouldPowerOff = true
}

func fnReboot() {
	shouldTerminate = true // normal end
	shouldReboot = true
}
