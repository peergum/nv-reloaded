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

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"io"
	"log"
	"os"
	"os/exec"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type Keyboards []Keyboard

type Keyboard struct {
	File        string
	Name        string
	Vendor      string
	Events      EventTypes
	doneChannel chan bool
}

type EventTypes map[int]*EventType

type EventType struct {
	Name  string
	Codes EventCodes
}

type EventCodes map[int]*EventCode

type EventCode struct {
	Name  string
	State int
}

type KeyEvent struct {
	Sec         uint64
	USec        uint64
	Type        int
	Keycode     int
	Value       rune
	TypeName    string
	KeyName     string
	Char        rune
	SpecialKeys bool
	MetaKeys    uint16
}

const (
	Shift uint16 = 1 << iota
	Ctrl
	Alt
	Meta
	CapsLock
	None uint16 = 0
)

const (
	KEY_SHIFT = 1 << iota
	KEY_ALTGR
	KEY_CTRL
	KEY_ALT
	KEY_LEFTSHIFT
	KEY_RIGHTSHIFT
	KEY_LEFTCTRL
	KEY_RIGHTCTRL
	LED_CAPSL // caps lock on or off
	KEY_LEFTALT
	KEY_RIGHTALT
	KEY_LEFTMETA
	KEY_RIGHTMETA
)

var (
	eventChannel    <-chan bool                 // receives termination event for all keyboards
	keyboardChannel = make(chan *Keyboard, 5)   // informs about new keyboards
	keyChannel      = make(chan *KeyEvent, 100) // informs key changes
	metakeys        uint16
	keyboards       Keyboards
)

func KeyChannel() <-chan *KeyEvent {
	return keyChannel
}

func KeyboardChannel() <-chan *Keyboard {
	return keyboardChannel
}

func Metakeys() uint16 {
	return metakeys
}

func (keyboard *Keyboard) ReadKeyboard() error {
	Debug("reading keyboard %s at %s", keyboard.Name, keyboard.File)
	file, err := os.Open(keyboard.File)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)
	reader := bufio.NewReader(file)
	buffer := make([]byte, 24)
	keyboardReadTicker := time.NewTicker(time.Duration(5) * time.Millisecond)
	defer keyboardReadTicker.Stop()
	done := false
	for !done {
		select {
		case <-keyboard.doneChannel:
			// if we receive anything here, it means keyboard's gone
			return nil
		case event := <-eventChannel:
			if event {
				done = true
			}
		case <-keyboardReadTicker.C: // check keyboard
			for i := 0; i < 24; i++ {
				buffer[i], err = reader.ReadByte()
				if err != nil {
					break
				}
			}
			if err != nil {
				break
			}
			keyEvent := KeyEvent{}
			keyEvent.Sec = binary.BigEndian.Uint64(buffer[0:8])
			keyEvent.USec = binary.BigEndian.Uint64(buffer[8:16])
			keyEvent.Type = int(binary.LittleEndian.Uint16(buffer[16:18]))
			keyEvent.Keycode = int(binary.LittleEndian.Uint16(buffer[18:20]))
			keyEvent.Value, _ = utf8.DecodeRune(buffer[20:24])
			keyEvent.TypeName = "?"
			keyEvent.KeyName = "?"
			if keyboard.Events[keyEvent.Type] != nil {
				keyEvent.TypeName = keyboard.Events[keyEvent.Type].Name
				if keyboard.Events[keyEvent.Type].Codes[keyEvent.Keycode] != nil {
					keyEvent.KeyName = keyboard.Events[keyEvent.Type].Codes[keyEvent.Keycode].Name
				}
			}
			//Debug("KeyEvent: %v", keyboard.Event(keyEvent))
			//_ = keyboard.Event(keyEvent)
			//Debug("%v", keyboard)

			specialKey := None
			switch keyEvent.KeyName {
			case "KEY_SHIFT", "KEY_LEFTSHIFT", "KEY_RIGHTSHIFT":
				specialKey = Shift
			case "KEY_CTRL", "KEY_LEFTCTRL", "KEY_RIGHTCTRL":
				specialKey = Ctrl
			case "KEY_ALT", "KEY_LEFTALT", "KEY_RIGHTALT":
				specialKey = Alt
			case "KEY_META", "KEY_LEFTMETA", "KEY_RIGHTMETA":
				specialKey = Meta
			case "LED_CAPSL": // caps lock on or off
				specialKey = CapsLock
			}
			if specialKey != None {
				keyEvent.SpecialKeys = true
				if keyEvent.Value == 0 {
					metakeys &= ^specialKey
				} else {
					metakeys |= specialKey
				}
			}
			keyEvent.MetaKeys = metakeys
			if keymapEnUS[keyEvent.KeyName] != nil {
				keyEvent.Char = keymapEnUS[keyEvent.KeyName][metakeys]
			}
			keyChannel <- &keyEvent
		}
	}
	Debug("We're done")
	return nil
}

