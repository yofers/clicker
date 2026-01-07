#include <Carbon/Carbon.h>
#include <CoreGraphics/CoreGraphics.h>

// Declare the Go function that we will call
extern void onF8Pressed();
extern void onF6Pressed();
extern void onF7Pressed();
extern void onF9Pressed();

// Event callback function
CGEventRef eventCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *refcon) {
    if (type != kCGEventKeyDown) {
        return event;
    }

    // Ignore auto-repeat events to prevent rapid toggling
    if (CGEventGetIntegerValueField(event, kCGKeyboardEventAutorepeat) != 0) {
        return event;
    }

    CGKeyCode keycode = (CGKeyCode)CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
    CGEventFlags flags = CGEventGetFlags(event);
    
    // F8 keycode is 100, F6 is 97, F7 is 98, F9 is 101
    // kCGEventFlagMaskControl = 0x00040000
    if (flags & kCGEventFlagMaskControl) {
        if (keycode == 100) { // F8
            onF8Pressed();
            return NULL;
        } else if (keycode == 97) { // F6
            onF6Pressed();
            return NULL;
        } else if (keycode == 98) { // F7
            onF7Pressed();
            return NULL;
        } else if (keycode == 101) { // F9
            onF9Pressed();
            return NULL;
        }
    }

    return event;
}

int startKeyboardListener() {
    CGEventMask eventMask = CGEventMaskBit(kCGEventKeyDown);
    CFMachPortRef eventTap = CGEventTapCreate(
        kCGSessionEventTap,
        kCGHeadInsertEventTap,
        kCGEventTapOptionDefault,
        eventMask,
        eventCallback,
        NULL
    );

    if (!eventTap) {
        // Failed to create event tap
        return -1;
    }

    CFRunLoopSourceRef runLoopSource = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, eventTap, 0);
    CFRunLoopAddSource(CFRunLoopGetCurrent(), runLoopSource, kCFRunLoopCommonModes);
    CGEventTapEnable(eventTap, true);
    CFRunLoopRun();
    return 0;
}
