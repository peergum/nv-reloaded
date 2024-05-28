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

package display

import (
	"encoding/binary"
	"errors"
	it8951 "github.com/peergum/IT8951-go"
	"math"
)

type Buffer struct {
	X    int // actual buffer x location
	Y    int // actual buffer y location
	ww   int // ww width in words
	wh   int // hh height in words
	inX  int // visible x
	inY  int // visible y
	inW  int // visible w
	inH  int // visible h
	bpp  int // buffer bpp
	data it8951.DataBuffer
}

type View struct {
	X, Y, W, H                     int // absolute coordinates
	InnerX, InnerY, InnerW, InnerH int // absolute internal coordinates
	BgColor                        uint16
	buffer                         Buffer
	TextArea                       TextArea
	content                        Content
	Xb, Yb, Wb, Hb                 int
	Views                          []*View
	//parent                         *View
}

const (
	Black it8951.Color = iota
	Gray1
	Gray2
	Gray3
	Gray4
	Gray5
	Gray6
	Gray7
	Gray8
	Gray9
	Gray10
	Gray11
	Gray12
	Gray13
	Gray14
	White
	Transparent it8951.Color = 0xffff
)

func (view *View) setBuffer(bpp int) {
	ppw := 2
	if bpp > 1 {
		ppw = 16 / bpp
	}
	// calculate drawing area (word rounded)
	xmin := (view.X / ppw) * ppw              // low word limit
	xmax := ((view.X+view.W-1)/ppw+1)*ppw - 1 // high word limit
	ww := (xmax - xmin + 1) / ppw
	wh := view.H
	view.buffer = Buffer{
		X:    xmin,
		Y:    view.Y,
		ww:   ww,
		wh:   wh,
		inX:  view.X,
		inY:  view.Y,
		inW:  view.W,
		inH:  view.H,
		bpp:  bpp,
		data: make(it8951.DataBuffer, ww*wh),
	}
	Debug("Set writing buffer words=(%d x %d) (xmin=%d/xmax=%d,y=%d) bpp=%d", ww, wh, xmin, xmax, view.Y, bpp)
}

func (view *View) NewView(x, y, w, h int, bpp int) *View {
	Debug("NewView (%d,%d,%d,%d bpp=%d)", x, y, w, h, bpp)
	newView := View{
		X: view.InnerX + x,
		Y: view.InnerY + y,
		W: min(view.InnerW, w),
		H: min(view.InnerH, h),
		// inner dimensions are the same by default
		InnerX:  view.InnerX + x,
		InnerY:  view.InnerY + y,
		InnerW:  min(view.InnerW, w),
		InnerH:  min(view.InnerH, h),
		BgColor: view.BgColor,
	}
	newView.setBuffer(bpp)
	newView.CopyPixels(newView.X, newView.Y, view, view.X+x, view.Y+y, w, h)

	//newView.parent = view
	Debug("NewView = (%d,%d,%d,%d)", newView.X, newView.Y, newView.W, newView.H)
	return &newView
}

//func setPixel(x int, buffer it8951.DataBuffer, offset int, ppw int, color uint16) {
//	//Debug("buffer cap %d, offset %d", cap(buffer), offset)
//	buffer[offset] = color << (x % 2)
//	//buffer[offset] &= (color << ((16 / ppw) * ppw * (x % ppw))) | ^(((1 << (16 / ppw)) - 1) << (ppw * (x % ppw)))
//	//buffer[offset] |= (color << ((16 / ppw) * ppw * (x % ppw))) | ((1<<(16/ppw))-1)<<(ppw*(x%ppw))
//}

func (view *View) DrawHLine(x0, y0, w int, stroke int, fgColor it8951.Color) *View {
	view.FillRectangle(x0, y0, w, stroke, stroke, fgColor, fgColor)
	return view
}

