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
	"fmt"
	"nv/display"
	"nv/input"
	"os"
	"strings"
	"time"
)

type Document struct {
	Title            string
	Filename         string
	Words            *Element   // Word pointer to easily navigate/edit
	Paragraphs       *Paragraph // Paragraph pointer to easily navigate/edit
	LastParagraph    *Paragraph
	cCount           int
	wCount           int
	pCount           int
	lCount           int
	currentParagraph *Paragraph // current paragraph
	currentElement   *Element   // current char in element
	pos              int
	Modified         bool
	view             *display.View
	x, y             int
	Ready            bool // document ready to edit
	//Editable         bool
	//TopWord        int   // word at top of displayed area
	//BottomWord     int   // word at bottom of displayed area
}

type Paragraph struct {
	next         *Paragraph
	prev         *Paragraph
	firstElement *Element
	lastElement  *Element
}

// Elements are space separated tokens in a paragraph
type Elements []Element

type Element struct {
	word      []rune   // a simple word
	before    []rune   // an attached prefix or punctuation
	after     []rune   // an attached suffix or punctuation
	emphasis  Emphasis // special effects
	prev      *Element // points to next Element
	next      *Element // points to next Element
	paragraph *Paragraph
}

//type Words []Word
//type Word string

type Emphasis uint8

const (
	Bold Emphasis = 1 << iota
	Italic
	Underline
	Strikethrough
)

var (
	cursorOn   bool
	cursorTS   time.Time
	keyTS      time.Time
	typedChars string
	charCount  int
	lines      map[int]Elements
	cLine      int
)

func (document *Document) Type() string {
	return "document"
}

func (document *Document) parse(text string, pp **Paragraph, pw **Element) {
	runeScanner := bufio.NewScanner(strings.NewReader(text))
	runeScanner.Split(bufio.ScanRunes)
	quoteMode := int32(0)
	for runeScanner.Scan() {
		w := *pw
		p := *pp
		word := runeScanner.Text()
		//fmt.Printf("parag. = %p, word = %p -> %s\n", p, w, w.word)
		for _, char := range word {
			switch char {
			case '\r':
				// do nothing
			case '\n':
				if string(w.before)+string(w.word)+string(w.after) != "" {
					fmt.Println(w)
					p = &Paragraph{
						prev: p,
						next: p.next,
					}
					p.firstElement = &Element{
						paragraph: p,
					}
					if p.next == nil {
						document.LastParagraph = p
					}
					document.wCount++
					document.pCount++
					*pp = p
					*pw = p.firstElement
				}
			case ' ':
				if string(w.before)+string(w.word)+string(w.after) != "" {
					fmt.Println(w)
					w.next = &Element{
						prev:      w,
						paragraph: p,
					}
					p.lastElement = w.next
					document.wCount++
					*pw = w.next
				}
			//case ',', '.', ':', ')', ']', '}':
			//	w.after = append(w.after, rune(char))
			//	fmt.Println(w)
			//	w.next = &Element{
			//		prev:      w,
			//		paragraph: p,
			//	}
			//	p.lastElement = w.next
			//	document.wCount++
			//	*pw = w.next
			case '-', '"':
				if quoteMode == 0 {
					w.before = append(w.before, rune(char))
					quoteMode = char
				} else {
					w.after = append(w.after, rune(char))
					quoteMode = 0
				}
			case '[', '{', '(':
				w.before = append(w.before, rune(char))
			default:
				w.word = append(w.word, rune(char))
			}
		}
	}
}

