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

package content

import (
	"bufio"
	it8951 "github.com/peergum/IT8951-go"
	"nv/display"
	"os"
	"strings"
)

type Document struct {
	Title      string
	Filename   string
	Words      []string
	Paragraphs []string
	Cursor     int // position in text
	Modified   bool
	//TopWord        int   // word at top of displayed area
	//BottomWord     int   // word at bottom of displayed area
}

func (document *Document) Load() {
	if document.Filename == "" {
		document.Title = "New Document"
		document.Words = []string{}
		document.Paragraphs = []string{}
		document.Cursor = 0
		document.Modified = false
		Debug("New document")
		return
	}
	f, err := os.Open(os.Getenv("HOME") + "/" + document.Filename)
	if err != nil {
		Debug("Document open error: %s", err.Error())
		return
	}
	defer f.Close()
	lineScanner := bufio.NewScanner(f)
	lineScanner.Split(bufio.ScanLines)
	for lineScanner.Scan() {
		paragraph := lineScanner.Text()
		Debug("Paragraph: %s", paragraph)
		document.Paragraphs = append(document.Paragraphs, paragraph)
		wordScanner := bufio.NewScanner(strings.NewReader(paragraph))
		wordScanner.Split(bufio.ScanWords)
		for wordScanner.Scan() {
			word := wordScanner.Text()
			Debug("Word: %s", word)
			document.Words = append(document.Words, word)
		}

	}
	Debug("Document loaded: %s (%d words)", document.Filename, len(document.Words))
}

func (document *Document) Refresh() {
}

func (document *Document) Save() {
}

func (document *Document) GetTitle() string {
	return document.Title
}

func (document *Document) Display(view *display.View) {
	for i := 0; i < 4; i++ {
		view.WriteAt(100, 100+30*i, "This is a test", 0x0, it8951.Color(view.TextArea.BgColor))
	}
	view.Update()
}

func (document *Document) Print(view *display.View) {
	y := 0
	var xb, yb, wb, hb int
	var text string
	for _, paragraph := range document.Paragraphs {
		text += paragraph + "\n"
		view.GetTextBounds(text, 0, y, &xb, &yb, &wb, &hb)
		if y+hb > view.InnerH {
			break
		}
	}
	view.WriteAt(0, y, text, 0x0, it8951.Color(view.TextArea.BgColor))
}
