/*
   input,
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

package input

var (
	keymapEnUS = map[string][]rune{
		"KEY_A":          {'A', 'a'},
		"KEY_B":          {'B', 'b'},
		"KEY_C":          {'C', 'c'},
		"KEY_D":          {'D', 'd'},
		"KEY_E":          {'E', 'e'},
		"KEY_F":          {'F', 'f'},
		"KEY_G":          {'G', 'g'},
		"KEY_H":          {'H', 'h'},
		"KEY_I":          {'I', 'i'},
		"KEY_J":          {'J', 'j'},
		"KEY_K":          {'K', 'k'},
		"KEY_L":          {'L', 'l'},
		"KEY_M":          {'M', 'm'},
		"KEY_N":          {'N', 'n'},
		"KEY_O":          {'O', 'o'},
		"KEY_P":          {'P', 'p'},
		"KEY_Q":          {'Q', 'q'},
		"KEY_R":          {'R', 'r'},
		"KEY_S":          {'S', 's'},
		"KEY_T":          {'T', 't'},
		"KEY_U":          {'U', 'u'},
		"KEY_V":          {'V', 'v'},
		"KEY_W":          {'W', 'w'},
		"KEY_X":          {'X', 'x'},
		"KEY_Y":          {'Y', 'y'},
		"KEY_Z":          {'Z', 'z'},
		"KEY_0":          {'0', ')'},
		"KEY_1":          {'1', '!'},
		"KEY_2":          {'2', '@'},
		"KEY_3":          {'3', '#'},
		"KEY_4":          {'4', '$'},
		"KEY_5":          {'5', '%'},
		"KEY_6":          {'6', '^'},
		"KEY_7":          {'7', '&'},
		"KEY_8":          {'8', '*'},
		"KEY_9":          {'9', '('},
		"KEY_MINUS":      {'-', '_'},
		"KEY_EQUAL":      {'=', '+'},
		"KEY_GRAVE":      {'`', '~'},
		"KEY_LEFTBRACE":  {'[', '{'},
		"KEY_RIGHTBRACE": {']', '}'},
		"KEY_BACKSLASH":  {'\\' | '}'},
		"KEY_SEMICOLON":  {';', ':'},
		"KEY_APOSTROPHE": {'\'', '"'},
		"KEY_COMMA":      {',', '<'},
		"KEY_DOT":        {'.', '>'},
		"KEY_SLASH":      {'/', '?'},
	}
)