func (view *View) DrawVLine(x0, y0, h int, stroke int, fgColor it8951.Color) *View {
	view.FillRectangle(x0, y0, stroke, h, stroke, fgColor, fgColor)
	return view
}

func (view *View) DrawLine(x0, y0, x1, y1 int, stroke int, fgColor it8951.Color) *View {
	if x0 == x1 {
		return view.DrawVLine(x0, y0, y1-y0+1, stroke, fgColor)
	} else if y0 == y1 {
		return view.DrawHLine(x0, y0, x1-x0+1, stroke, fgColor)
	} else if x0 == x1 && y0 == y1 {
		return view
	}

	a := float64(y1-y0) / float64(x1-x0)
	b := float64(y0) - a*float64(x0)

	if x1 < x0 {
		x0, x1 = x1, x0
	}
	for x := x0; x <= x1; x++ {
		y := int(math.Round(a*float64(x) + b))
		ymin := y - stroke/2
		ymax := y + stroke/2
		if ymin == ymax {
			ymax++
		}
		for i := ymin; i < ymax; i++ {
			view.buffer.writePixel(x, y+i, uint16(fgColor))
		}
	}
	if y1 < y0 {
		y0, y1 = y1, y0
	}
	for y := y0; y < y1; y++ {
		x := int(math.Round((float64(y) - b) / a))
		xmin := x - stroke/2
		xmax := x + stroke/2
		if xmin == xmax {
			xmax++
		}
		for i := xmin; i < xmax; i++ {
			view.buffer.writePixel(x+i, y, uint16(fgColor))
		}
	}

	return view
}

func (view *View) Fill(stroke int, bgColor it8951.Color, fgColor it8951.Color) *View {
	Debug("Fill view")
	view.FillRectangle(0, 0, view.InnerW, view.InnerH, stroke, bgColor, fgColor)
	view.BgColor = uint16(bgColor)
	return view
}

func (view *View) FillRounded(stroke int, bgColor it8951.Color, fgColor it8951.Color, radius int) *View {
	Debug("Fill rounded view")
	view.FillRoundedRectangle(0, 0, view.InnerW, view.InnerH, stroke, bgColor, fgColor, radius)
	view.BgColor = uint16(bgColor)
	return view
}

func (view *View) FillTopRounded(stroke int, bgColor it8951.Color, fgColor it8951.Color, radius int) *View {
	Debug("Fill rounded view")
	view.FillTopRoundedRectangle(0, 0, view.InnerW, view.InnerH, stroke, bgColor, fgColor, radius)
	view.BgColor = uint16(bgColor)
	return view
}

func (view *View) Rectangle(x, y, w, h int, stroke int, fgColor it8951.Color) *View {
	view.DrawHLine(x, y, w, stroke, fgColor)
	view.DrawHLine(x, y+h-stroke-1, w, stroke, fgColor)
	view.DrawVLine(x, y, h, stroke, fgColor)
	view.DrawVLine(x+w-stroke-1, y, h, stroke, fgColor)
	return view
}

func (view *View) RoundedRectangle(x, y, w, h int, stroke int, fgColor it8951.Color, radius int) *View {
	view.DrawHLine(x+radius, y, w-2*radius, stroke, fgColor)
	view.DrawHLine(x+radius, y+h-stroke, w-2*radius, stroke, fgColor)
	view.DrawVLine(x, y+radius, h-2*radius, stroke, fgColor)
	view.DrawVLine(x+w-stroke, y+radius, h-2*radius, stroke, fgColor)
	view.drawCircleHelper(x+radius, y+radius, radius, 1, stroke, fgColor)
	view.drawCircleHelper(x+w-1-radius, y+radius, radius, 2, stroke, fgColor)
	view.drawCircleHelper(x+w-1-radius, y+h-1-radius, radius, 4, stroke, fgColor)
	view.drawCircleHelper(x+radius, y+h-1-radius, radius, 8, stroke, fgColor)
	return view
}

