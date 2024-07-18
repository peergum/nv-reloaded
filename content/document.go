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
	"embed"
	"io/fs"
	"math"
	rand2 "math/rand"
	display "nv/display"
	"nv/display/fonts-go"
	"nv/input"
	"os"
	"strings"
	"time"
)

//go:embed assets
var Files embed.FS

type Document struct {
	Title                string
	Filename             string
	Filetype             string
	Words                *Element   // Word pointer to easily navigate/edit
	Paragraphs           *Paragraph // Paragraph pointer to easily navigate/edit
	LastParagraph        *Paragraph
	Lines                map[int]Elements
	cCount               int
	wCount               int
	pCount               int
	lCount               int
	cLine                int
	topLine              int
	bottomLine           int
	currentParagraph     *Paragraph // current paragraph
	currentElement       *Element   // current char in element
	pos                  int        // cursor position in element
	Modified             bool
	RefreshNeeded        bool
	view                 *display.View
	scrollBarView        *display.View
	scrollBar            bool
	x, y                 int
	mx, my               int
	Ready                bool // document ready to edit
	paragraphIndent      bool
	paragraphIndentValue string
	paragraphSpacing     bool
	refreshChannel       chan bool
	closingChannel       chan bool
}

type Paragraph struct {
	next         *Paragraph
	prev         *Paragraph
	firstElement *Element
	lastElement  *Element
	yStart       int
	yEnd         int
}

// Elements are space separated tokens in a paragraph
type Elements []*Element

type Element struct {
	word      []rune   // a simple word
	before    []rune   // an attached prefix or punctuation
	after     []rune   // an attached suffix or punctuation
	emphasis  Emphasis // special effects
	prev      *Element // points to next Element
	next      *Element // points to next Element
	xStart    int      // x start position in line
	xWidth    int      // x end position in line
	yStart    int      // y start position in line
	yHeight   int      // y end position in line
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
	cursorOn             bool
	cursorTS             time.Time
	keyTS                time.Time
	typedChars           string
	charCount            int
	lines                map[int]Elements
	typingBufferMaxDelay = 200 * time.Millisecond
	movingKeyCount       int // how many times arrows were pressed
	direction            int
)

func (document *Document) Init(view *display.View, refreshChannel chan bool) (views []*display.View) {
	document.refreshChannel = refreshChannel
	document.closingChannel = make(chan bool)
	// set up default margins
	Debug("Initializing document")
	if document.mx == 0 || document.my == 0 {
		document.mx = 20
		document.my = 20
		document.scrollBar = true
		document.paragraphIndent = defaultParagraphIndent
		document.paragraphIndentValue = defaultParagraphIndentValue
		document.paragraphSpacing = defaultParagraphSpacing

	}
	scrollBarWidth := 0
	if document.scrollBar {
		scrollBarWidth = defaultScrollBarWidth
	}
	document.view = view.NewView(0, 0, view.InnerW-scrollBarWidth, view.InnerH, 4).
		Fill(0, display.White, display.Black).
		SetTextArea(regularFont, document.mx, document.my) /*.
		Update()*/
	views = append(views, document.view)
	if document.scrollBar {
		document.scrollBarView = view.NewView(view.InnerW-scrollBarWidth, 0, scrollBarWidth, view.InnerH, 4).
			Fill(0, defaultScrollBarBgColor, display.Black).
			DrawVLine(0, 0, view.InnerH, 1, display.Black).
			Update()
		document.scrollBarView.InnerX += 1
		document.scrollBarView.InnerW -= 1
		views = append(views, document.scrollBarView)
	}
	document.view.SetCursor(document.view.InnerX+document.view.TextArea.MarginX, document.view.InnerY+document.view.TextArea.MarginY)
	go document.ToggleCursor()
	return views
}

func (document *Document) Close() {
	Debug("Closing document")
	document.closingChannel <- true // stops go routines
}

func (document *Document) Type() string {
	return "document"
}

