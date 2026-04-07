package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreGraphics -framework Foundation -framework Carbon

extern int startKeyboardListener();
*/
import "C"
import (
	"runtime"
	"sync"
)

var (
	darwinListenerMu      sync.Mutex
	darwinListenerStarted bool
)

func startGlobalListener() {
	// Only start listener if permission is granted, otherwise we risk blocking the app or causing issues
	if !CheckAccessibility() {
		return
	}

	darwinListenerMu.Lock()
	if darwinListenerStarted {
		darwinListenerMu.Unlock()
		return
	}
	darwinListenerStarted = true
	darwinListenerMu.Unlock()

	go func() {
		runtime.LockOSThread()
		C.startKeyboardListener()
	}()
}

//export onF8Pressed
func onF8Pressed() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				println("Recovered from panic in onF8Pressed:", r)
			}
		}()
		if globalApp != nil {
			globalApp.triggerShortcut("f8")
		}
	}()
}

//export onF6Pressed
func onF6Pressed() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				println("Recovered from panic in onF6Pressed:", r)
			}
		}()
		if globalApp != nil {
			globalApp.triggerShortcut("f6")
		}
	}()
}

//export onF7Pressed
func onF7Pressed() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				println("Recovered from panic in onF7Pressed:", r)
			}
		}()
		if globalApp != nil {
			globalApp.triggerShortcut("f7")
		}
	}()
}

//export onF9Pressed
func onF9Pressed() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				println("Recovered from panic in onF9Pressed:", r)
			}
		}()
		if globalApp != nil {
			globalApp.triggerShortcut("f9")
		}
	}()
}

//export onF10Pressed
func onF10Pressed() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				println("Recovered from panic in onF10Pressed:", r)
			}
		}()
		if globalApp != nil {
			globalApp.triggerShortcut("f10")
		}
	}()
}

//export isRecording
func isRecording() int {
	globalRecorder.mu.Lock()
	defer globalRecorder.mu.Unlock()
	if globalRecorder.isRecording {
		return 1
	}
	return 0
}

// Reverse mapping for key codes
var macKeyCodeMap map[int]string

func init() {
	macKeyCodeMap = make(map[int]string)
	for k, v := range macKeyMap {
		macKeyCodeMap[v] = k
	}
}

//export onRecordInput
func onRecordInput(eventType int, x, y int, button int, keyCode int, scrollX int, scrollY int) {
	// eventType: 0=Move, 1=Down, 2=Up, 3=KeyDown, 4=KeyUp, 5=Scroll

	// Fast check again (though C side should have checked)
	// We do this asynchronously to not block the C callback too much?
	// Actually C callback blocks the event tap. We should be fast.
	// But RecordEvent takes a lock.

	action := Action{
		X: x,
		Y: y,
	}

	switch eventType {
	case 0: // Move
		action.Type = ActionMouseMove
	case 1: // Down
		action.Type = ActionMouseDown
		switch button {
		case 0:
			action.Button = "left"
		case 1:
			action.Button = "right"
		case 2:
			action.Button = "center"
		case 3:
			action.Button = "side1"
		case 4:
			action.Button = "side2"
		default:
			action.Button = "unknown"
		}
	case 2: // Up
		action.Type = ActionMouseUp
		switch button {
		case 0:
			action.Button = "left"
		case 1:
			action.Button = "right"
		case 2:
			action.Button = "center"
		case 3:
			action.Button = "side1"
		case 4:
			action.Button = "side2"
		default:
			action.Button = "unknown"
		}
	case 3: // KeyDown
		action.Type = ActionKeyDown
		if name, ok := macKeyCodeMap[keyCode]; ok {
			action.Key = name
		} else {
			// Unknown key, maybe ignore or record code?
			return
		}
	case 4: // KeyUp
		action.Type = ActionKeyUp
		if name, ok := macKeyCodeMap[keyCode]; ok {
			action.Key = name
		} else {
			return
		}
	case 5: // Scroll
		action.Type = ActionScroll
		action.ScrollX = scrollX
		action.ScrollY = scrollY
	}

	RecordEvent(action)
}
