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
		// Global delay for both Clicker and Presser modes
		// This prevents shortcut interference (e.g. Ctrl+F8 being interpreted as Ctrl+Click)
		// and gives user time to release keys.
		time.Sleep(1000 * time.Millisecond)

		// Check if stopped during sleep
		select {
		case <-stopChan:
			return
		default:
		}

		if clickType == "hold" {

			// Hold mode: Press Down -> Wait Stop -> Press Up
			if mode == "mouse" {
				for _, code := range btnCodes {
					mouseHold(code, true)
				}
			} else {
				for _, k := range keysOrButtons {
					keyHold(k, true)
				}
			}

			<-stopChan

			if mode == "mouse" {
				for _, code := range btnCodes {
					mouseHold(code, false)
				}
			} else {
				for _, k := range keysOrButtons {
					keyHold(k, false)
				}
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
				for _, k := range keysOrButtons {
					pressKey(k, clickType, longPressDuration)
				}
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
	return CheckAccessibility()
}
