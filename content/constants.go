/*
   content,
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
	"nv/display"
	"nv/display/fonts-go"
)

const (
	cursorWidth                 = 2
	cursorHeight                = 40
	cursorOnDuration            = 500 // 1s
	cursorOffDuration           = 500 // 1s
	cursorRestartDelay          = 300
	defaultParagraphIndent      = true
	defaultParagraphSpacing     = true
	defaultParagraphIndentValue = "  "
	defaultScrollBarWidth       = 30
	defaultScrollBarBgColor     = display.Gray14
	defaultScrollBarColor       = display.Gray1
)

var (
	regularFont    = &fonts.UbuntuSans_Regular20pt8b
	boldFont       = &fonts.UbuntuSans_Bold20pt8b
	italicFont     = &fonts.UbuntuSans_Italic20pt8b
	boldItalicFont = &fonts.UbuntuSans_BoldItalic20pt8b

	panelFont = &fonts.Montserrat_Medium16pt8b
)
