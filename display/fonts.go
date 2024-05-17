/*
   main,
   Copyright (C) 2024  Phil Hilger

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package display

import (
	it8951 "github.com/peergum/IT8951-go"
	"nv/display/fonts-go"
)

// TextArea follows the Content interface
type TextArea struct {
	MarginX  int            // margin X
	MarginY  int            // margin Y
	CurrentW int            // Display width as modified by current rotation
	CurrentH int            // Display height as modified by current rotation
	CX       int            // x location to start writing text
	CY       int            // y location to start writing text
	fgColor  uint16         // 16-bit background color for text
	BgColor  uint16         // 16-bit text color for text
	sX       int            // Desired magnification in X-axis of text to print()
	sY       int            // Desired magnification in Y-axis of text to print()
	Rotation uint16         // Display rotation (0 thru 3)
	Wrap     bool           // If set, 'wrap' text at right edge of display
	cp437    bool           // If set, use correct CP437 charset (default is off)
	Font     *fonts.GfxFont // Pointer to special Font
}

//
//func (area *TextArea) setBuffer(bpp int) {
//	ppw := 16 / bpp
//	// calculate drawing area (word rounded)
//	xmin := (area.X / ppw) * ppw              // low word limit
//	xmax := ((area.X+area.W-1)/ppw+1)*ppw - 1 // high word limit
//	ww := (xmax - xmin + 1) / ppw
//	wh := area.H
//	area.buffer = Buffer{
//		X:    xmin,
//		Y:    area.Y,
//		ww:   ww,
//		wh:   wh,
//		bpp:  bpp,
//		data: make(it8951.DataBuffer, ww*wh),
//	}
//	Debug("Set writing buffer (%d x %d) (xmin=%d/xmax=%d,y=%d) bpp=%d", ww*2, wh, xmin, xmax, area.Y, bpp)
//}

func (view *View) SetTextArea(font *fonts.GfxFont, mX, mY int) *View {
	area := TextArea{
		MarginX:  mX,
		MarginY:  mY,
		CurrentW: view.InnerW - 2*mX,
		CurrentH: view.InnerH - 2*mY,
		CX:       0,
		CY:       0,
		sX:       1,
		sY:       1,
		Wrap:     true,
		fgColor:  0x0,
		BgColor:  view.BgColor,
	}
	area.SetFont(font)
	//area.setBuffer(4)
	//Debug("area buffer = %v", area.buffer)
	view.TextArea = area
	return view
}

func (view *View) SetCursor(x, y int) *View {
	view.TextArea.CX = x // - view.TextArea.MarginX
	view.TextArea.CY = y // - view.TextArea.MarginY
	return view
}

func (view *View) GetCursor() (x, y int) {
	x = view.TextArea.CX // + view.TextArea.MarginX
	y = view.TextArea.CY // + view.TextArea.MarginY
	return x, y
}

/**************************************************************************/
/*!
  @brief   Draw a single character
   @param    x   Bottom left corner x coordinate
   @param    y   Bottom left corner y coordinate
   @param    c   The 8-bit font-indexed character (likely ascii)
   @param    color 16-bit 5-6-5 Color to draw chraracter with
   @param    bg 16-bit 5-6-5 Color to fill background with (if same as color,
  no background)
   @param    size_x  Font magnification level in X-axis, 1 is 'original' size
   @param    size_y  Font magnification level in Y-axis, 1 is 'original' size
*/
/**************************************************************************/
func (view *View) drawChar(x, y int, c rune,
	fg uint16, bg uint16, sizeX,
	sizeY int) {
	//Debug("drawChar (%d,%d) -> %c in buffer (%d,%d,%d,%d)",
	//	x, y, c, view.buffer.X, view.buffer.Y, view.buffer.ww, view.buffer.wh)
	if view.TextArea.Font == nil { // 'Classic' built-in font
		if (x >= view.TextArea.CurrentW) || // Clip right
			(y >= view.TextArea.CurrentH) || // Clip bottom
			((x + 6*view.TextArea.sX - 1) < 0) || // Clip left
			((y + 8*view.TextArea.sY - 1) < 0) { // Clip top
			return
		}

		if !view.TextArea.cp437 && (c >= 176) {
			c++ // Handle 'classic' charset behavior
		}
		for i := 0; i < 5; i++ { // Char bitmap = 5 columns
			line := defaultFont[int(c)*5+i]
			for j := 0; j < 8; j++ {
				xx := x + i*sizeX
				yy := y + j*sizeY
				if xx >= 0 && xx < view.TextArea.CurrentW &&
					yy >= 0 && yy < view.TextArea.CurrentH {
					if line&1 != 0 {
						if sizeX == 1 && sizeY == 1 {
							view.buffer.writePixel(view.InnerX+view.TextArea.MarginX+x, view.InnerY+view.TextArea.MarginY+yy, fg)
						} else {
							view.buffer.writeFillRectangle(view.InnerX+view.TextArea.MarginX+xx, view.InnerY+view.TextArea.MarginY+yy, sizeX, sizeY,
								fg)
						}
					} else if bg != 0xffff { // transparent color
						if sizeX == 1 && sizeY == 1 {
							view.buffer.writePixel(view.InnerX+view.TextArea.MarginX+xx, view.InnerY+view.TextArea.MarginY+yy, bg)
						} else {
							view.buffer.writeFillRectangle(view.InnerX+view.TextArea.MarginX+xx, view.InnerY+view.TextArea.MarginY+yy, sizeX, sizeY, bg)
						}
					}
				}
				line >>= 1
			}
		}
		if bg != fg { // If opaque, draw vertical line for last column
			if sizeX == 1 && sizeY == 1 {
				view.buffer.writeFastVLine(view.InnerX+view.TextArea.MarginX+x+5, view.InnerY+view.TextArea.MarginY+y, 8, bg)
			} else {
				view.buffer.writeFillRectangle(view.InnerX+view.TextArea.MarginX+x+5*sizeX, view.InnerY+view.TextArea.MarginY+y, sizeX, 8*sizeY, bg)
			}
		}
		//view.buffer.write(x, y, 6*sizeX, 8*sizeY)
	} else { // Custom font

		// Character is assumed previously filtered by write() to eliminate
		// newlines, returns, non-printable characters, etc.  Calling
		// drawChar() directly with 'bad' characters of font may cause mayhem!

		var w, h uint16
		bitmap := view.TextArea.Font.Bitmap
		c -= rune(view.TextArea.Font.First)
		glyph := view.TextArea.Font.Glyphs[c]
		bo := glyph.BitmapOffset
		w = glyph.Width
		h = glyph.Height
		xo := glyph.XOffset
		yo := glyph.YOffset

		var bits uint8
		var bit int

		//if sizeX > 1 || sizeY > 1 {
		//	xo16 = xo
		//	yo16 = yo
		//}

		// TODO: fix the buffer size

		//// calculate drawing area (word rounded)
		//xmin := ((x + int(xo)*sizeX) / ppw) * ppw              // low word limit
		//xmax := ((x+((int(xo)+int(w))*sizeX)-1)/ppw+1)*ppw - 1 // high word limit
		//ww := (xmax - xmin + 1) / ppw
		//wh := int(h) * sizeY
		//buffer := Buffer{
		//	xmin,
		//	y + int(yo)*sizeY,
		//	ww,
		//	wh,
		//	make(it8951.DataBuffer, ww*wh),
		//}
		//Debug("Set writing buffer (%d x %d), lcdFont (%d bytes), (%d/%d,%d)", ww*2, wh, len(defaultFont), xmin, xmax, buffer.y)

		// Todo: Add character clipping here

		// NOTE: THERE IS NO 'BACKGROUND' COLOR OPTION ON CUSTOM FONTS.
		// THIS IS ON PURPOSE AND BY DESIGN.  The background color feature
		// has typically been used with the 'classic' font to overwrite old
		// screen contents with new data.  This ONLY works because the
		// characters are a uniform size; it's not a sensible thing to do with
		// proportionally-spaced fonts with glyphs of varying sizes (and that
		// may overlap).  To replace previously-drawn text when using a custom
		// font, use the GetTextBounds() function to determine the smallest
		// rectangle encompassing a string, erase the area with fillRect(),
		// then draw new text.  This WILL unfortunately 'blink' the text, but
		// is unavoidable.  Drawing 'background' pixels will NOT fix this,
		// only creates a new set of problems.  Have an idea to work around
		// this (a canvas object type for MCUs that can afford the RAM and
		// displays supporting setAddrWindow() and pushColors()), but haven't
		// implemented this yet.

		for yy := 0; yy < int(h); yy++ {
			for xx := 0; xx < int(w); xx++ {
				// reload bits every byte
				if bit&7 == 0 {
					bits = bitmap[bo]
					bo++
				}
				bit++
				px := x + (int(xo)+xx)*sizeX
				py := y + (int(yo)+yy)*sizeY
				if px >= 0 && px < view.TextArea.CurrentW &&
					py >= 0 && py < view.TextArea.CurrentH {

					// bits left to right correspond to increasing x
					if bits&0x80 != 0 {
						if sizeX == 1 && sizeY == 1 {
							view.buffer.writePixel(view.InnerX+view.TextArea.MarginX+px, view.InnerY+view.TextArea.MarginY+py, fg)
						} else {
							view.buffer.writeFillRectangle(view.InnerX+view.TextArea.MarginX+px, view.InnerY+view.TextArea.MarginY+py,
								sizeX, sizeY, fg)
						}
					} else {
						if sizeX == 1 && sizeY == 1 {
							view.buffer.writePixel(view.InnerX+view.TextArea.MarginX+px, view.InnerY+view.TextArea.MarginY+py, bg)
						} else {
							view.buffer.writeFillRectangle(view.InnerX+view.TextArea.MarginX+px, view.InnerY+view.TextArea.MarginY+py,
								sizeX, sizeY, bg)
						}
					}
				}
				bits <<= 1
			}
		}
		//view.buffer.write(x+int(xo)*sizeX, y+int(yo)*sizeY, int(w)*sizeX, int(h)*sizeY)
	} // End classic vs custom font
}