func (view *View) TopRoundedRectangle(x, y, w, h int, stroke int, fgColor it8951.Color, radius int) *View {
	view.DrawHLine(x+radius, y, w-2*radius, stroke, fgColor)
	view.DrawHLine(x, y+h-stroke, w, stroke, fgColor)
	view.DrawVLine(x, y+radius, h-radius, stroke, fgColor)
	view.DrawVLine(x+w-stroke, y+radius, h-radius, stroke, fgColor)
	view.drawCircleHelper(x+radius, y+radius, radius, 1, stroke, fgColor)
	view.drawCircleHelper(x+w-1-radius, y+radius, radius, 2, stroke, fgColor)
	return view
}

func (view *View) copyRectangle(x, y, w, h int) (buffer it8951.DataBuffer) {
	buffer = make(it8951.DataBuffer, w*h/2)
	Debug("Allocated %d bytes, orig cap = %d", w*h/2, cap(view.buffer.data))
	X := view.X
	Y := view.Y
	W := view.W
	H := view.H
	var src int
	var dst int
	Debug("X,Y,W,H-x,y,w,h = %d,%d,%d,%d-%d,%d,%d,%d", X, Y, W, H, x, y, w, h)
	for i := 0; i < h; i++ {
		for j := 0; j < w; j++ {
			src = (y+i)*W/2 + (x+j)/2
			dst = i*(w/2) + j/2
			//Debug("src=%d,dst=%d", src, dst)
			//buffer[dst] |= (window.buffer[src] >> (8 * ((j + x) % 2))) << (8 * (j % 2))
			buffer[dst] = view.buffer.data[src]
			//if w*h/2 < 10000 {
			//	Debug("orig (i=%d,j=%d) = %04x", i, j, buffer[dst])
			//}
		}
	}
	Debug("Copy OK")
	return buffer
}

func (view *View) updateRectangle(buffer it8951.DataBuffer, x, y, w, h int) {
	Debug("Allocated %d bytes, orig cap = %d", w*h/2, cap(view.buffer.data))
	X := view.X
	Y := view.Y
	W := view.W
	H := view.H
	var src int
	var dst int
	Debug("X,Y,W,H-x,y,w,h = %d,%d,%d,%d-%d,%d,%d,%d", X, Y, W, H, x, y, w, h)
	for i := 0; i < h; i++ {
		for j := 0; j < w; j++ {
			src = (y+i)*W/2 + (x+j)/2
			dst = i*(w/2) + j/2
			//Debug("src=%d,dst=%d", src, dst)
			//buffer[dst] |= (window.buffer[src] >> (8 * ((j + x) % 2))) << (8 * (j % 2))
			view.buffer.data[src] = buffer[dst]
			//if w*h/2 < 10000 {
			//	Debug("orig (i=%d,j=%d) = %04x", i, j, buffer[dst])
			//}
		}
	}
	Debug("Update OK")
}

//	func Rectangle(x,y,w,h int, stroke int, bpp int, bgColor it8951.Color, fgColor it8951.Color) {
//		DrawHLine(x, y, w, stroke, bpp, fgColor)
//		DrawHLine(x, y+h-uint16(stroke)-1, w, stroke, bpp, fgColor)
//		DrawVLine(x, y+uint16(stroke), h-2*uint16(stroke), stroke, bpp, fgColor)
//		DrawVLine(x+w-uint16(stroke)-1, y+uint16(stroke), h-2*uint16(stroke), stroke, bpp, fgColor)
//	}
func (view *View) FillRectangle(x, y, w, h int, stroke int, bgColor it8951.Color, fgColor it8951.Color) *View {
	Debug("FillRectangle %d,%d,%d,%d stroke %d", x, y, w, h, stroke)
	buffer := view.buffer
	var color uint16
	var transparent bool
	for yy := y; yy < y+h; yy++ {
		for xx := x; xx < x+w; xx++ {
			transparent = false
			if yy < y+stroke || yy >= y+h-stroke ||
				xx < x+stroke || xx >= x+w-stroke {
				color = uint16(fgColor)
			} else if bgColor != 0xffff {
				// if not transparent background, use bgColor
				color = uint16(bgColor) // << 4
			} else {
				transparent = true
			}
			if !transparent {
				buffer.writePixel(view.InnerX+xx, view.InnerY+yy, color)
			}
		}
	}
	return view
}

