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
	// Use a goroutine to decouple from the C callback thread
	// and catch any potential panics
	go func() {
		defer func() {
			if r := recover(); r != nil {
				println("Recovered from panic in onF8Pressed:", r)
			}
		}()

		if globalApp != nil {
			globalApp.triggerShortcut()
		}
	}()
}