func (view *View) WriteChar(c rune, color it8951.Color, bgColor it8951.Color) int {
	//Debug("area %v", area)
	//Debug("%d (%c)", c, c)
	if view.TextArea.Font == nil { // 'Classic' built-in font
		if c == '\n' { // Newline?
			view.TextArea.CX = 0                     // Reset x to zero,
			view.TextArea.CY += view.TextArea.sY * 8 // advance y one line
		} else if c != '\r' { // Ignore carriage returns
			if view.TextArea.Wrap && ((view.TextArea.CX + view.TextArea.sX*6) > view.TextArea.CurrentW) { // Off right?
				view.TextArea.CX = 0                     // Reset x to zero,
				view.TextArea.CY += view.TextArea.sY * 8 // advance y one line
			}
			view.drawChar(view.TextArea.CX, view.TextArea.CY, c, uint16(color), uint16(bgColor), view.TextArea.sX,
				view.TextArea.sY)
			view.TextArea.CX += view.TextArea.sX * 6 // Advance x one char
		}
	} else { // Custom font
		if c == '\n' {
			view.TextArea.CX = 0
			view.TextArea.CY +=
				view.TextArea.sY * int(view.TextArea.Font.YAdvance)
		} else if c != '\r' {
			first := view.TextArea.Font.First
			if (c >= rune(first)) && (c <= rune(view.TextArea.Font.Last)) {
				glyph := view.TextArea.Font.Glyphs[c-rune(first)]
				w := glyph.Width
				h := glyph.Height
				if (w > 0) && (h > 0) { // Is there an associated bitmap?
					xo := glyph.XOffset // sic
					if view.TextArea.Wrap && ((view.TextArea.CX + view.TextArea.sX*(int(xo)+int(w))) > view.TextArea.CurrentW) {
						view.TextArea.CX = 0
						view.TextArea.CY += view.TextArea.sY *
							int(view.TextArea.Font.YAdvance)
					}
					view.drawChar(view.TextArea.CX, view.TextArea.CY, c, uint16(color), uint16(bgColor), view.TextArea.sX,
						view.TextArea.sY)
				}
				view.TextArea.CX +=
					int(glyph.XAdvance) * view.TextArea.sX
			}
		}
	}
	return 1
}