func CheckKeyboard(filename string) {
	Debug("File %s", filename)
	if !strings.Contains(filename, "event") {
		Debug("Not an event file")
		return
	}
	ready := false
	for !ready {
		cmd := exec.Command("/usr/bin/udevadm", "info", filename)
		var kbd Keyboard
		if out, err := cmd.Output(); err == nil {
			scanner := bufio.NewScanner(bytes.NewReader(out))
			scanner.Split(bufio.ScanLines)
			kbd = Keyboard{
				Events: make(EventTypes),
			}
			isKeys := false
			for scanner.Scan() {
				text := scanner.Text()
				//Debug(text)
				if strings.Contains(text, "ID_VENDOR=") {
					kbd.Vendor = "" + strings.Split(text, "=")[1]
					Debug("Found vendor: %s", kbd.Vendor)
				}
				if strings.Contains(text, "ID_INPUT_KEYBOARD=1") ||
					strings.Contains(text, "ID_INPUT_KEY=1") {
					Debug("%s is a keyboard", filename)
					kbd.File = filename
					isKeys = true
				}
				if strings.Contains(text, "ID_INPUT=1") {
					// this should tell input is ready
					ready = true
				}
				if strings.Contains(text, "hdmi-event") {
					isKeys = false
				}
			}
			// wait until input is seen as valid...
			if !ready {
				continue
			}
			if !isKeys {
				Debug("%s is not a keyboard", filename)
				return
			}
			Debug("Checking keyboard %s (%s)", kbd.File, kbd.Vendor)
			cmd := exec.Command("/usr/bin/evtest", kbd.File)
			//var stdin io.WriteCloser
			var stdout io.ReadCloser
			//stdin, err = cmd.StdinPipe()
			stdout, err = cmd.StdoutPipe()
			if err = cmd.Start(); err == nil {
				scanner := bufio.NewScanner(stdout)
				scanner.Split(bufio.ScanLines)
				typeNum := -1
				var eventType *EventType
				for scanner.Scan() {
					Debug("Scanning %s", scanner.Text())
					text := scanner.Text()
					if strings.Contains(text, "Input device name:") {
						kbd.Name = strings.Clone(strings.Split(text, "\"")[1])
					} else if strings.Contains(text, "Event type") {
						eventType = &EventType{
							Codes: make(EventCodes),
						}
						ev := strings.Fields(text)
						typeNum, err = strconv.Atoi(ev[2])
						if err == nil {
							eventType.Name = "" + ev[3][1:len(ev[3])-1]
							kbd.Events[typeNum] = eventType
						} else {
							typeNum = -1
						}
					} else if typeNum >= 0 && strings.Contains(text, "Event code") {
						eventCode := &EventCode{}
						ev := strings.Fields(text)
						var codeNum int
						codeNum, err = strconv.Atoi(ev[2])
						if err == nil {
							eventCode.Name = "" + ev[3][1:len(ev[3])-1]
							if len(ev) >= 6 && ev[4] == "state" {
								eventCode.State, err = strconv.Atoi(ev[5])
							}
							eventType.Codes[codeNum] = eventCode
							kbd.Events[typeNum] = eventType
						}
					} else if strings.Contains(text, "interrupt to exit") {
						if cmd.Process.Kill() != nil {
							Debug("Error killing process")
						}
					}
				}
			} else {
				Debug("Error %v", err)
			}
			kbd.doneChannel = make(chan bool) // termination channel
			Debug("Keyboard %s added", filename)
			keyboards = append(keyboards, kbd)
			keyboardChannel <- &kbd
			go kbd.ReadKeyboard()
		}
	}
}