func (view *View) FillRoundedRectangle(x, y, w, h int, stroke int, bgColor it8951.Color, fgColor it8951.Color, radius int) *View {
	Debug("FillRoundedRectangle %d,%d,%d,%d stroke %d", x, y, w, h, stroke)
	view.FillRectangle(x+radius, y, w-2*radius, h, 0, bgColor, fgColor)
	view.fillCircleHelper(x+w-radius-1, y+radius, radius, 1, h-2*radius-1, bgColor)
	view.fillCircleHelper(x+radius, y+radius, radius, 2, h-2*radius-1, bgColor)
	view.RoundedRectangle(x, y, w, h, stroke, fgColor, radius)
	return view
}

func (view *View) FillTopRoundedRectangle(x, y, w, h int, stroke int, bgColor it8951.Color, fgColor it8951.Color, radius int) *View {
	Debug("FillRoundedRectangle %d,%d,%d,%d stroke %d", x, y, w, h, stroke)
	view.FillRectangle(x+radius, y, w-2*radius, h, 0, bgColor, fgColor)
	view.fillCircleHelper(x+w-radius-1, y+radius, radius, 1, h-radius-1, bgColor)
	view.fillCircleHelper(x+radius, y+radius, radius, 2, h-radius-1, bgColor)
	view.TopRoundedRectangle(x, y, w, h, stroke, fgColor, radius)
	return view
}

// drawCircleHelper draws an arc given the center, the radius and the corner(s) to draw
// cornerName's bits:
// - 0 top left
// - 1 top right
// - 2 bottom right
// - 3 bottom left
func (view *View) drawCircleHelper(x0, y0, r int, corners int, stroke int, color it8951.Color) {
	c := uint16(color)
	//r := radius
	for s := stroke; s > 0; s-- {
		f := 1 - r
		ddFx := 1
		ddFy := -2 * r
		x := 0
		y := r

		for x < y {
			if f >= 0 {
				y--
				ddFy += 2
				f += ddFy
			}
			x++
			ddFx += 2
			f += ddFx
			if corners&0x4 != 0 {
				view.buffer.writePixel(view.InnerX+x0+x-s, view.InnerY+y0+y-s, c)
				view.buffer.writePixel(view.InnerX+x0+y-s, view.InnerY+y0+x-s, c)
			}
			if corners&0x2 != 0 {
				view.buffer.writePixel(view.InnerX+x0+x-s, view.InnerY+y0-y+s, c)
				view.buffer.writePixel(view.InnerX+x0+y-s, view.InnerY+y0-x+s, c)
			}
			if corners&0x8 != 0 {
				view.buffer.writePixel(view.InnerX+x0-y+s, view.InnerY+y0+x-s, c)
				view.buffer.writePixel(view.InnerX+x0-x+s, view.InnerY+y0+y-s, c)
			}
			if corners&0x1 != 0 {
				view.buffer.writePixel(view.InnerX+x0-y+s, view.InnerY+y0-x+s, c)
				view.buffer.writePixel(view.InnerX+x0-x+s, view.InnerY+y0-y+s, c)
			}
		}
	}
}

