#include <Carbon/Carbon.h>
#include <CoreGraphics/CoreGraphics.h>

// Declare the Go function that we will call
extern void onF8Pressed();

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
    
    // F8 keycode is 100, Check if Control key is pressed
    // kCGEventFlagMaskControl = 0x00040000
    if (keycode == 100 && (flags & kCGEventFlagMaskControl)) {
        onF8Pressed();
        // Consume the event so it doesn't propagate
        return NULL;
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
