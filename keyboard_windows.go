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
	WM_KEYDOWN     = 0x0100
	WM_SYSKEYDOWN  = 0x0104
	VK_F6          = 0x75
	VK_F7          = 0x76
	VK_F8          = 0x77
	VK_F9          = 0x78
	VK_CONTROL     = 0x11
)

type KBDLLHOOKSTRUCT struct {
	VkCode      uint32
	ScanCode    uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

func startGlobalListener() {
	// Callback for the hook
	hookCallback := syscall.NewCallback(func(nCode int, wParam uintptr, lParam uintptr) uintptr {
		if nCode >= 0 {
			if wParam == WM_KEYDOWN || wParam == WM_SYSKEYDOWN {
				kbd := (*KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam))

				// Check Control key first
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
					}

					if key != "" {
						if globalApp != nil {
							globalApp.triggerShortcut(key)
						}
						// Consume event
						return 1
					}
				}
			}
		}
		ret, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
		return ret
	})

	// Set the hook
	hook, _, _ := procSetWindowsHookEx.Call(
		WH_KEYBOARD_LL,
		hookCallback,
		0,
		0,
	)
	// Avoid unused variable error
	_ = hook

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

	// Ideally we should unhook when done, but for this simple app, it's fine.
	// user32.NewProc("UnhookWindowsHookEx").Call(hook)
}