func (document *Document) Load() {
	Debug("Loading document")
	view := document.view
	cursorTS = time.Now()
	//view.SetCursor(view.InnerX+view.TextArea.MarginX+paragraphIndent, view.InnerY+view.TextArea.MarginY+paragraphSpacing)
	view.SetCursor(0, 0)
	//view.SetCursor(view.InnerX+view.TextArea.MarginX, view.InnerY+view.TextArea.MarginY+paragraphSpacing)

	spinner := view.NewSpinner("Loading...")
	doneChannel := make(chan bool)
	go spinner.Run(doneChannel)
	go document.loader(doneChannel)
	<-spinner.Done // wait for spinner to end
}

func (document *Document) loader(doneChannel chan<- bool) {
	document.Words = &Element{} // create empty first element
	// then a first paragraph pointing to that element
	document.Paragraphs = &Paragraph{
		firstElement: document.Words,
		lastElement:  document.Words,
	}
	document.Words.paragraph = document.Paragraphs // first element should also point at first paragraph
	document.cLine = 0
	document.topLine = 0
	document.bottomLine = 0
	direction = 0
	if document.Filename == "" {
		document.Title = "New Document"
		document.currentElement = document.Words
		document.currentParagraph = document.Paragraphs
		document.Modified = false
		Debug("New document")
		//view.SetCursor(view.InnerX+view.TextArea.MarginX+paragraphIndent, view.InnerY+view.TextArea.MarginY+paragraphSpacing)
		document.Ready = true
		doneChannel <- true
		return
	}
	var filename string

	var lineScanner *bufio.Scanner
	if document.Filetype == "asset" {
		// internal asset read
		filename = "assets/" + document.Filename
		f, err := Files.Open(filename)
		if err != nil {
			Debug("Can't open asset [%s]: %s", document.Filename, err.Error())
			return
		}
		defer func(f fs.File) {
			err := f.Close()
			if err != nil {
				Debug("Can't close asset [%s]: %s", document.Filename, err.Error())
			}
		}(f)
		lineScanner = bufio.NewScanner(f)
	} else {
		// external file read
		filename = "/var/nv/" + document.Filename
		f, err := os.Open(filename)
		if err != nil {
			Debug("Document open error: %s", err.Error())
			return
		}
		defer func(f *os.File) {
			err := f.Close()
			if err != nil {
				Debug("Document close error: %s", err.Error())
			}
		}(f)
		lineScanner = bufio.NewScanner(f)
	}
	lineScanner.Split(bufio.ScanLines)
	document.scanWords(lineScanner)
	document.scanLines()
	Debug("Document loaded: %s (%d paragraphs, %d words, %d lines)", document.Filename, document.pCount, document.wCount, document.lCount)
	document.Ready = true
	doneChannel <- true
}

func (document *Document) scanWords(lineScanner *bufio.Scanner) {
	p := document.Paragraphs
	w := document.Words
	document.cCount = 0
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
		document.cCount += document.parse(paragraph, &p, &w)
		if len(w.before)+len(w.word)+len(w.after) > 0 {
			//fmt.Println(w)
			document.wCount++
			p.lastElement = w
		}
		document.LastParagraph = p
		document.pCount++
		start = false
	}
	if err := lineScanner.Err(); err != nil {
		Debug("Error scanning doc: %s", err)
	}
}

func (document *Document) parse(text string, pp **Paragraph, pw **Element) int {
	runeScanner := bufio.NewScanner(strings.NewReader(text))
	runeScanner.Split(bufio.ScanRunes)
	//quoteMode := int32(0)
	cCount := 0
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
					//fmt.Println(w)
					w.paragraph = p
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
					//fmt.Println(w)
					w.paragraph = p
					w.next = &Element{
						prev:      w,
						paragraph: p,
					}
					p.lastElement = w.next
					document.wCount++
					*pw = w.next
				}
			//case ',', '.', ';', ':', ')', ']', '}':
			//	w.after = append(w.after, rune(char))
			//	//fmt.Println(w)
			//	w.paragraph = p
			//	w.next = &Element{
			//		prev:      w,
			//		paragraph: p,
			//	}
			//	p.lastElement = w.next
			//	document.wCount++
			//	*pw = w.next
			//case '"':
			//	if quoteMode == 0 {
			//		w.before = append(w.before, rune(char))
			//		quoteMode = char
			//	} else {
			//		w.after = append(w.after, rune(char))
			//		quoteMode = 0
			//	}
			//case '[', '{', '(':
			//	w.before = append(w.before, rune(char))
			default:
				char = filter(char)
				w.word = append(w.word, rune(char))
			}
		}
		cCount++
	}
	return cCount
}