// same as drawCircleHelper but fill the arc
func (view *View) fillCircleHelper(x0, y0, r int,
	corners int, delta int,
	color it8951.Color) {
	c := uint16(color)

	f := 1 - r
	ddFx := 1
	ddFy := -2 * r
	x := 0
	y := r
	px := x
	py := y

	delta++ // Avoid some +1's in the loop

	for x < y {
		if f >= 0 {
			y--
			ddFy += 2
			f += ddFy
		}
		x++
		ddFx += 2
		f += ddFx
		// These checks avoid double-drawing certain lines, important
		// for the SSD1306 library which has an INVERT drawing mode.
		if x < (y + 1) {
			if corners&1 != 0 {
				view.buffer.writeFastVLine(view.InnerX+x0+x, view.InnerY+y0-y, 2*y+delta, c)
			}
			if corners&2 != 0 {
				view.buffer.writeFastVLine(view.InnerX+x0-x, view.InnerY+y0-y, 2*y+delta, c)
			}
		}
		if y != py {
			if corners&1 != 0 {
				view.buffer.writeFastVLine(view.InnerX+x0+py, view.InnerY+y0-px, 2*px+delta, c)
			}
			if corners&2 != 0 {
				view.buffer.writeFastVLine(view.InnerX+x0-py, view.InnerY+y0-px, 2*px+delta, c)
			}
			py = y
		}
		px = x
	}
}

func (view *View) Update() *View {
	Debug("Updating view (%d,%d,%d,%d)", view.X, view.Y, view.W, view.H)
	return view.Refresh(view.InnerX, view.InnerY, view.InnerW, view.InnerH)
}

func (view *View) Refresh(x, y, w, h int) *View {
	//UpdateScreen(&ScreenUpdate{
	//	view,
	//	x, y, w, h,
	//	Refresh,
	//	nil,
	//	0, 0,
	//	0,
	//	0,
	//})

	imageInfo := it8951.LoadImgInfo{
		EndianType:       it8951.LoadImgLittleEndian,
		PixelFormat:      pixelFormat(view.buffer.bpp),
		Rotate:           it8951.Rotate0,
		SourceBufferAddr: view.buffer.data,
		TargetMemAddr:    DeviceInfo.TargetAddress(),
	}

	areaInfo := it8951.AreaImgInfo{
		X: uint16(view.X),
		Y: uint16(view.Y),
		W: uint16(view.W),
		H: uint16(view.H),
	}
	imageInfo.HostAreaPackedPixelWrite(areaInfo, view.buffer.bpp, true)
	mode := it8951.GC16Mode
	if view.buffer.bpp == 1 {
		mode = it8951.A2Mode
		it8951.Display1bpp(uint16(x), uint16(y), uint16(w), uint16(h), mode, DeviceInfo.TargetAddress(), 0xff, 0x00)
	} else {
		it8951.DisplayArea(uint16(x), uint16(y), uint16(w), uint16(h), mode)
		it8951.WaitForDisplayReady()
	}
	return view
}

func (view *View) DrawCentered(innerView *View, bpp int) {
	//innerView.X = (view.X + (view.w-innerView.w)/2) / 16 * 16 // ensure word aligned
	//innerView.Y = view.Y + (view.h-innerView.h)/2
	view.Draw(innerView, (view.W-innerView.W)/2, (view.H-innerView.H)/2, bpp)
}
func (view *View) Draw(innerView *View, xOffset, yOffset int, bpp int) {
	ppw := 16 / bpp

	innerView.X = view.X + xOffset
	innerView.Y = view.Y + yOffset
	Debug("Loading innerView (%d,%d,%d,%d) inside view (%d,%d,%d,%d)",
		innerView.X, innerView.Y, innerView.W, innerView.H,
		view.X, view.Y, view.W, view.H)
	xend := innerView.X + innerView.W - 1
	xmin := (innerView.X / ppw) * ppw
	xmax := (xend/ppw+1)*ppw - 1
	ww := (xmax - xmin + 1) / ppw
	temp := make(it8951.DataBuffer, ww*innerView.H)
	for i := 0; i < innerView.H; i++ {
		for x := xmin; x <= xmax; x++ {
			if x < innerView.X || x > xend {
				continue
			}
			temp[i*ww+(x-xmin)/ppw] |=
				innerView.buffer.data[(i*innerView.W+(x-innerView.X))/ppw] >> (bpp * ((xmin - innerView.X) % ppw))
		}
	}
	imageInfo := it8951.LoadImgInfo{
		EndianType:       it8951.LoadImgLittleEndian,
		PixelFormat:      pixelFormat(bpp),
		Rotate:           it8951.Rotate0,
		SourceBufferAddr: temp,
		TargetMemAddr:    DeviceInfo.TargetAddress(),
	}
	areaInfo := it8951.AreaImgInfo{
		X: uint16(innerView.X),
		Y: uint16(innerView.Y),
		W: uint16(innerView.W),
		H: uint16(innerView.H),
	}
	imageInfo.HostAreaPackedPixelWrite(areaInfo, bpp, true)
	it8951.DisplayArea(uint16(innerView.X), uint16(innerView.Y), uint16(innerView.W), uint16(innerView.H), it8951.GC16Mode)
}

