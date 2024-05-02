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
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Keyboard struct {
	File   string
	Name   string
	Vendor string
	Events EventTypes
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
	Sec      uint64
	USec     uint64
	Type     int
	Keycode  int
	Value    rune
	TypeName string
	KeyName  string
	Char     rune
}

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
	keyChannel = make(chan *KeyEvent, 100)
	metaKey    uint16
)

func KeyChannel() <-chan *KeyEvent {
	return keyChannel
}

func ReadKeyboard(keyboard Keyboard, eventChannel <-chan string) error {
	log.Println("->", keyboard.Name)
	file, err := os.Open("/dev/input/" + keyboard.File)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	buffer := make([]byte, 24)
	done := false
	for !done {
		select {
		case event := <-eventChannel:
			if event == "done" {
				done = true
			}
		default:
			//Debug("Waiting for keys...")
		ReadEvent:
			for {
				keyEvent := KeyEvent{}
				for i := 0; i < 24; i++ {
					buffer[i], err = reader.ReadByte()
					if err != nil {
						break ReadEvent
					}
				}
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
				Debug("KeyEvent: %v", keyboard.Event(keyEvent))
				//_ = keyboard.Event(keyEvent)
				//Debug("%v", keyboard)

				switch keyEvent.KeyName {
				case "KEY_LEFTSHIFT":
				case "KEY_RIGHTSHIFT":
				case "KEY_LEFTCTRL":
				case "KEY_RIGHTCTRL":
				case "KEY_LEFTALT":
				case "KEY_RIGHTALT":
				case "KEY_LEFTMETA":
				case "KEY_RIGHTMETA":
				case "LED_CAPSL": // caps lock on or off
				}
				if keymapEnUS[keyEvent.KeyName] != nil {
					keyEvent.Char = keymapEnUS[keyEvent.KeyName][0]
				}
				keyChannel <- &keyEvent
				break
			}
		}
	}
	Debug("We're done")
	return nil
}

func Search() (keyboards []Keyboard, err error) {
	Debug("Searching for keyboards...")
	var files []os.DirEntry
	if files, err = os.ReadDir("/dev/input"); err != nil {
		return nil, errors.New("input directory not found")
	}
	for _, file := range files {
		Debug("File %s", file.Name())
		//if strings.Contains(file.Name(), "event") {
		var out []byte
		cmd := exec.Command("/usr/bin/udevadm", "info", "/dev/input/"+file.Name())
		if out, err = cmd.Output(); err == nil {
			scanner := bufio.NewScanner(bytes.NewReader(out))
			scanner.Split(bufio.ScanLines)
			kbd := Keyboard{
				Events: make(EventTypes),
			}
			isKeys := false
			for scanner.Scan() {
				text := scanner.Text()
				if strings.Contains(text, "ID_VENDOR=") {
					kbd.Vendor = "" + strings.Split(text, "=")[1]
					Debug("Found vendor: %s", kbd.Vendor)
				}
				if strings.Contains(text, "ID_INPUT_KEYBOARD=1") {
					Debug("%s is a keyboard", file.Name())
					kbd.File = file.Name()
					isKeys = true
				}
			}
			if isKeys {
				keyboards = append(keyboards, kbd)
			}
		}
	}

	if len(keyboards) == 0 {
		return nil, errors.New("no keyboard found")
	}

	for _, kbd := range keyboards {
		Debug("Checking keyboard %s (%s)", kbd.File, kbd.Vendor)
		cmd := exec.Command("/usr/bin/evtest", "/dev/input/"+kbd.File)
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
				//Debug("Scanning %s", scanner.Text())
				text := scanner.Text()
				if strings.Contains(text, "Input device name:") {
					kbd.Name = strings.Split(text, "\"")[1]
				} else if strings.Contains(text, "Event type") {
					eventType = &EventType{
						Codes: make(EventCodes),
					}
					ev := strings.Fields(text)
					typeNum, err = strconv.Atoi(ev[2])
					if err == nil {
						eventType.Name = "" + ev[3][1:len(ev[3])-1]
						kbd.Events[typeNum] = eventType
						//Debug("eventtype: %d,%s", typeNum, eventType.Name)
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
						//Debug("eventcode: %d,%s - %d", codeNum, eventCode.Name, eventCode.State)
					}
				} else if strings.Contains(text, "interrupt to exit") {
					cmd.Process.Kill()
					//cmd.Wait()
				}
			}
			//Debug("Found keyboard %s by %s on %s - events:\n %s", kbd.Name, kbd.Vendor, kbd.File, kbd.Events)
		} else {
			Debug("Error %v", err)
		}
	}

	Debug("Done with input")
	return keyboards, nil
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
