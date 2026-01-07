package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreGraphics -framework Foundation -framework Carbon

extern int startKeyboardListener();
*/
import "C"
import (
	"runtime"
)

func startGlobalListener() {
	// Only start listener if permission is granted, otherwise we risk blocking the app or causing issues
	if CheckAccessibility() {
		go func() {
			runtime.LockOSThread()
			C.startKeyboardListener()
		}()
	}
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