func (view *View) LoadBitmapAt(x, y int, name string, bpp int) (*View, error) {
	bitmapView, err := view.LoadBitmap(name, bpp)
	if err != nil {
		return nil, err
	}
	bitmapView.X = view.X + x
	bitmapView.Y = view.Y + y
	bitmapView.InnerX = bitmapView.X
	bitmapView.InnerY = bitmapView.Y
	return bitmapView, err
}

func (view *View) LoadBitmapVCenteredAt(x int, name string, bpp int) (*View, error) {
	bitmapView, err := view.LoadBitmap(name, bpp)
	if err != nil {
		return nil, err
	}
	bitmapView.X = view.X + x
	bitmapView.Y = (view.H - bitmapView.H) / 2
	bitmapView.InnerX = bitmapView.X
	bitmapView.InnerY = bitmapView.Y
	return bitmapView, err
}

func (view *View) LoadBitmapCentered(name string, bpp int) (*View, error) {
	bitmapView, err := view.LoadBitmap(name, bpp)
	if err != nil {
		return nil, err
	}
	bitmapView.X = (view.W - bitmapView.W) / 2
	bitmapView.Y = (view.H - bitmapView.H) / 2
	bitmapView.InnerX = bitmapView.X
	bitmapView.InnerY = bitmapView.Y
	return bitmapView, err
}

func (view *View) LoadBitmap(name string, bpp int) (*View, error) {
	Debug("Loading Bitmap [%s]", name)
	bmp, err := Files.ReadFile(name)
	var newView *View
	if err != nil {
		Debug("Error loading Bitmap: %v", err)
		return newView, err
	}
	if bmp[0] != 'B' || bmp[1] != 'M' {
		Debug("Not a bitmap")
		return newView, errors.New("not a bitmap")
	}
	if (uint16(bmp[26]) | (uint16(bmp[27]) << 8)) != 1 {
		Debug("Not 1 color plane")
		return newView, errors.New("not 1 color plane")
	}
	if (uint16(bmp[28]) | (uint16(bmp[29]) << 8)) != 16 {
		Debug("Not 16 bits")
		return newView, errors.New("not 16 bits")
	}
	// if ((bmp[30] | (bmp[31] << 8) | (bmp[32] << 16) | (bmp[33] << 24)) != 0) {
	//     Debug( "Compressed")
	//     return NULL
	// }
	w := int(binary.LittleEndian.Uint32(bmp[18:22]))
	h := int(binary.LittleEndian.Uint32(bmp[22:26]))
	newView = view.NewView(0,
		0,
		w,
		h,
		bpp)
	offset := int(binary.LittleEndian.Uint32(bmp[10:14]))
	Debug("Loading bitmap: (%d x %d, offset=%d)", w, h, offset)
	src := 0
	// reminder: image is upside-down
	for y := 0; y < newView.H; y++ {
		for x := 0; x < newView.W; x++ {
			color := binary.LittleEndian.Uint16(bmp[offset+src : offset+src+2])
			newView.buffer.writePixel(newView.X+x, newView.Y+newView.H-y-1, RGBToBpp(color, bpp))
			//dst++
			src += 2
		}
	}
	return newView, nil
}

