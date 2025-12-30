package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreGraphics -framework Foundation

#include <CoreGraphics/CoreGraphics.h>
#include <ApplicationServices/ApplicationServices.h>

bool checkAccessibility() {
    NSDictionary *options = @{(__bridge id)kAXTrustedCheckOptionPrompt: @YES};
    return AXIsProcessTrustedWithOptions((__bridge CFDictionaryRef)options);
}

void cgoMouseLocation(int *x, int *y) {
    CGEventRef event = CGEventCreate(NULL);
    CGPoint point = CGEventGetLocation(event);
    *x = (int)point.x;
    *y = (int)point.y;
    CFRelease(event);
}

void cgoMouseDown(int x, int y, int button) {
    CGEventSourceRef source = CGEventSourceCreate(kCGEventSourceStateHIDSystemState);
    CGEventType down;
    CGMouseButton mouseButton;

    switch (button) {
    case 0: // Left
        down = kCGEventLeftMouseDown;
        mouseButton = kCGMouseButtonLeft;
        break;
    case 1: // Right
        down = kCGEventRightMouseDown;
        mouseButton = kCGMouseButtonRight;
        break;
    case 2: // Center
        down = kCGEventOtherMouseDown;
        mouseButton = kCGMouseButtonCenter;
        break;
    case 3: // Side 1 (Back)
        down = kCGEventOtherMouseDown;
        mouseButton = 3;
        break;
    case 4: // Side 2 (Forward)
        down = kCGEventOtherMouseDown;
        mouseButton = 4;
        break;
    default:
        CFRelease(source);
        return;
    }

    CGPoint point = CGPointMake(x, y);
    CGEventRef eventDown = CGEventCreateMouseEvent(source, down, point, mouseButton);
    CGEventPost(kCGHIDEventTap, eventDown);
    CFRelease(eventDown);
    CFRelease(source);
}

void cgoMouseUp(int x, int y, int button) {
    CGEventSourceRef source = CGEventSourceCreate(kCGEventSourceStateHIDSystemState);
    CGEventType up;
    CGMouseButton mouseButton;

    switch (button) {
    case 0: // Left
        up = kCGEventLeftMouseUp;
        mouseButton = kCGMouseButtonLeft;
        break;
    case 1: // Right
        up = kCGEventRightMouseUp;
        mouseButton = kCGMouseButtonRight;
        break;
    case 2: // Center
        up = kCGEventOtherMouseUp;
        mouseButton = kCGMouseButtonCenter;
        break;
    case 3: // Side 1 (Back)
        up = kCGEventOtherMouseUp;
        mouseButton = 3;
        break;
    case 4: // Side 2 (Forward)
        up = kCGEventOtherMouseUp;
        mouseButton = 4;
        break;
    default:
        CFRelease(source);
        return;
    }

    CGPoint point = CGPointMake(x, y);
    CGEventRef eventUp = CGEventCreateMouseEvent(source, up, point, mouseButton);
    CGEventPost(kCGHIDEventTap, eventUp);
    CFRelease(eventUp);
    CFRelease(source);
}

void cgoKeyDown(int keyCode) {
    CGEventSourceRef source = CGEventSourceCreate(kCGEventSourceStateHIDSystemState);
    CGEventRef keyDown = CGEventCreateKeyboardEvent(source, (CGKeyCode)keyCode, true);
    CGEventPost(kCGHIDEventTap, keyDown);
    CFRelease(keyDown);
    CFRelease(source);
}

void cgoKeyUp(int keyCode) {
    CGEventSourceRef source = CGEventSourceCreate(kCGEventSourceStateHIDSystemState);
    CGEventRef keyUp = CGEventCreateKeyboardEvent(source, (CGKeyCode)keyCode, false);
    CGEventPost(kCGHIDEventTap, keyUp);
    CFRelease(keyUp);
    CFRelease(source);
}

bool checkAccessibility();
void cgoMouseLocation(int *x, int *y);
void cgoMouseDown(int x, int y, int button);
void cgoMouseUp(int x, int y, int button);
void cgoKeyDown(int keyCode);
void cgoKeyUp(int keyCode);
*/
import "C"
import "time"

func CheckAccessibility() bool {
	return bool(C.checkAccessibility())
}

var macKeyMap = map[string]int{
	"A": 0, "S": 1, "D": 2, "F": 3, "H": 4, "G": 5, "Z": 6, "X": 7, "C": 8, "V": 9,
	"B": 11, "Q": 12, "W": 13, "E": 14, "R": 15, "Y": 16, "T": 17, "1": 18, "2": 19,
	"3": 20, "4": 21, "6": 22, "5": 23, "=": 24, "9": 25, "7": 26, "-": 27, "8": 28,
	"0": 29, "]": 30, "O": 31, "U": 32, "[": 33, "I": 34, "P": 35, "L": 37, "J": 38,
	"'": 39, "K": 40, ";": 41, "\\": 42, ",": 43, "/": 44, "N": 45, "M": 46, ".": 47,
	"Tab": 48, "Space": 49, "`": 50, "Delete": 51, "Enter": 36, "Esc": 53,
	"F1": 122, "F2": 120, "F3": 99, "F4": 118, "F5": 96, "F6": 97, "F7": 98, "F8": 100,
	"F9": 101, "F10": 109, "F11": 103, "F12": 111,
	"Left": 123, "Right": 124, "Down": 125, "Up": 126,
	"Backspace": 51,
}

func click(btnCode int, clickType string, duration int) {
	var x, y C.int
	C.cgoMouseLocation(&x, &y)

	switch clickType {
	case "double":
		C.cgoMouseDown(x, y, C.int(btnCode))
		C.cgoMouseUp(x, y, C.int(btnCode))
		time.Sleep(50 * time.Millisecond)
		C.cgoMouseDown(x, y, C.int(btnCode))
		C.cgoMouseUp(x, y, C.int(btnCode))
	case "long":
		C.cgoMouseDown(x, y, C.int(btnCode))
		time.Sleep(time.Duration(duration) * time.Millisecond)
		C.cgoMouseUp(x, y, C.int(btnCode))
	default: // single
		C.cgoMouseDown(x, y, C.int(btnCode))
		C.cgoMouseUp(x, y, C.int(btnCode))
	}
}

func mouseHold(btnCode int, start bool) {
	var x, y C.int
	C.cgoMouseLocation(&x, &y)
	if start {
		C.cgoMouseDown(x, y, C.int(btnCode))
	} else {
		C.cgoMouseUp(x, y, C.int(btnCode))
	}
}

func pressKey(key string, clickType string, duration int) {
	if code, ok := macKeyMap[key]; ok {
		switch clickType {
		case "double":
			C.cgoKeyDown(C.int(code))
			C.cgoKeyUp(C.int(code))
			time.Sleep(50 * time.Millisecond)
			C.cgoKeyDown(C.int(code))
			C.cgoKeyUp(C.int(code))
		case "long":
			C.cgoKeyDown(C.int(code))
			time.Sleep(time.Duration(duration) * time.Millisecond)
			C.cgoKeyUp(C.int(code))
		default: // single
			C.cgoKeyDown(C.int(code))
			C.cgoKeyUp(C.int(code))
		}
	}
}

func keyHold(key string, start bool) {
	if code, ok := macKeyMap[key]; ok {
		if start {
			C.cgoKeyDown(C.int(code))
		} else {
			C.cgoKeyUp(C.int(code))
		}
	}
}