func (document *Document) Load(view *display.View) {
	Debug("Loading document")
	document.view = view
	cursorTS = time.Now()
	//view.SetCursor(view.InnerX+view.TextArea.MarginX+paragraphIndent, view.InnerY+view.TextArea.MarginY+paragraphSpacing)
	view.SetCursor(paragraphIndent, paragraphSpacing)
	//view.SetCursor(view.InnerX+view.TextArea.MarginX, view.InnerY+view.TextArea.MarginY+paragraphSpacing)

	document.Words = &Element{}
	document.Paragraphs = &Paragraph{
		firstElement: document.Words,
		lastElement:  document.Words,
	}
	if document.Filename == "" {
		document.Title = "New Document"
		document.currentElement = document.Words
		document.currentParagraph = document.Paragraphs
		document.Modified = false
		Debug("New document")
		//view.SetCursor(view.InnerX+view.TextArea.MarginX+paragraphIndent, view.InnerY+view.TextArea.MarginY+paragraphSpacing)
		view.SetCursor(paragraphIndent, paragraphSpacing)
		document.Ready = true
		return
	}
	var filename string
	if document.Filename[0] == '.' {
		filename = "assets/" + document.Filename
	} else {
		filename = os.Getenv("HOME") + "/" + document.Filename
	}
	f, err := os.Open(filename)
	if err != nil {
		Debug("Document open error: %s", err.Error())
		return
	}
	defer f.Close()
	lineScanner := bufio.NewScanner(f)
	lineScanner.Split(bufio.ScanLines)
	p := document.Paragraphs
	w := document.Words
	start := true
	for lineScanner.Scan() {
		if !start {
			p.next = &Paragraph{
				prev: p,
			}
			p = p.next
			w = &Element{
				paragraph: p,
			}
		}
		p.firstElement = w
		p.lastElement = w
		paragraph := lineScanner.Text()
		Debug("Paragraph: %s", paragraph)
		document.parse(paragraph, &p, &w)
		if len(w.before)+len(w.word)+len(w.after) > 0 {
			fmt.Println(w)
			document.wCount++
			p.lastElement = w
		}
		document.LastParagraph = p
		document.pCount++
		start = false
	}
	Debug("Document loaded: %s (%d paragraphs, %d words)", document.Filename, document.pCount, document.wCount)
	document.Ready = true
}

func (document *Document) Refresh() {
}

func (document *Document) Save() {
}

func (document *Document) GetTitle() string {
	return document.Title
}