func colorToBpp(color uint16, bpp int) uint16 {
	if bpp == 1 {
		if color == 0 {
			return 0x00
		}
		return 0xff
	}

	return color & ((1 << bpp) - 1)
}

func RGBToBpp(color uint16, bpp int) uint16 {

	r := (color >> 11) & 0x1f
	g := (color >> 5) & 0x3f
	b := color & 0x1f

	grey := 0.299*float64(r*255/31) + 0.587*float64(g*255/63) + 0.114*float64(b*255/31)

	res := uint16(grey) >> (8 - bpp)
	//res := uint16(math.Floor(grey)) * ((1 << bpp) - 1) / 255
	//Debug("color %d => %d", color, res)
	return res
}

func (buffer *Buffer) writePixel(x, y int, fgColor uint16) {
	//defer func() {
	//	if err := recover(); err != nil {
	//		Debug("error on pixel (%d,%d) in buffer(%d,%d,%d,%d)",
	//			x, y, buffer.X, buffer.Y, buffer.ww, buffer.wh)
	//	}
	//}()

	//Debug("writing pixel (%d,%d) -> (%d,%d)", x, y, x-buffer.X, y-buffer.Y)
	bpp := buffer.bpp
	if bpp == 1 {
		bpp = 8 // 1bpp data is stored as 8bpp
	}
	ppw := 16 / bpp

	if y-buffer.Y < 0 || y-buffer.Y >= buffer.wh || (x-buffer.X)/ppw < 0 || (x-buffer.X)/ppw >= buffer.ww {
		Debug("error writing pixel (%d,%d) in buffer(%d,%d,%d,%d)",
			x, y, buffer.X, buffer.Y, buffer.ww, buffer.wh)
		return
	}
	color := colorToBpp(fgColor, buffer.bpp) << (bpp * ((x - buffer.X) % ppw))
	mask := uint16(^(((1 << bpp) - 1) << (bpp * ((x - buffer.X) % ppw))))
	buffer.data[((y-buffer.Y)*buffer.ww)+(x-buffer.X)/ppw] &= mask
	buffer.data[((y-buffer.Y)*buffer.ww)+(x-buffer.X)/ppw] |= color
}

func (buffer *Buffer) readPixel(x, y int) (fgColor uint16) {
	//defer func() {
	//	if err := recover(); err != nil {
	//		Debug("error reading pixel (%d,%d) in buffer(%d,%d,%d,%d)",
	//			x, y, buffer.X, buffer.Y, buffer.ww, buffer.wh)
	//	}
	//}()
	bpp := buffer.bpp
	if bpp == 1 {
		bpp = 8 // 1bpp data is stored as 8bpp
	}
	ppw := 16 / bpp

	if y-buffer.Y < 0 || y-buffer.Y >= buffer.wh || (x-buffer.X)/ppw < 0 || (x-buffer.X)/ppw >= buffer.ww {
		//Debug("error writing pixel (%d,%d) in buffer(%d,%d,%d,%d)",
		//	x, y, buffer.X, buffer.Y, buffer.ww, buffer.wh)
		return
	}

	//Debug("writing pixel (%d,%d) -> (%d,%d)", x, y, x-buffer.X, y-buffer.Y)
	//if y-buffer.Y < 0 || y-buffer.Y >= buffer.wh || (x-buffer.X)/ppw < 0 || (x-buffer.X)/ppw >= buffer.ww {
	//	return
	//}
	color := buffer.data[((y-buffer.Y)*buffer.ww)+(x-buffer.X)/ppw]
	if buffer.bpp != 1 {
		fgColor = color >> (bpp * ((x - buffer.X) % ppw))
	} else {
		fgColor = 0x00
		if color != 0xff {
			fgColor = 0xff
		}
	}
	return fgColor
}

