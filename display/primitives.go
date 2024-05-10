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
	X, Y, W, H                     int
	InnerX, InnerY, InnerW, InnerH int
	BgColor                        uint16
	buffer                         Buffer
	TextArea                       TextArea
	content                        Content
	Xb, Yb, Wb, Hb                 int
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
	ppw := 16 / bpp
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
	Debug("NewView (%d,%d,%d,%d)", x, y, w, h)
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
	view.FillRectangle(0, 0, view.W, view.H, stroke, bgColor, fgColor)
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
			} else if bgColor != 0x00 {
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

func (view *View) Update() *View {
	Debug("Updating view (%d,%d,%d,%d)", view.X, view.Y, view.W, view.H)
	return view.Refresh(view.InnerX, view.InnerY, view.InnerW, view.InnerH)
}

func (view *View) Refresh(x, y, w, h int) *View {
	imageInfo := it8951.LoadImgInfo{
		EndianType:       it8951.LoadImgSmallEndian,
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
	}
	it8951.DisplayArea(uint16(x), uint16(y), uint16(w), uint16(h), mode)
	it8951.WaitForDisplayReady()
	return view
}

func (view *View) DrawCentered(innerView View, bpp int) {
	//innerView.X = (view.X + (view.w-innerView.w)/2) / 16 * 16 // ensure word aligned
	//innerView.Y = view.Y + (view.h-innerView.h)/2
	view.Draw(innerView, (view.W-innerView.W)/2, (view.H-innerView.H)/2, bpp)
}
func (view *View) Draw(innerView View, xOffset, yOffset int, bpp int) {
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
		EndianType:       it8951.LoadImgSmallEndian,
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

func LoadBitmap(name string, bpp int) (View, error) {
	ppw := 16 / bpp
	Debug("Loading Bitmap [%s]", name)
	bmp, err := Files.ReadFile(name)
	var view View
	if err != nil {
		Debug("Error loading Bitmap: %v", err)
		return view, err
	}
	if bmp[0] != 'B' || bmp[1] != 'M' {
		Debug("Not a bitmap")
		return view, errors.New("not a bitmap")
	}
	if (bmp[26] | (bmp[27] << 8)) != 1 {
		Debug("Not 1 color plane")
		return view, errors.New("not 1 color plane")
	}
	if (bmp[28] | (bmp[29] << 8)) != 16 {
		Debug("Not 16 bits")
		return view, errors.New("not 16 bits")
	}
	// if ((bmp[30] | (bmp[31] << 8) | (bmp[32] << 16) | (bmp[33] << 24)) != 0) {
	//     Debug( "Compressed");
	//     return NULL;
	// }
	view.W = int(binary.LittleEndian.Uint32(bmp[18:22]))
	view.H = int(binary.LittleEndian.Uint32(bmp[22:26]))
	view.X = 0
	view.Y = 0
	offset := int(binary.LittleEndian.Uint32(bmp[10:14]))
	Debug("Loading bitmap: (%d x %d, offset=%d)", view.W, view.H, offset)
	src := 0
	ww := (view.W / ppw) * view.H
	dst := 0
	view.buffer.data = make(it8951.DataBuffer, ww)
	for y := 0; y < view.H; y++ {
		for x := 0; x < view.W; x++ {
			color := binary.LittleEndian.Uint16(bmp[offset+src : offset+src+2])
			gray := colorToBpp(color, bpp)
			colorMask := uint16(^(((1 << bpp) - 1) << (bpp * (x % ppw))))
			grayMask := uint16(gray << (bpp * (x % ppw)))
			dst = ((view.H-y-1)*(view.W) + x) / ppw
			view.buffer.data[dst] &= colorMask
			view.buffer.data[dst] |= grayMask
			//dst++
			src += 2
		}
	}
	return view, nil
}

func colorToBpp(color uint16, bpp int) uint16 {
	//if bpp == 1 && color == 0xffff {
	//	return (1 << bpp) - 1 // white
	//} else if bpp == 1 {
	//	return 0
	//}

	if bpp < 8 {
		return color & ((1 << bpp) - 1)
	}

	r := (color >> 11) & 0x1f
	g := (color >> 5) & 0x3f
	b := color & 0x1f

	grey := 0.299*float64(r*255/31) + 0.587*float64(g*255/63) + 0.114*float64(b*255/31)

	res := uint16(math.Floor(grey)) * ((1 << bpp) - 1) / 255
	//Debug("color %d => %d", color, res)
	return res
}

func (buffer *Buffer) writePixel(x, y int, fgColor uint16) {
	defer func() {
		if err := recover(); err != nil {
			Debug("error on pixel (%d,%d) in buffer(%d,%d,%d,%d)",
				x, y, buffer.X, buffer.Y, buffer.ww, buffer.wh)
		}
	}()
	//Debug("writing pixel (%d,%d) -> (%d,%d)", x, y, x-buffer.X, y-buffer.Y)
	bpp := buffer.bpp
	ppw := 16 / bpp

	if y-buffer.Y < 0 || y-buffer.Y >= buffer.wh || (x-buffer.X)/ppw < 0 || (x-buffer.X)/ppw >= buffer.ww {
		Debug("error writing pixel (%d,%d) in buffer(%d,%d,%d,%d)",
			x, y, buffer.X, buffer.Y, buffer.ww, buffer.wh)
		return
	}
	color := colorToBpp(fgColor, bpp) << (bpp * ((x - buffer.X) % ppw))
	mask := uint16(^(((1 << bpp) - 1) << (bpp * ((x - buffer.X) % ppw))))
	buffer.data[((y-buffer.Y)*buffer.ww)+(x-buffer.X)/ppw] &= mask
	buffer.data[((y-buffer.Y)*buffer.ww)+(x-buffer.X)/ppw] |= color
}

func (buffer *Buffer) readPixel(x, y int) (fgColor uint16) {
	defer func() {
		if err := recover(); err != nil {
			Debug("error reading pixel (%d,%d) in buffer(%d,%d,%d,%d)",
				x, y, buffer.X, buffer.Y, buffer.ww, buffer.wh)
		}
	}()
	//Debug("writing pixel (%d,%d) -> (%d,%d)", x, y, x-buffer.X, y-buffer.Y)
	bpp := buffer.bpp
	ppw := 16 / bpp
	//if y-buffer.Y < 0 || y-buffer.Y >= buffer.wh || (x-buffer.X)/ppw < 0 || (x-buffer.X)/ppw >= buffer.ww {
	//	return
	//}
	color := buffer.data[((y-buffer.Y)*buffer.ww)+(x-buffer.X)/ppw]
	fgColor = color >> (bpp * ((x - buffer.X) % ppw))
	return fgColor
}

func (buffer *Buffer) writeFillRectangle(x, y, w, h int, fgColor uint16) {
	Debug("Write rectangle (%d,%d,%d,%d)", x, y, w, h)
	bpp := buffer.bpp
	ppw := 16 / bpp
	for i := 0; i < w; i++ {
		for j := 0; j < h; j++ {
			color := colorToBpp(fgColor, bpp) << (bpp * ((x + i - buffer.X) % ppw))
			mask := uint16(^(((1 << bpp) - 1) << (bpp * ((x + i - buffer.X) % ppw))))
			buffer.data[((y-buffer.Y+j)*buffer.ww)+(x+i-buffer.X)/ppw] &= mask
			buffer.data[((y-buffer.Y+j)*buffer.ww)+(x+i-buffer.X)/ppw] |= color
		}
	}
}

func (buffer *Buffer) write(x, y, w, h int) {
	bpp := buffer.bpp
	Debug("Write (%d,%d,%d,%d) in buffer(%d,%d,%d,%d)", x, y, w, h, buffer.X, buffer.Y, buffer.ww*bpp, buffer.wh)
	imageInfo := it8951.LoadImgInfo{
		EndianType:       it8951.LoadImgSmallEndian,
		PixelFormat:      pixelFormat(bpp),
		Rotate:           it8951.Rotate0,
		SourceBufferAddr: buffer.data,
		TargetMemAddr:    DeviceInfo.TargetAddress(),
	}
	areaInfo := it8951.AreaImgInfo{
		X: uint16(x),
		Y: uint16(y),
		W: uint16(w),
		H: uint16(h),
	}
	imageInfo.HostAreaPackedPixelWrite(areaInfo, bpp, true)
	it8951.DisplayArea(uint16(x), uint16(y), uint16(w), uint16(h), it8951.A2Mode)
	it8951.WaitForDisplayReady()
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
