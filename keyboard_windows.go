package main

import (
	"syscall"
	"unsafe"
)

var (
	moduser32            = syscall.NewLazyDLL("user32.dll")
	procSetWindowsHookEx = moduser32.NewProc("SetWindowsHookExW")
	procCallNextHookEx   = moduser32.NewProc("CallNextHookEx")
	procGetAsyncKeyState = moduser32.NewProc("GetAsyncKeyState")
	procGetMessage       = moduser32.NewProc("GetMessageW")
)

const (
	WH_KEYBOARD_LL = 13
	WH_MOUSE_LL    = 14
	WM_KEYDOWN     = 0x0100
	WM_KEYUP       = 0x0101
	WM_SYSKEYDOWN  = 0x0104
	WM_SYSKEYUP    = 0x0105

	WM_MOUSEMOVE   = 0x0200
	WM_LBUTTONDOWN = 0x0201
	WM_LBUTTONUP   = 0x0202
	WM_RBUTTONDOWN = 0x0204
	WM_RBUTTONUP   = 0x0205
	WM_MBUTTONDOWN = 0x0207
	WM_MBUTTONUP   = 0x0208

	VK_F6          = 0x75
	VK_F7          = 0x76
	VK_F8          = 0x77
	VK_F9          = 0x78
	VK_F10         = 0x79
	VK_CONTROL     = 0x11
)

type KBDLLHOOKSTRUCT struct {
	VkCode      uint32
	ScanCode    uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

type MSLLHOOKSTRUCT struct {
	Pt          struct{ X, Y int32 }
	MouseData   uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

// Windows key mapping to our string format
var winKeyCodeMap = map[uint32]string{
	0x41: "A", 0x42: "B", 0x43: "C", 0x44: "D", 0x45: "E",
	0x46: "F", 0x47: "G", 0x48: "H", 0x49: "I", 0x4A: "J",
	0x4B: "K", 0x4C: "L", 0x4D: "M", 0x4E: "N", 0x4F: "O",
	0x50: "P", 0x51: "Q", 0x52: "R", 0x53: "S", 0x54: "T",
	0x55: "U", 0x56: "V", 0x57: "W", 0x58: "X", 0x59: "Y", 0x5A: "Z",
	0x30: "0", 0x31: "1", 0x32: "2", 0x33: "3", 0x34: "4",
	0x35: "5", 0x36: "6", 0x37: "7", 0x38: "8", 0x39: "9",
	0x20: "Space", 0x0D: "Enter", 0x1B: "Esc", 0x08: "Backspace", 0x09: "Tab",
	0x25: "Left", 0x26: "Up", 0x27: "Right", 0x28: "Down",
	0x70: "F1", 0x71: "F2", 0x72: "F3", 0x73: "F4", 0x74: "F5",
	0x75: "F6", 0x76: "F7", 0x77: "F8", 0x78: "F9", 0x79: "F10",
	0x7A: "F11", 0x7B: "F12",
}

func startGlobalListener() {
	// KEYBOARD HOOK
	keyboardCallback := syscall.NewCallback(func(nCode int, wParam uintptr, lParam uintptr) uintptr {
		if nCode >= 0 {
			kbd := (*KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam))
			
			// 1. Handle Shortcuts (Ctrl+Fx) - Only on KeyDown
			if wParam == WM_KEYDOWN || wParam == WM_SYSKEYDOWN {
				ret, _, _ := procGetAsyncKeyState.Call(uintptr(VK_CONTROL))
				isCtrlDown := ret&0x8000 != 0

				if isCtrlDown {
					var key string
					if kbd.VkCode == VK_F8 {
						key = "f8"
					} else if kbd.VkCode == VK_F7 {
						key = "f7"
					} else if kbd.VkCode == VK_F6 {
						key = "f6"
					} else if kbd.VkCode == VK_F9 {
						key = "f9"
					} else if kbd.VkCode == VK_F10 {
						key = "f10"
					}

					if key != "" {
						if globalApp != nil {
							globalApp.triggerShortcut(key)
						}
						// Consume shortcut event
						return 1
					}
				}
			}

			// 2. Handle Recording
			globalRecorder.mu.Lock()
			isRecording := globalRecorder.isRecording
			globalRecorder.mu.Unlock()

			if isRecording {
				var actionType ActionType
				if wParam == WM_KEYDOWN || wParam == WM_SYSKEYDOWN {
					actionType = ActionKeyDown
				} else if wParam == WM_KEYUP || wParam == WM_SYSKEYUP {
					actionType = ActionKeyUp
				}
				
				if actionType != "" {
					keyName, ok := winKeyCodeMap[kbd.VkCode]
					if ok {
						RecordEvent(Action{
							Type: actionType,
							Key:  keyName,
						})
					}
				}
			}
		}
		ret, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
		return ret
	})

	keyboardHook, _, _ := procSetWindowsHookEx.Call(
		WH_KEYBOARD_LL,
		keyboardCallback,
		0,
		0,
	)

	// MOUSE HOOK
	mouseCallback := syscall.NewCallback(func(nCode int, wParam uintptr, lParam uintptr) uintptr {
		if nCode >= 0 {
			globalRecorder.mu.Lock()
			isRecording := globalRecorder.isRecording
			globalRecorder.mu.Unlock()

			if isRecording {
				ms := (*MSLLHOOKSTRUCT)(unsafe.Pointer(lParam))
				x, y := int(ms.Pt.X), int(ms.Pt.Y)
				
				var action Action
				action.X = x
				action.Y = y
				
				recorded := true
				
				switch wParam {
				case WM_MOUSEMOVE:
					action.Type = ActionMouseMove
				case WM_LBUTTONDOWN:
					action.Type = ActionMouseDown
					action.Button = "left"
				case WM_LBUTTONUP:
					action.Type = ActionMouseUp
					action.Button = "left"
				case WM_RBUTTONDOWN:
					action.Type = ActionMouseDown
					action.Button = "right"
				case WM_RBUTTONUP:
					action.Type = ActionMouseUp
					action.Button = "right"
				case WM_MBUTTONDOWN:
					action.Type = ActionMouseDown
					action.Button = "center"
				case WM_MBUTTONUP:
					action.Type = ActionMouseUp
					action.Button = "center"
				default:
					recorded = false
				}
				
				if recorded {
					RecordEvent(action)
				}
			}
		}
		ret, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
		return ret
	})

	mouseHook, _, _ := procSetWindowsHookEx.Call(
		WH_MOUSE_LL,
		mouseCallback,
		0,
		0,
	)

	// Avoid unused variable error
	_ = keyboardHook
	_ = mouseHook

	// Message loop
	var msg struct {
		hwnd    syscall.Handle
		message uint32
		wParam  uintptr
		lParam  uintptr
		time    uint32
		pt      struct{ x, y int32 }
	}

	for {
		ret, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if ret == 0 {
			break
		}
	}
}