func Search(mainEventChannel <-chan bool) {
	var err error

	eventChannel = mainEventChannel // will receive termination from main

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("Keyboard watcher error: %s", err)
	}
	defer watcher.Close()

	err = watcher.Add("/dev/input")
	if err != nil {
		Debug("Can't check for input events")
		keyboardChannel <- nil
	}

	// initial check for keyboards
	var files []os.DirEntry
	if files, err = os.ReadDir("/dev/input"); err != nil {
		Debug("Input directory not found")
		keyboardChannel <- nil // inform error
		return
	}
	Debug("Waiting for keyboards...")
	for _, file := range files {
		if !strings.Contains(file.Name(), "event") {
			continue
		}
		CheckKeyboard("/dev/input/" + file.Name())
	}
	if len(keyboards) == 0 {
		keyboardChannel <- &Keyboard{
			File: "none",
		}
	}
SearchLoop:
	for {
		select {
		case event := <-watcher.Events:
			Debug("input event: %v", event)
			if event.Has(fsnotify.Create) {
				CheckKeyboard(event.Name)
			} else if event.Has(fsnotify.Remove) {
				kbd := slices.IndexFunc(keyboards, func(kbd Keyboard) bool {
					return kbd.File == event.Name
				})
				if kbd >= 0 {
					Debug("Keyboard %s removed", keyboards[kbd].Name)
					keyboards[kbd].doneChannel <- true
					keyboards = slices.Delete(keyboards, kbd, kbd+1)
				}
				if len(keyboards) == 0 {
					keyboardChannel <- &Keyboard{
						File: "none",
					}
				}
			}

		case <-mainEventChannel:
			break SearchLoop
		}
		//time.Sleep(time.Duration(500) * time.Millisecond)
	}
	Debug("Done with input")
}

func (events EventTypes) String() (res string) {
	var typeKeys = make([]int, 0, len(events))
	for i := range events {
		typeKeys = append(typeKeys, i)
	}
	sort.Ints(typeKeys)
	for _, i := range typeKeys {
		etype := events[i]
		var codeKeys = make([]int, 0, len(etype.Codes))
		res += fmt.Sprintf("Event %d (%x) -> %s\n", i, i, etype.Name)
		for j := range etype.Codes {
			codeKeys = append(codeKeys, j)
		}
		sort.Ints(codeKeys)

		for _, j := range codeKeys {
			ecode := etype.Codes[j]
			res += fmt.Sprintf(" Code %d (%x) -> %s (state %d|%x)\n", j, j, ecode.Name, ecode.State, ecode.State)
		}
		//res+= fmt.Sprintf("%s", etype.Codes)
	}
	return res
}

func (keyboard *Keyboard) Event(event KeyEvent) (res string) {
	etype := "?"
	code := "?"
	if keyboard.Events[event.Type] != nil {
		etype = keyboard.Events[event.Type].Name
		if keyboard.Events[event.Type].Codes[event.Keycode] != nil {
			code = keyboard.Events[event.Type].Codes[event.Keycode].Name
		}
	}
	res += fmt.Sprintf("Type: %d (%s), Code: %d (%s), Value: %d(%x)",
		event.Type,
		etype,
		event.Keycode,
		code,
		event.Value, event.Value)
	return res
}

func (keyboard *Keyboard) String() (res string) {
	res += fmt.Sprintf("keyboard: %s, Events: %v\n", keyboard.Name, keyboard.Events)
	return res
}
