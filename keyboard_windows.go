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
	VK_F8          = 0x77
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
				if kbd.VkCode == VK_F8 {
					// Check if Control key is pressed
					// GetAsyncKeyState returns a short, high bit set means key is down
					ret, _, _ := procGetAsyncKeyState.Call(uintptr(VK_CONTROL))
					if ret&0x8000 != 0 {
						if globalApp != nil {
							globalApp.triggerShortcut()
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