func filter(char rune) rune {
	switch char {
	case '’': // apostrophe
		char = '\''
	}
	return char
}

func (document *Document) Refresh() {
}

func (document *Document) Save() {
}

func (document *Document) GetTitle() string {
	return document.Title
}

func (document *Document) Print() {
	Debug("Printing document")
	view := document.view
	//document.cLine = rand.Int() % max(1, document.lCount)
	if document.Modified || document.RefreshNeeded {
		view.Fill(0, display.White, display.Black)
	}
	//cLine = document.cLine
	var x, y int
	y = view.InnerY
	fontSize := 8
	if view.TextArea.Font != nil {
		fontSize = int(view.TextArea.Font.YAdvance)
	}
	maxLines := (view.InnerH - 2*view.TextArea.MarginY) / fontSize
	var startLine int
	if direction >= 0 { // top down
		startLine = document.topLine
		document.bottomLine = startLine + maxLines - 1
	} else {
		startLine = document.bottomLine - maxLines
		document.topLine = startLine
	}
	Debug("cline=%d, topline=%d, bottomline=%d", document.cLine, document.topLine, document.bottomLine)

	for l := startLine; l >= 0 && l < document.lCount && l < startLine+maxLines; l++ {
		ll := document.Lines[l]

		// quickly handle blank lines
		if len(ll) == 0 {
			y += int(view.TextArea.Font.YAdvance)
			continue
		}

		indent := ""
		if ll[0] == ll[0].paragraph.firstElement {
			indent = document.paragraphIndentValue // start paragraph with an indentation
		}
		maxWidth := view.InnerW - 2*view.TextArea.MarginX
		lineWidth := 0
		wordCount := len(ll)
		var spaceSize int
		// calculate space size if not last line
		text := indent
		for _, elem := range ll {
			text += string(elem.before) + string(elem.word) + string(elem.after)
			view.TextArea.SetFont(getFont(elem.emphasis))
			x0, y0 := view.InnerX, view.InnerY
			x, y := x0, y0
			xb, _, wb, _ := view.GetTextBounds(text, &x, &y)
			elem.xWidth = wb + (xb - x0)
			//Debug("%d,%d - %d,%d,%d,%d", x, y, xb, yb, wb, hb)
			lineWidth += wb + (xb - x0)
			spaceSize = int(math.Round(float64(maxWidth-lineWidth) / float64(wordCount)))
			text = ""
		}
		if ll[wordCount-1] == ll[0].paragraph.lastElement {
			spaceSize = 15
		}
		text = indent
		x = view.InnerX
		for _, elem := range ll {
			text += string(elem.before) + string(elem.word) + string(elem.after)
			view.TextArea.SetFont(getFont(elem.emphasis))
			view.SetCursor(x, y)
			view.Write(text, display.Black, display.White)
			x += elem.xWidth + spaceSize
			text = ""
		}
		y += int(view.TextArea.Font.YAdvance)
	}
	view.Update()

	// update scrollbar
	if document.scrollBar {
		Debug("printing scrollbar")
		scrollBarView := document.scrollBarView
		if document.lCount == 0 {
			// full bar
			scrollBarView.FillRectangle(0, 0, scrollBarView.InnerW, scrollBarView.InnerH, 0, defaultScrollBarColor, display.Black)
		} else {
			// 2 or 3 segments
			if document.topLine > 0 {
				scrollBarView.FillRectangle(0, 0, scrollBarView.InnerW, document.topLine*scrollBarView.InnerH/document.lCount, 0, defaultScrollBarBgColor, display.White)
			}
			h := maxLines
			if maxLines+document.topLine > document.lCount {
				h = document.lCount - document.topLine
			}
			scrollBarView.FillRectangle(0, document.topLine*scrollBarView.InnerH/document.lCount, scrollBarView.InnerW, scrollBarView.InnerH*h/document.lCount, 0, defaultScrollBarColor, display.White)
			if document.topLine+maxLines < document.lCount {
				scrollBarView.FillRectangle(0, scrollBarView.InnerH*(document.topLine+maxLines)/document.lCount, scrollBarView.InnerW, scrollBarView.InnerH*(document.lCount-document.topLine-maxLines)/document.lCount, 0, defaultScrollBarBgColor, display.White)
			}
		}
		scrollBarView.Update()
	}
}

