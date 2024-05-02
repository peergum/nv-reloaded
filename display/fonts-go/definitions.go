/*
   fonts,
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

// Package fonts does ...
package fonts

// GfxGlyph Font data stored PER GLYPH
type GfxGlyph struct {
	BitmapOffset uint16 ///< Pointer into GFXfont->bitmap
	Width        uint16 ///< Bitmap dimensions in pixels
	Height       uint16 ///< Bitmap dimensions in pixels
	XAdvance     uint16 ///< Distance to advance cursor (x axis)
	XOffset      int16  ///< X dist from cursor pos to UL corner
	YOffset      int16  ///< Y dist from cursor pos to UL corner
}

// GfxFont Data stored for FONT AS A WHOLE
type GfxFont struct {
	Bitmap   []uint8    ///< Glyph bitmaps, concatenated
	Glyphs   []GfxGlyph ///< Glyph array
	First    uint16     ///< ASCII extents (First char)
	Last     uint16     ///< ASCII extents (Last char)
	YAdvance uint16     ///< Newline distance (y axis)
}