func (view *View) Write(text string, color it8951.Color, bgColor it8951.Color) {
	for _, c := range text {
		view.WriteChar(c, color, bgColor)
	}
}

func (view *View) WriteAt(x, y int, text string, color it8951.Color, bgColor it8951.Color) *View {
	defer func() {
		if err := recover(); err != nil {
			Debug("view: %v, (%d,%d), %s", view, x, y, text)
		}
	}()
	//area := view.textArea
	x0, y0 := x+0, y+0
	view.Xb, view.Yb, view.Wb, view.Hb = view.GetTextBounds(text, &x0, &y0)
	view.TextArea.CX = x + (x - view.Xb)
	view.TextArea.CY = y + (y - view.Yb) // + view.Hb // - view.Yb
	view.Write(text, color, bgColor)
	return view
}

func (view *View) WriteCenteredIn(x, y, w, h int, text string, color it8951.Color, bgColor it8951.Color) *View {
	x0, y0 := x, y
	xb, yb, wb, hb := view.GetTextBounds(text, &x0, &y0)
	view.Xb = xb
	view.Yb = yb
	view.Wb = wb
	view.Hb = hb
	Debug("text bounds [%s],(%d,%d) -> (%d,%d,%d,%d)", text, x, y, xb, yb, wb, hb)
	return view.WriteAt(x+(w-wb)/2, y+(h-hb)/2, text, color, bgColor)
}
func (view *View) WriteVCenteredAt(x int, text string, color it8951.Color, bgColor it8951.Color) *View {
	x0, y0 := x, 0
	xb, yb, wb, hb := view.GetTextBounds(text, &x0, &y0)
	view.Xb = xb
	view.Yb = yb
	view.Wb = wb
	view.Hb = hb
	Debug("text bounds [%s],x=%d -> (%d,%d,%d,%d)", text, x, xb, yb, wb, hb)
	return view.WriteAt(x, (view.H-hb)/2, text, color, bgColor)
}