func (buffer *Buffer) writeFillRectangle(x, y, w, h int, fgColor uint16) {
	Debug("Write rectangle (%d,%d,%d,%d)", x, y, w, h)
	//bpp := buffer.bpp
	//ppw := 16 / bpp
	for i := 0; i < w; i++ {
		for j := 0; j < h; j++ {
			buffer.writePixel(x, y, fgColor)
			//color := colorToBpp(fgColor, bpp) << (bpp * ((x + i - buffer.X) % ppw))
			//mask := uint16(^(((1 << bpp) - 1) << (bpp * ((x + i - buffer.X) % ppw))))
			//buffer.data[((y-buffer.Y+j)*buffer.ww)+(x+i-buffer.X)/ppw] &= mask
			//buffer.data[((y-buffer.Y+j)*buffer.ww)+(x+i-buffer.X)/ppw] |= color
		}
	}
}

func (buffer *Buffer) writeFastVLine(x, y, h int, color uint16) {
	for i := 0; i < h; i++ {
		buffer.writePixel(x, y+i, color)
	}
}

func (buffer *Buffer) writeFastHLine(x, y, w int, color uint16) {
	for i := 0; i < w; i++ {
		buffer.writePixel(x+i, y, color)
	}
}

func pixelFormat(bpp int) it8951.PixelMode {
	switch bpp {
	case 1:
		return it8951.BPP8
	case 2:
		return it8951.BPP2
	case 3:
		return it8951.BPP3
	case 4:
		return it8951.BPP4
	case 8:
		return it8951.BPP8
	}
	return it8951.BPP8
}

func (view *View) CopyPixels(dx, dy int, source *View, sx, sy, sw, sh int) {
	//for _, parent := range source.Views {
	//	if parent != view {
	//		view.CopyPixels(dx, dy, parent, sx, sy, sw, sh)
	//	}
	//}
	for y := 0; y < sh; y++ {
		for x := 0; x < sw; x++ {
			if dx+x >= view.X && dy+y >= view.Y && dx+x < view.X+view.W && dy+y < view.Y+view.H &&
				sx+x >= source.X && sy+y >= source.Y && sx+x < source.X+source.W && sy+y < source.Y+source.H {
				view.buffer.writePixel(dx+x, dy+y, source.buffer.readPixel(sx+x, sy+y))
			}
		}
	}

	//UpdateScreen(&ScreenUpdate{
	//	view,
	//	dx, dy, sw, sh,
	//	Copy,
	//	source,
	//	sx, sy,
	//	0, 0,
	//})
}

func (view *View) CopyPixelsWithTransparency(dx, dy int, source *View, sx, sy, sw, sh int, transparency float64, bgColor uint16) {
	for x := 0; x < sw; x++ {
		for y := 0; y < sh; y++ {
			if dx+x >= view.X && dy+y >= view.Y && dx+x < view.X+view.W && dy+y < view.Y+view.H &&
				sx+x >= source.X && sy+y >= source.Y && sx+x < source.X+source.W && sy+y < source.Y+source.H {
				sColor := source.buffer.readPixel(sx+x, sy+y)
				dColor := min(0xf, uint16(math.Round(float64(bgColor&0xf)*(1-transparency)+float64(sColor&0xf)*transparency)))
				view.buffer.writePixel(dx+x, dy+y, dColor)
			}
		}
	}

	//UpdateScreen(&ScreenUpdate{
	//	view,
	//	dx, dy, sw, sh,
	//	CopyTransparent,
	//	source,
	//	sx, sy,
	//	transparency,
	//	bgColor,
	//})
}
