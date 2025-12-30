package main

import (
	"syscall"
	"time"
	"unsafe"
)

var (
	user32           = syscall.NewLazyDLL("user32.dll")
	procSetCursorPos = user32.NewProc("SetCursorPos")
	procGetCursorPos = user32.NewProc("GetCursorPos")
	procMouseEvent   = user32.NewProc("mouse_event")
	procKeybdEvent   = user32.NewProc("keybd_event")
	procVkKeyScan    = user32.NewProc("VkKeyScanW")
)

const (
	MOUSEEVENTF_LEFTDOWN   = 0x0002
	MOUSEEVENTF_LEFTUP     = 0x0004
	MOUSEEVENTF_RIGHTDOWN  = 0x0008
	MOUSEEVENTF_RIGHTUP    = 0x0010
	MOUSEEVENTF_MIDDLEDOWN = 0x0020
	MOUSEEVENTF_MIDDLEUP   = 0x0040
	MOUSEEVENTF_XDOWN      = 0x0080
	MOUSEEVENTF_XUP        = 0x0100

	XBUTTON1 = 0x0001
	XBUTTON2 = 0x0002

	KEYEVENTF_KEYUP = 0x0002
)

type POINT struct {
	X, Y int32
}

var winKeyMap = map[string]byte{
	"Space": 0x20, "Enter": 0x0D, "Tab": 0x09, "Esc": 0x1B, "Backspace": 0x08, "Delete": 0x2E,
	"Up": 0x26, "Down": 0x28, "Left": 0x25, "Right": 0x27,
	"F1": 0x70, "F2": 0x71, "F3": 0x72, "F4": 0x73, "F5": 0x74, "F6": 0x75,
	"F7": 0x76, "F8": 0x77, "F9": 0x78, "F10": 0x79, "F11": 0x7A, "F12": 0x7B,
	"0": 0x30, "1": 0x31, "2": 0x32, "3": 0x33, "4": 0x34,
	"5": 0x35, "6": 0x36, "7": 0x37, "8": 0x38, "9": 0x39,
	"A": 0x41, "B": 0x42, "C": 0x43, "D": 0x44, "E": 0x45, "F": 0x46, "G": 0x47,
	"H": 0x48, "I": 0x49, "J": 0x4A, "K": 0x4B, "L": 0x4C, "M": 0x4D, "N": 0x4E,
	"O": 0x4F, "P": 0x50, "Q": 0x51, "R": 0x52, "S": 0x53, "T": 0x54, "U": 0x55,
	"V": 0x56, "W": 0x57, "X": 0x58, "Y": 0x59, "Z": 0x5A,
}

func CheckAccessibility() bool {
	return true
}

func click(btnCode int, clickType string, duration int) {
	var pt POINT
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))

	var down, up, data uint32
	switch btnCode {
	case 0: // Left
		down = MOUSEEVENTF_LEFTDOWN
		up = MOUSEEVENTF_LEFTUP
	case 1: // Right
		down = MOUSEEVENTF_RIGHTDOWN
		up = MOUSEEVENTF_RIGHTUP
	case 2: // Center
		down = MOUSEEVENTF_MIDDLEDOWN
		up = MOUSEEVENTF_MIDDLEUP
	case 3: // Side 1
		down = MOUSEEVENTF_XDOWN
		up = MOUSEEVENTF_XUP
		data = XBUTTON1
	case 4: // Side 2
		down = MOUSEEVENTF_XDOWN
		up = MOUSEEVENTF_XUP
		data = XBUTTON2
	}

	doClick := func() {
		procMouseEvent.Call(uintptr(down), 0, 0, uintptr(data), 0)
		procMouseEvent.Call(uintptr(up), 0, 0, uintptr(data), 0)
	}

	switch clickType {
	case "double":
		doClick()
		time.Sleep(50 * time.Millisecond)
		doClick()
	case "long":
		procMouseEvent.Call(uintptr(down), 0, 0, uintptr(data), 0)
		time.Sleep(time.Duration(duration) * time.Millisecond)
		procMouseEvent.Call(uintptr(up), 0, 0, uintptr(data), 0)
	default: // single
		doClick()
	}
}

func mouseHold(btnCode int, start bool) {
	var pt POINT
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))

	var down, up, data uint32
	switch btnCode {
	case 0: // Left
		down = MOUSEEVENTF_LEFTDOWN
		up = MOUSEEVENTF_LEFTUP
	case 1: // Right
		down = MOUSEEVENTF_RIGHTDOWN
		up = MOUSEEVENTF_RIGHTUP
	case 2: // Center
		down = MOUSEEVENTF_MIDDLEDOWN
		up = MOUSEEVENTF_MIDDLEUP
	case 3: // Side 1
		down = MOUSEEVENTF_XDOWN
		up = MOUSEEVENTF_XUP
		data = XBUTTON1
	case 4: // Side 2
		down = MOUSEEVENTF_XDOWN
		up = MOUSEEVENTF_XUP
		data = XBUTTON2
	}

	if start {
		procMouseEvent.Call(uintptr(down), 0, 0, uintptr(data), 0)
	} else {
		procMouseEvent.Call(uintptr(up), 0, 0, uintptr(data), 0)
	}
}

func pressKey(key string, clickType string, duration int) {
	var vk byte
	if val, ok := winKeyMap[key]; ok {
		vk = val
	} else {
		// Fallback for simple chars if not in map
		if len(key) == 1 {
			ret, _, _ := procVkKeyScan.Call(uintptr(key[0]))
			vk = byte(ret)
		}
	}

	if vk != 0 {
		switch clickType {
		case "double":
			procKeybdEvent.Call(uintptr(vk), 0, 0, 0)
			procKeybdEvent.Call(uintptr(vk), 0, KEYEVENTF_KEYUP, 0)
			time.Sleep(50 * time.Millisecond)
			procKeybdEvent.Call(uintptr(vk), 0, 0, 0)
			procKeybdEvent.Call(uintptr(vk), 0, KEYEVENTF_KEYUP, 0)
		case "long":
			procKeybdEvent.Call(uintptr(vk), 0, 0, 0)
			time.Sleep(time.Duration(duration) * time.Millisecond)
			procKeybdEvent.Call(uintptr(vk), 0, KEYEVENTF_KEYUP, 0)
		default: // single
			procKeybdEvent.Call(uintptr(vk), 0, 0, 0)
			procKeybdEvent.Call(uintptr(vk), 0, KEYEVENTF_KEYUP, 0)
		}
	}
}

func keyHold(key string, start bool) {
	var vk byte
	if val, ok := winKeyMap[key]; ok {
		vk = val
	} else {
		if len(key) == 1 {
			ret, _, _ := procVkKeyScan.Call(uintptr(key[0]))
			vk = byte(ret)
		}
	}

	if vk != 0 {
		if start {
			procKeybdEvent.Call(uintptr(vk), 0, 0, 0)
		} else {
			procKeybdEvent.Call(uintptr(vk), 0, KEYEVENTF_KEYUP, 0)
		}
	}
}