func (area *TextArea) SetFont(font *fonts.GfxFont) *TextArea {
	if font != nil { // Font struct pointer passed in?
		if area.Font == nil { // And no current font struct?
			// Switching from classic to new font behavior.
			// Move cursor pos down 6 pixels so it's on baseline.
			area.CY += 6
		}
	} else if area.Font != nil { // NULL passed.  Current font struct defined?
		// Switching from new to classic font behavior.
		// Move cursor pos up 6 pixels so it's at top-left of char.
		area.CY -= 6
	}
	area.Font = font
	return area
}

func (area *TextArea) SetSize(size int) *TextArea {
	area.sX = size
	area.sY = size
	return area
}

func (area *TextArea) SetSizes(sizeX, sizeY int) *TextArea {
	area.sX = sizeX
	area.sY = sizeY
	return area
}

func (view *View) GetCharBounds(c rune, x, y *int) (minX, minY, maxX, maxY int) {
	minX = 10000
	minY = 10000
	maxX = -1
	maxY = -1
	view.getCharBounds(c, x, y, &minX, &minY, &maxX, &maxY)
	return minX, minY, maxX, maxY
}

func (view *View) getCharBounds(c rune, x, y *int, minX, minY, maxX, maxY *int) {
	area := view.TextArea
	if area.Font != nil {
		if c == '\n' { // Newline?
			*x = 0 // Reset x to zero, advance y by one line
			*y += area.sY * int(area.Font.YAdvance)
		} else if c != '\r' { // Not a carriage return; is normal char
			first := int(area.Font.First)
			last := int(area.Font.Last)
			if (c >= rune(first)) && (c <= rune(last)) { // Char present in this font?
				glyph := area.Font.Glyphs[c-rune(first)]
				gw := glyph.Width
				gh := glyph.Height
				xa := glyph.XAdvance
				xo := glyph.XOffset
				yo := glyph.YOffset
				if area.Wrap && (*x+((int(xo)+int(gw))*area.sX)) > area.CurrentW {
					*x = 0 // Reset x to zero, advance y by one line
					*y += area.sY * int(area.Font.YAdvance)
				}
				tsx := area.sX
				tsy := area.sY
				x1 := *x + int(xo)*tsx
				y1 := *y + int(yo)*tsy
				x2 := x1 + int(gw)*tsx - 1
				y2 := y1 + int(gh)*tsy - 1
				if x1 < *minX {
					*minX = x1
				}
				if y1 < *minY {
					*minY = y1
				}
				if x2 > *maxX {
					*maxX = x2
				}
				if y2 > *maxY {
					*maxY = y2
				}
				*x += int(xa) * tsx
			}
		}
	} else { // Default font
		if c == '\n' { // Newline?
			*x = 0            // Reset x to zero,
			*y += area.sY * 8 // advance y one line
			// min/max x/y unchaged -- that waits for next 'normal' character
		} else if c != '\r' { // Normal char; ignore carriage returns
			if area.Wrap && ((*x + area.sX*6) > area.CurrentW) { // Off right?
				*x = 0            // Reset x to zero,
				*y += area.sY * 8 // advance y one line
			}
			x2 := *x + area.sX*6 - 1 // Lower-right pixel of char
			y2 := *y + area.sY*8 - 1
			if x2 > *maxX {
				*maxX = x2 // Track max x, y
			}
			if y2 > *maxY {
				*maxY = y2
			}
			if *x < *minX {
				*minX = *x // Track min x, y
			}
			if *y < *minY {
				*minY = *y
			}
			*x += area.sX * 6 // Advance x one char
		}
	}
}

func (view *View) GetTextBounds(text string, x, y *int) (xb, yb, wb, hb int) {
	minX := 10000
	minY := 10000
	maxX := -1
	maxY := -1
	// Bound rect is intentionally initialized inverted, so 1st char sets it

	xb = *x // Initial position is value passed in
	yb = *y
	wb = 0
	hb = 0 // Initial size is zero

	for _, c := range text {
		// charBounds() modifies x/y to advance for each character,
		// and min/max x/y are updated to incrementally build bounding rect.
		view.getCharBounds(c, x, y, &minX, &minY, &maxX, &maxY)
	}

	if maxX >= minX { // If legit string bounds were found...
		xb = minX            // Update x1 to least X coord,
		wb = maxX - minX + 1 // And w to bound rect width
	}
	if maxY >= minY { // Same for height
		yb = minY
		hb = maxY - minY + 1
	}
	return xb, yb, wb, hb
}
