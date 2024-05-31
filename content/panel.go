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
)

type Panel struct {
	view *display.View
	InputFields
}

type InputFields []*InputField

type List interface {
	GetValues(done chan<- bool)
}

type InputField struct {
	name      string
	format    string
	fieldType FieldType
	value     interface{}
	values    List
}

type FieldType int

const (
	TextField FieldType = iota
	PasswordField
	OptionField
	CheckboxField
	SelectField
)

func (panel *Panel) Init(view *display.View) (views []*display.View) {
	panel.view = view
	view.
		SetTextArea(panelFont, 10, 10).
		Update()
	return append(views, view)
}

func (panel *Panel) GetTitle() string {
	return "Panel"
}
func (panel *Panel) Load() {

}

func (panel *Panel) Refresh() {}
func (panel *Panel) Save()    {}
func (panel *Panel) Print() {
	view := panel.view
	for i, field := range panel.InputFields {
		view.WriteAt(0, i*int(view.TextArea.Font.YAdvance), field.name, display.Black, display.White)
	}
	view.Update()
}

func (panel *Panel) Type() string {
	return "Panel"
}