func (document *Document) Print() {
	view := document.view
	y := view.InnerY + view.TextArea.MarginY
	view.FillRectangle(0, 0, view.InnerW, view.InnerH, 0, display.White, display.Black)
	cLine = 0
	//outOfScreen:=false
printLoop:
	for p := document.Paragraphs; p != nil; p = p.next {
		space := ""
		_, y = view.GetCursor()
		view.SetCursor(paragraphIndent, y+paragraphSpacing)
		document.currentParagraph = p
		var x, y int
		document.currentElement = document.Words
		for w := p.firstElement; w != nil; w = w.next {
			color := display.Black
			text := space + string(w.before) + string(w.word) + string(w.after)
			x, y = view.GetCursor()
			//Debug("%d,%d", x, y)
			x0, y0 := x, y
			xb, yb, _, hb := view.GetTextBounds(text, &x0, &y0)
			//Debug("%d,%d - %d,%d,%d,%d", x, y, xb, yb, wb, hb)
			if xb < x {
				text = "\n" + text
			}
			if yb+hb > view.InnerH {
				break printLoop
			}
			document.currentElement = w
			view.Write(text, color, display.Transparent)
			space = " "
		}
		view.Write("\n", display.Black, display.Transparent)
		x, y = view.GetCursor()
		view.SetCursor(paragraphIndent, y+paragraphSpacing)

		// new
		//		//_, y = view.GetCursor()
		////view.SetCursor(paragraphIndent, y+paragraphSpacing)
		//document.currentParagraph = p
		//x, y := paragraphIndent, paragraphSpacing
		//document.currentElement = document.Words
		//lines[cLine] = Elements{}
		//for w := p.firstElement; w != nil; w = w.next {
		//	color := display.Black
		//	text = space + string(w.before) + string(w.word) + string(w.after)
		//	x, y = view.GetCursor()
		//	//Debug("%d,%d", x, y)
		//	xb, yb, _, hb := view.GetTextBounds(text, &x, &y)
		//	//Debug("%d,%d - %d,%d,%d,%d", x, y, xb, yb, wb, hb)
		//	if xb < x {
		//		//
		//		text = "\n" + text
		//	}
		//	if yb+hb > view.InnerH {
		//		outOfScreen = true
		//	}
		//	lines[cLine] = append(lines[cLine], w)
		//	document.currentElement = w
		//	view.SetCursor(x, y)
		//	view.Write(text, color, display.Transparent)
		//	space = " "
		//}
		//
		//x, y = view.GetCursor()
		//view.SetCursor(paragraphIndent, y+paragraphSpacing)
		////view.Write("\n", display.Black, display.Transparent)
	}
}
func (document *Document) BackSpace() {
	keyTS = time.Now()
	cursorTS = keyTS
	document.forceCursor(false)
	var char rune
	foundChar := false
	var cP *Paragraph
	var cE *Element
	for !foundChar {
		cP = document.currentParagraph
		cE = document.currentElement
		if cP == nil || cE == nil {
			return
		}
		if len(cE.after) > 0 {
			char = cE.after[len(cE.after)-1]
			cE.after = cE.after[:len(cE.after)-1]
			foundChar = true
		} else if len(cE.word) > 0 {
			char = cE.word[len(cE.word)-1]
			cE.word = cE.word[:len(cE.word)-1]
			foundChar = true
		} else if len(cE.before) > 0 {
			char = cE.before[len(cE.before)-1]
			cE.before = cE.before[:len(cE.before)-1]
			foundChar = true
		} else {
			//previous word
			if cE.prev != nil {
				Debug("jump to prev element")
				Debug("previous word: %s", string(cE.prev.word))
				// back 1 element and skip the one deleted
				if cE.next != nil {
					Debug("next word: %s", string(cE.next.word))
					cE.next.prev = cE.prev
				}
				cE.prev.next = cE.next
				if cP.lastElement == cE {
					// if it was last element in paragraph, update
					cP.lastElement = cE.prev
				}
				document.currentElement = cE.prev
				char = ' '
				foundChar = true
			} else if cP.prev != nil {
				Debug("jump to prev paragraph")
				if cP.next != nil {
					cP.next.prev = cP.prev
				} else {
					document.LastParagraph = nil
				}
				cP.prev.next = cP.next
				document.currentParagraph = cP.prev
				document.currentElement = cP.prev.lastElement
				Debug("previous paragraph elements: first=%s, last=%s",
					string(cP.prev.firstElement.word),
					string(cP.prev.lastElement.word))
				// move cursor and loop
			} else {
				// we're at the top... nothing to backspace
				return
			}
		}
	}
	view := document.view

	Debug("backspace on [%s], char = %s", cE, string(char))

	x0, y0 := view.GetCursor()
	x, y := x0, y0
	minX, minY, maxX, maxY := view.GetCharBounds(char, &x, &y)
	Debug("%d,%d,%d,%d - %d,%d,%d,%d", x0, y0, x, y, minX, minY, maxX, maxY)
	view.SetCursor(x0-(x-x0), y0)
	if char != ' ' {
		characterView := view.NewView(x0-(maxX-minX+1)-(minX-x0)-view.InnerX, minY, maxX-minX+3, maxY-minY+3, 4)
		//Debug("%d,%d,%d,%d", minX, minY, maxX, maxY)
		characterView.FillRectangle(0, 0, maxX-minX+3, maxY-minY+3, 0, display.White, display.Black)
		characterView.Update()
	}
}