func getFont(emphasis Emphasis) *fonts.GfxFont {
	switch Emphasis(rand2.Int()) % 8 {
	case Bold:
		return boldFont
	case Italic:
		return italicFont
	case Bold | Italic:
		return boldItalicFont
	default:
		return regularFont
	}
}

func (document *Document) scanLines() {
	view := document.view

	cLine := 0
	document.Lines = make(map[int]Elements, 20000)
	document.Lines[cLine] = Elements{}

	for p := document.Paragraphs; p != nil; p = p.next {

		indent := ""
		if document.paragraphIndent {
			indent = document.paragraphIndentValue // start paragraph with an indentation
		}
		maxWidth := view.InnerW - 2*view.TextArea.MarginX
		for nElem := p.firstElement; nElem != nil; {
			lineWidth := 0
			wordCount := 0
			elem := nElem
			for ; elem != nil; elem = elem.next {
				text := indent + string(elem.before) + string(elem.word) + string(elem.after)
				view.TextArea.SetFont(getFont(elem.emphasis))
				x0, y0 := view.InnerX, view.InnerY
				x, y := x0, y0
				xb, _, wb, _ := view.GetTextBounds(text, &x, &y)
				//Debug("%d,%d - %d,%d,%d,%d", x, y, xb, yb, wb, hb)
				lineWidth += wb + (xb - x0)
				spaceSize := 15
				if wordCount > 0 {
					spaceSize = int(math.Round(float64(maxWidth-lineWidth) / float64(wordCount)))
				}
				if lineWidth >= maxWidth || spaceSize < 15 {
					break
				}
			}
			for ww := nElem; ww != nil && ww != elem; ww = ww.next {
				document.Lines[cLine] = append(document.Lines[cLine], ww)
			}
			cLine++
			nElem = elem
		}

		//for w := p.firstElement; w != nil; w = w.next {
		//	text := string(w.before) + string(w.word) + string(w.after)
		//	x, y = view.GetCursor()
		//	x0, y0 := x, y
		//	//Debug("%d,%d", x, y)
		//	_, _, wb, _ := view.GetTextBounds(line+space+text, &x0, &y0)
		//	//Debug("%d,%d - %d,%d,%d,%d", x, y, xb, yb, wb, hb)
		//	space = " "
		//	if y != y0 || wb >= view.InnerW-2*view.TextArea.MarginX { // line wrapped should print
		//		//fmt.Printf("%d: [%s]\n", cLine, line)
		//		line = "" // continuation lines don't start with a space indentation
		//		space = ""
		//		cLine++
		//		//x = 0
		//	}
		//	line += space + text
		//	document.Lines[cLine] = append(document.Lines[cLine], w)
		//	document.currentElement = w
		//	view.SetCursor(x, y)
		//}
		//fmt.Printf("%d: [%s]\n", cLine, line)
		if document.paragraphSpacing {
			document.Lines[cLine] = Elements{}
			cLine++
		}
	}
	document.lCount = cLine
	document.cLine = 0
}

