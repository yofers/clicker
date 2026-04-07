#include <Carbon/Carbon.h>
#include <CoreGraphics/CoreGraphics.h>

// Declare the Go function that we will call
extern void onF8Pressed();
extern void onF6Pressed();
extern void onF7Pressed();
extern void onF9Pressed();
extern void onF10Pressed();

extern int isRecording();
extern void onRecordInput(int type, int x, int y, int button, int keyCode, int scrollX, int scrollY);

#define EVT_MOVE 0
#define EVT_DOWN 1
#define EVT_UP 2
#define EVT_KEYDOWN 3
#define EVT_KEYUP 4
#define EVT_SCROLL 5

// Event callback function
CGEventRef eventCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *refcon) {
    // Shortcuts (Always Active)
    if (type == kCGEventKeyDown) {
         CGEventFlags flags = CGEventGetFlags(event);
         CGKeyCode keycode = (CGKeyCode)CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
         // Ctrl+Fx
         if (flags & kCGEventFlagMaskControl) {
            if (keycode == 100) { onF8Pressed(); return NULL; }
            else if (keycode == 97) { onF6Pressed(); return NULL; }
            else if (keycode == 98) { onF7Pressed(); return NULL; }
            else if (keycode == 101) { onF9Pressed(); return NULL; }
            else if (keycode == 109) { onF10Pressed(); return NULL; } // F10
        }
    }

    // Recording Logic
    if (isRecording()) {
        CGPoint loc = CGEventGetLocation(event);
        int x = (int)loc.x;
        int y = (int)loc.y;

        if (type == kCGEventMouseMoved || type == kCGEventLeftMouseDragged || type == kCGEventRightMouseDragged || type == kCGEventOtherMouseDragged) {
            onRecordInput(EVT_MOVE, x, y, 0, 0, 0, 0);
        } 
        else if (type == kCGEventLeftMouseDown) {
            onRecordInput(EVT_DOWN, x, y, 0, 0, 0, 0);
        } 
        else if (type == kCGEventLeftMouseUp) {
            onRecordInput(EVT_UP, x, y, 0, 0, 0, 0);
        } 
        else if (type == kCGEventRightMouseDown) {
            onRecordInput(EVT_DOWN, x, y, 1, 0, 0, 0);
        } 
        else if (type == kCGEventRightMouseUp) {
            onRecordInput(EVT_UP, x, y, 1, 0, 0, 0);
        } 
        else if (type == kCGEventOtherMouseDown) {
            int btn = (int)CGEventGetIntegerValueField(event, kCGMouseEventButtonNumber);
            onRecordInput(EVT_DOWN, x, y, btn, 0, 0, 0);
        } 
        else if (type == kCGEventOtherMouseUp) {
            int btn = (int)CGEventGetIntegerValueField(event, kCGMouseEventButtonNumber);
            onRecordInput(EVT_UP, x, y, btn, 0, 0, 0);
        } 
        else if (type == kCGEventScrollWheel) {
            int scrollY = (int)CGEventGetIntegerValueField(event, kCGScrollWheelEventDeltaAxis1);
            int scrollX = (int)CGEventGetIntegerValueField(event, kCGScrollWheelEventDeltaAxis2);
            if (scrollX != 0 || scrollY != 0) {
                onRecordInput(EVT_SCROLL, x, y, 0, 0, scrollX, scrollY);
            }
        } 
        else if (type == kCGEventKeyDown) {
            CGKeyCode keycode = (CGKeyCode)CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
            // Don't record autorepeat if we want clean actions? 
            // Usually recorders capture everything. But autorepeat might spam.
            // Let's keep autorepeat for now as some games rely on it, or filter it if needed.
            if (CGEventGetIntegerValueField(event, kCGKeyboardEventAutorepeat) == 0) {
                 onRecordInput(EVT_KEYDOWN, x, y, 0, (int)keycode, 0, 0);
            }
        } 
        else if (type == kCGEventKeyUp) {
            CGKeyCode keycode = (CGKeyCode)CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
             onRecordInput(EVT_KEYUP, x, y, 0, (int)keycode, 0, 0);
        }
    }

    return event;
}

int startKeyboardListener() {
    CGEventMask eventMask = CGEventMaskBit(kCGEventKeyDown) | 
                            CGEventMaskBit(kCGEventKeyUp) | 
                            CGEventMaskBit(kCGEventMouseMoved) |
                            CGEventMaskBit(kCGEventLeftMouseDown) |
                            CGEventMaskBit(kCGEventLeftMouseUp) |
                            CGEventMaskBit(kCGEventRightMouseDown) |
	                            CGEventMaskBit(kCGEventRightMouseUp) |
	                            CGEventMaskBit(kCGEventOtherMouseDown) |
	                            CGEventMaskBit(kCGEventOtherMouseUp) |
	                            CGEventMaskBit(kCGEventScrollWheel) |
	                            CGEventMaskBit(kCGEventLeftMouseDragged) |
                            CGEventMaskBit(kCGEventRightMouseDragged) |
                            CGEventMaskBit(kCGEventOtherMouseDragged);
                            
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