func (document *Document) Editor(event *input.KeyEvent) {
	view := document.view

	// check if fake event (to force printing)
	if event != nil {
		// this is a real event
		Debug("%s (%c)", event.KeyName, event.Char)
		if event.Value == 0 {
			// do nothing on key release
			return
		}

		keyTS = time.Now()
		cursorTS = time.Now()
		document.forceCursor(false)
		if event.Char == 0 {
			switch event.KeyName {
			case "KEY_BACKSPACE":
				// still in typing buffer
				if charCount > 0 {
					typedChars = typedChars[:len(typedChars)-1]
					charCount--
					return
				}
				// already printed
				document.BackSpace()
				keyTS = time.Now()
				cursorTS = time.Now()
				return
			case "KEY_ENTER":
				event.Char = '\n'
			}
		}

		if cursorOn {
			document.forceCursor(false)
		}

		elapsed := time.Since(keyTS)
		typedChars += string(event.Char)
		//Debug("elapsed: %d, charCount: %d", elapsed, charCount)
		if elapsed.Milliseconds() < 70 && charCount < 5 {
			charCount++
			if event.Char != '\n' {
				return
			}
		}
	}

	document.parse(typedChars, &document.currentParagraph, &document.currentElement)

	lineSplit := strings.Split(typedChars, "\n")
	Debug("%s", lineSplit)
	for _, text := range lineSplit {
		//x, y := 0, ParagraphSpacing
		//view.SetCursor(0, ParagraphSpacing)
		x0, y0 := view.GetCursor()
		x, y := x0, y0
		//height := int(view.TextArea.Font.YAdvance)
		//minX, minY, maxX, maxY := 10000, 10000, -1, -1
		xb, yb, wb, hb := view.GetTextBounds(text, &x, &y)
		if text != "" && wb > 0 && hb > 0 {
			//Debug("%s: %d,%d,%d,%d,%d,%d,%d,%d", typedChars, x0, y0, x, y, xb, yb, wb, hb)
			characterView := view.NewView(xb, yb, wb, hb, 4)
			characterView.FillRectangle(0, 0, wb, hb, 0, display.White, display.Black)
			characterView.SetTextArea(view.TextArea.Font, 0, 0).
				SetCursor(x0-xb, y0-yb)
			characterView.Write(text, display.Black, display.Transparent)
			x2, _ := characterView.GetCursor()
			characterView.Update()
			//view.SetCursor(x-view.TextArea.MarginX, y-view.TextArea.MarginY)
			view.SetCursor(x0+(xb-x0)+x2, y0)
		}
		if len(lineSplit) > 1 {
			//view.WriteChar('\n', display.Black, display.Transparent)
			x, y = view.GetCursor()
			view.SetCursor(paragraphIndent, y0+paragraphSpacing)
		}
	}
	//paragraphNum := len(document.Paragraphs) - 1
	//document.Paragraphs[paragraphNum] += typedChars

	typedChars = ""
	charCount = 0
	cursorTS = time.Now()
	keyTS = time.Now()
}

func (document *Document) ToggleCursor() {
	//elapsed := time.Now().Sub(cursorTS)
	keyElapsed := time.Now().Sub(keyTS)

	// check if we need to print something...
	if keyElapsed.Milliseconds() > 100 && charCount > 0 {
		document.Editor(nil) // force printing
	}

	// handle possible different durations for cursor on and off
	//if ((cursorOn && elapsed.Milliseconds() > cursorOnDuration) ||
	//	(!cursorOn && elapsed.Milliseconds() > cursorOffDuration)) &&
	//	keyElapsed.Milliseconds() > cursorRestartDelay {
	if keyElapsed.Milliseconds() > cursorRestartDelay {
		document.forceCursor(true)
	}
}

func (document *Document) forceCursor(on bool) {
	view := document.view
	x, y := view.GetCursor()
	height := int(view.TextArea.Font.YAdvance) - 1
	if (on && !cursorOn) || (!on && cursorOn) {
		cursorView := view.NewView(x+1, y-height+8, cursorWidth, height, 4)
		color := display.White
		if on {
			color = display.Black
		}
		cursorView.Fill(0, color, color)
		cursorOn = on
		cursorView.Update()
	}
	cursorTS = time.Now()
}

func (element Element) String() string {
	return "[" + string(element.before) + string(element.word) + string(element.after) + "]"
}