func (document *Document) BackSpace(metaKeys uint16) {
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

		document.forceCursor(false)

		fontSize := 8
		if view.TextArea.Font != nil {
			fontSize = int(view.TextArea.Font.YAdvance)
		}
		maxLines := (view.InnerH - 2*view.TextArea.MarginY) / fontSize
		if event.Char == 0 {
			switch event.KeyName {
			case "KEY_BACKSPACE":
				// still in typing buffer
				if charCount > 0 && event.MetaKeys == 0 {
					typedChars = typedChars[:len(typedChars)-1]
					charCount--
					return
				}
				// already printed
				document.BackSpace(event.MetaKeys)
				keyTS = time.Now()
				cursorTS = time.Now()
				return
			case "KEY_UP":
				if document.cLine == 0 {
					return
				}
				m := event.MetaKeys
				if m == 0 {
					document.cLine--
				} else if m&input.Shift != 0 {
					document.cLine -= (document.cLine - document.topLine) + maxLines
				} else if m&input.Ctrl != 0 {
					document.cLine = 0
				}
				if document.cLine < 0 {
					document.cLine = 0
				}
				movingKeyCount++
				document.RefreshNeeded = true
				document.refreshChannel <- true
				document.checkRefresh()
				keyTS = time.Now()
				cursorTS = time.Now()
				return
			case "KEY_DOWN":
				if document.cLine > document.lCount-1 {
					return
				}
				m := event.MetaKeys
				if m == 0 {
					document.cLine++
				} else if m&input.Shift != 0 {
					document.cLine += (document.bottomLine - document.cLine) + maxLines
				} else if m&input.Ctrl != 0 {
					document.cLine = document.lCount - 1
				}
				if document.cLine >= document.lCount {
					document.cLine = document.lCount - 1
				}
				movingKeyCount++
				document.RefreshNeeded = true
				document.checkRefresh()
				keyTS = time.Now()
				cursorTS = time.Now()
				return
			case "KEY_ENTER":
				event.Char = '\n'
			}
		}

		//if cursorOn {
		//	document.forceCursor(false)
		//}

		elapsed := time.Since(keyTS)
		typedChars += string(event.Char)
		keyTS = time.Now()
		cursorTS = time.Now()

		//Debug("elapsed: %d, charCount: %d", elapsed, charCount)
		if elapsed < typingBufferMaxDelay && charCount < 5 {
			charCount++
			if event.Char != '\n' {
				return
			}
		}
	}

	document.checkRefresh()
	if len(typedChars) == 0 {
		//cursorTS = time.Now()
		//keyTS = time.Now()
		return
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
			view.SetCursor(0, y0)
		}
	}
	//paragraphNum := len(document.Paragraphs) - 1
	//document.Paragraphs[paragraphNum] += typedChars

	typedChars = ""
	charCount = 0
	//cursorTS = time.Now()
	//keyTS = time.Now()
}

func (document *Document) checkRefresh() {
	elapsed := time.Since(keyTS)
	if elapsed < typingBufferMaxDelay || movingKeyCount < 3 {
		movingKeyCount++
		return
	}
	Debug("check Refreshing, cline=%d,top=%d,bottom=%d", document.cLine, document.topLine, document.bottomLine)
	// refresh at least one line (top or bottom if necessary)
	if document.cLine < document.topLine {
		document.topLine = document.cLine
		direction = 1 // top-down
		document.Print()
	} else if document.cLine > document.bottomLine {
		document.bottomLine = document.cLine
		direction = -1 // down-top
		document.Print()
	}
	movingKeyCount = 0
	document.RefreshNeeded = false
}

func (document *Document) ToggleCursor() {
	cursorTicker := time.NewTicker(cursorBlinkInterval)
cursorLoop:
	for {
		select {
		case <-cursorTicker.C:
			if document.Ready {
				keyElapsed := time.Now().Sub(keyTS)

				// check if we need to print something...
				if keyElapsed.Milliseconds() > 100 && charCount > 0 {
					document.refreshChannel <- true
				}

				// handle possible different durations for cursor on and off
				if keyElapsed.Milliseconds() > cursorRestartDelay {
					//if keyElapsed.Milliseconds() > cursorRestartDelay {
					cursorOn = !cursorOn
					document.forceCursor(cursorOn)
				}
			}
		case <-document.closingChannel:
			break cursorLoop
		}
	}
	cursorTicker.Stop()
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

func (document *Document) KeyEvent(event *input.KeyEvent) {
	Debug("Event %v", event)
	document.Editor(event)
}
