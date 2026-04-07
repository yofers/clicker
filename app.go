package main

import (
	"context"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var globalApp *App

// App struct
type App struct {
	ctx              context.Context
	mu               sync.Mutex
	mouseStopChan    chan struct{}
	keyboardStopChan chan struct{}
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	globalApp = a
	// 启动全局快捷键监听 (仅当有权限时)
	startGlobalListener()
}

func (a *App) triggerShortcut(key string) {
	runtime.EventsEmit(a.ctx, "shortcut-pressed", key)
}

func normalizeKeyboardSequence(keys []string) []string {
	modifierPriority := map[string]int{
		"Ctrl":     0,
		"Shift":    1,
		"Alt":      2,
		"Command":  3,
		"Win":      3,
		"Fn":       4,
		"CapsLock": 5,
	}

	seen := make(map[string]bool)
	orderedModifiers := make([]string, 0, len(keys))
	orderedKeys := make([]string, 0, len(keys))

	for _, key := range keys {
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		if _, isModifier := modifierPriority[key]; isModifier {
			insertAt := len(orderedModifiers)
			for i, existing := range orderedModifiers {
				if modifierPriority[key] < modifierPriority[existing] {
					insertAt = i
					break
				}
			}
			orderedModifiers = append(orderedModifiers, "")
			copy(orderedModifiers[insertAt+1:], orderedModifiers[insertAt:])
			orderedModifiers[insertAt] = key
			continue
		}
		orderedKeys = append(orderedKeys, key)
	}

	return append(orderedModifiers, orderedKeys...)
}

func pressKeyboardSequence(keys []string) {
	for _, key := range normalizeKeyboardSequence(keys) {
		keyToggle(key, true)
	}
}

func releaseKeyboardSequence(keys []string) {
	ordered := normalizeKeyboardSequence(keys)
	for i := len(ordered) - 1; i >= 0; i-- {
		keyToggle(ordered[i], false)
	}
}

func clickKeyboardSequence(keys []string, clickType string, duration int) {
	doChord := func() {
		pressKeyboardSequence(keys)
		releaseKeyboardSequence(keys)
	}

	switch clickType {
	case "double":
		doChord()
		time.Sleep(50 * time.Millisecond)
		doChord()
	case "long":
		pressKeyboardSequence(keys)
		time.Sleep(time.Duration(duration) * time.Millisecond)
		releaseKeyboardSequence(keys)
	default:
		doChord()
	}
}

// StartClicking starts the auto clicker
func (a *App) StartClicking(interval int, mode string, keysOrButtons []string, clickType string, longPressDuration int) {
	a.mu.Lock()
	defer a.mu.Unlock()

	var stopChan chan struct{}

	if mode == "mouse" {
		if a.mouseStopChan != nil {
			return // Already running
		}
		a.mouseStopChan = make(chan struct{})
		stopChan = a.mouseStopChan
	} else {
		if a.keyboardStopChan != nil {
			return // Already running
		}
		a.keyboardStopChan = make(chan struct{})
		stopChan = a.keyboardStopChan
	}

	// Parse Mouse Buttons if mode is mouse
	var btnCodes []int
	if mode == "mouse" {
		for _, k := range keysOrButtons {
			var code int
			switch k {
			case "right":
				code = 1
			case "center":
				code = 2
			case "side1":
				code = 3
			case "side2":
				code = 4
			default:
				code = 0 // left
			}
			btnCodes = append(btnCodes, code)
		}
	}

	go func() {
		if clickType == "hold" {

			// Hold mode: Press Down -> Wait Stop -> Press Up
			if mode == "mouse" {
				for _, code := range btnCodes {
					mouseHold(code, true)
				}
			} else {
				pressKeyboardSequence(keysOrButtons)
			}

			<-stopChan

			if mode == "mouse" {
				for _, code := range btnCodes {
					mouseHold(code, false)
				}
			} else {
				releaseKeyboardSequence(keysOrButtons)
			}
			// Cleanup channel reference after goroutine finishes?
			// Ideally StopClicking handles it, but if we need cleanup here:
			// a.mu.Lock()
			// if mode == "mouse" { a.mouseStopChan = nil } else { a.keyboardStopChan = nil }
			// a.mu.Unlock()
			// But since we close the channel in StopClicking, we should nil it there.
			return
		}

		for {
			// Check stop before action
			select {
			case <-stopChan:
				return
			default:
			}

			// Perform Action
			if mode == "mouse" {
				for _, code := range btnCodes {
					click(code, clickType, longPressDuration)
				}
			} else if mode == "keyboard" {
				clickKeyboardSequence(keysOrButtons, clickType, longPressDuration)
			}

			// Wait Interval
			select {
			case <-stopChan:
				return
			case <-time.After(time.Duration(interval) * time.Millisecond):
			}
		}
	}()
}

// StopClicking stops the auto clicker
func (a *App) StopClicking(mode string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if mode == "mouse" {
		if a.mouseStopChan != nil {
			close(a.mouseStopChan)
			a.mouseStopChan = nil
		}
	} else if mode == "keyboard" {
		if a.keyboardStopChan != nil {
			close(a.keyboardStopChan)
			a.keyboardStopChan = nil
		}
	} else if mode == "all" {
		if a.mouseStopChan != nil {
			close(a.mouseStopChan)
			a.mouseStopChan = nil
		}
		if a.keyboardStopChan != nil {
			close(a.keyboardStopChan)
			a.keyboardStopChan = nil
		}
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return "Hello " + name
}

func (a *App) CheckPermission() bool {
	allowed := CheckAccessibility()
	if allowed {
		startGlobalListener()
	}
	return allowed
}
